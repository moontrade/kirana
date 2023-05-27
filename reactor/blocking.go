package reactor

import (
	"context"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/mpmc"
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/runtimex"
	"github.com/moontrade/kirana/pkg/timex"
	"github.com/moontrade/kirana/pkg/util"
	logger "github.com/moontrade/log"
	"runtime"
	"sync"
	"time"
)

func InvokeBlocking(task func()) bool {
	return blocking.Invoke(task)
}

// BlockingPool executes tasks that may block, but *should execute rather quickly <1s.
// These tasks are forbidden to sleep. Use a worker for those types of tasks. This pool
// has a fixed number of worker goroutines and tasks are spread among them in round-robin
// fashion.
type BlockingPool struct {
	started     int64
	workers     []*blockingWorker
	workersMask int64
	idleCount   counter.Counter
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	jobs        counter.Counter
	jobsDur     counter.TimeCounter
	jobsDurMin  counter.Counter
	jobsDurMax  counter.Counter
	profile     bool
}

func NewBlockingPool(numWorkers, queueSize int) *BlockingPool {
	if numWorkers < 1 {
		numWorkers = runtime.GOMAXPROCS(0)
		if numWorkers > 1 {
			numWorkers /= 2
		}
	}
	queueSize = pmath.CeilToPowerOf2(queueSize)
	numWorkers = pmath.CeilToPowerOf2(numWorkers)
	workers := make([]*blockingWorker, numWorkers)
	bp := &BlockingPool{
		started: timex.NanoTime(),
		workers: workers,
		profile: false,
	}
	bp.ctx, bp.cancel = context.WithCancel(context.Background())
	bp.wg.Add(len(workers))
	for i := 0; i < len(workers); i++ {
		worker := &blockingWorker{
			pool:    bp,
			started: bp.started,
			queue:   mpmc.NewBoundedWake[func()](int64(queueSize), nil),
			wg:      sync.WaitGroup{},
		}
		worker.ctx, worker.cancel = context.WithCancel(context.Background())
		worker.wg.Add(1)
		workers[i] = worker
		go worker.run()
	}
	bp.wg.Add(1)
	return bp
}

func (b *BlockingPool) Close() error {
	b.cancel()
	for _, worker := range b.workers {
		_ = worker.Close()
	}
	return nil
}

func (b *BlockingPool) Invoke(fn func()) bool {
	worker := b.workers[b.jobs.Incr()&b.workersMask]
	return worker.queue.EnqueueUnsafe(runtimex.FuncToPointer(fn))
}

type blockingWorker struct {
	pool       *BlockingPool
	started    int64
	queue      *mpmc.BoundedWake[func()]
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	jobs       counter.Counter
	jobsDur    counter.TimeCounter
	jobsDurMin counter.Counter
	jobsDurMax counter.Counter
}

func (w *blockingWorker) Close() error {
	w.cancel()
	return nil
}

func (w *blockingWorker) run() {
	defer func() {
		w.wg.Done()
		w.pool.wg.Done()
	}()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var (
		queue     = w.queue
		queueWake = queue.Wake()
		done      = w.ctx.Done()
		profile   = w.pool.profile
		begin     = timex.NanoTime()
		elapsed   = begin
		timer     = time.NewTimer(time.Microsecond * 50)
	)
	_ = timer
	onTask := func(task func()) {
		w.jobs.Incr()
		if profile {
			begin = timex.NanoTime()
		}
		w.invoke(task)
		if profile {
			elapsed = timex.NanoTime() - begin
			w.jobsDur.Add(elapsed)
			if w.jobsDurMin == 0 {
				w.jobsDurMin.Store(elapsed)
			}
			if w.jobsDurMax.Load() < elapsed {
				w.jobsDurMax.Store(elapsed)
			}
		}
	}
Loop:
	for {
		fn := queue.DequeueDeref()
		//n := queue.DequeueManyDeref(math.MaxUint32, onTask)

		if fn == nil {
			w.pool.idleCount.Incr()
			//timer.Reset(Time.Hour)
			select {
			case <-queueWake:
				w.pool.idleCount.Decr()
				continue
			//case <-timer.C:
			//	continue
			case <-done:
				break Loop
			}
		} else {
			onTask(fn)
			continue
		}
	}
}

func (w *blockingWorker) invoke(task func()) {
	defer func() {
		e := recover()
		if e != nil {
			err := util.PanicToError(e)
			logger.WarnErr(err, "panic")
		}
	}()
	task()
}
