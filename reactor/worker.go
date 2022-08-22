package reactor

import (
	"context"
	logger "github.com/moontrade/log"
	"github.com/moontrade/wormhole/pkg/counter"
	"github.com/moontrade/wormhole/pkg/pmath"
	"github.com/moontrade/wormhole/pkg/timex"
	"github.com/moontrade/wormhole/pkg/util"
	"sync"
	"sync/atomic"
	"time"
)

type WorkerPool struct {
	workCh     chan *func()
	slots      []*worker
	activeSize int
	size       int
	minWorkers int32
	reserved   counter.Counter
	workers    counter.Counter
	workersDur counter.TimeCounter
	idle       counter.Counter
	started    int64
	maxIdle    time.Duration
	evictTimer time.Duration
	jobs       counter.Counter
	jobsDur    counter.TimeCounter
	spawned    counter.Counter
	evicted    counter.Counter
	now        int64
	stop       bool
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewWorkerPool(
	minWorkers, maxWorkers int,
	spawnQueueSize, doneQueueSize int,
	maxIdle, evictTimer time.Duration,
) *WorkerPool {
	if maxWorkers < 4 {
		maxWorkers = 4
	}
	if minWorkers > 0 {
		minWorkers = pmath.CeilToPowerOf2(minWorkers)
	}
	maxWorkers = pmath.CeilToPowerOf2(maxWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	w := &WorkerPool{
		minWorkers: int32(minWorkers),
		maxIdle:    maxIdle,
		evictTimer: evictTimer,
		workCh:     make(chan *func(), pmath.CeilToPowerOf2(spawnQueueSize)),
		slots:      make([]*worker, maxWorkers, maxWorkers),
		ctx:        ctx,
		cancel:     cancel,
	}
	w.wg.Add(1)
	go w.run()
	return w
}

func (wp *WorkerPool) Close() error {
	wp.cancel()
	//wp.wg.Wait()
	return nil
}

func (wp *WorkerPool) Go(fn func()) error {
	return wp.GoRef(&fn)
}

func (wp *WorkerPool) GoRef(fn *func()) error {
	if fn == nil {
		return nil
	}
	select {
	case wp.workCh <- fn:
	}
	return nil
}

func (wp *WorkerPool) run() {
	defer wp.wg.Done()
	var (
		closeNotify = wp.ctx.Done()
		timer       = time.NewTimer(wp.evictTimer)
	)
	for {
		select {
		case <-timer.C:
			wp.now = timex.NanoTime()
			timer.Reset(wp.evictTimer)
			wp.evict()
		case <-closeNotify:
			wp.close()
			return
		}
	}
}

func (wp *WorkerPool) close() {
	wp.stop = true
	for i := 0; i < len(wp.slots); i++ {
		w := wp.slots[i]
		if w == nil {
			break
		}
		select {
		case w.ch <- nil:
		default:
			logger.Warn("could not send kill pill")
		}
	}
}

func (wp *WorkerPool) onCompleted(w *worker) bool {
	w.idle = wp.now
	wp.release(w)
	//wp.reserved.Decr()
	return true
}

func (wp *WorkerPool) evict() {
	if wp.minWorkers <= 0 {
		return
	}
	var (
		now        = wp.now
		minWorkers = int(atomic.LoadInt32(&wp.minWorkers))
	)
Loop:
	for idx := wp.activeSize; idx < wp.size && wp.size > minWorkers; {
		w := wp.slots[idx]
		if w == nil {
			break Loop
		}

		// Timeout?
		if time.Duration(now-w.idle) < wp.maxIdle {
			idx++
			continue Loop
		}

		//remaining--
		select {
		case w.ch <- nil:
		default:
		}

		wp.size--
		if idx < wp.size {
			// move last to fill in the gap
			last := wp.slots[wp.size]
			last.idx = idx
			wp.slots[idx] = last
			// nil out previous last slots
			wp.slots[wp.size] = nil
		} else {
			wp.slots[idx] = nil
		}
	}
}

//func (wp *WorkerPool) evict() {
//	remaining := wp.minWorkers - wp.activeSize
//	if remaining <= 0 {
//		return
//	}
//	wp.now = timex.NanoTime()
//	now := wp.now
//	maxIdle := wp.maxIdle
//	i := wp.activeSize
//
//Reactor:
//	for i < wp.size {
//		w := wp.slots[i]
//		if w == nil {
//			break Reactor
//		}
//		if Time.Duration(now-w.idle) > maxIdle {
//			remaining--
//			if remaining <= 0 {
//				break Reactor
//			}
//			wp.evicted.Incr()
//			select {
//			case w.ch <- nil:
//			default:
//			}
//
//			if i+1 == len(wp.slots) {
//				break Reactor
//			}
//			next := wp.slots[i+1]
//
//			if idx < wp.activeSize {
//				last := wp.get(wp.activeSize)
//				last.idx = idx
//				wp.set(idx, last)
//				w.idx = wp.activeSize
//				wp.set(wp.activeSize, w)
//			}
//		} else {
//			i++
//		}
//	}
//}

const (
	workerIdle   int32 = 0
	workerBusy   int32 = 1
	workerClosed int32 = 2
)

type worker struct {
	pool    *WorkerPool
	idx     int
	ch      chan func()
	started int64
	idle    int64
	jobs    counter.Counter
	jobsDur counter.TimeCounter
	panics  counter.Counter
	state   int32
	wg      sync.WaitGroup
}

func newWorker(pool *WorkerPool, idx int) *worker {
	w := &worker{
		started: pool.now,
		pool:    pool,
		idx:     idx,
		ch:      make(chan func(), 4096),
	}
	w.pool.workers.Incr()
	w.wg.Add(1)
	go w.run()
	return w
}

func (w *worker) run() {
	defer func() {
		elapsed := w.pool.now - w.pool.started
		w.pool.workersDur.Add(elapsed)
		w.pool.workers.Decr()
		atomic.StoreInt32(&w.state, workerClosed)
		w.wg.Done()
		//close(w.ch)
		//logger.Debug("worker stopped")
		if e := recover(); e != nil {
			logger.Error(util.PanicToError(e))
		}
	}()

	var (
		//job        *func()
		//workQ      = &w.pool.workQ
		//workNotify = workQ.
		work = w.pool.workCh
	)

	for {
		select {
		case job, ok := <-work:
			if !ok {
				return
			}
			w.jobs.Incr()
			w.pool.jobs.Incr()
			begin := timex.NanoTime()
			w.invoke(*job)
			elapsed := timex.NanoTime() - begin
			w.jobsDur.Add(elapsed)
			w.pool.jobsDur.Add(elapsed)
			//case _, ok := <-workNotify:
			//	if !ok {
			//		return
			//	}
		}
	}
}

func (w *worker) invoke(fn func()) {
	defer func() {
		if e := recover(); e != nil {
			w.panics.Incr()
			logger.Warn(util.PanicToError(e))
		}
	}()
	fn()
}

func (wp *WorkerPool) alloc() *worker {
	//if wp.size+1 == len(wp.slots) {
	//	return nil
	//}
	if wp.activeSize+1 >= len(wp.slots) {
		return nil
	}
	idx := wp.activeSize
	wp.activeSize++
	worker := wp.slots[idx]
	if worker == nil {
		wp.size++
		worker = newWorker(wp, idx)
		wp.slots[idx] = worker
	}
	return worker
}

func (wp *WorkerPool) release(w *worker) {
	idx := w.idx
	if wp.activeSize == 0 || idx >= wp.activeSize {
		panic("out of active size bounds")
		return
	}
	wp.activeSize--
	if idx < wp.activeSize {
		last := wp.slots[wp.activeSize]
		last.idx = idx
		wp.slots[idx] = last
		w.idx = wp.activeSize
		wp.slots[wp.activeSize] = w
	}
}
