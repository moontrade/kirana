package reactor

import (
	"context"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/fastrand"
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

func EnqueueBlocking(task func()) bool {
	return blocking.Enqueue(task)
}

// BlockingPool executes tasks that may block, but *should execute rather quickly <1s.
// These tasks are forbidden to sleep. Use a worker for those types of tasks. This pool
// has a fixed number of worker goroutines and tasks are spread among them in round-robin
// fashion.
type BlockingPool struct {
	started     int64
	workers     []*blockingWorker
	workersMask int32
	idleCount   counter.Counter
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	jobs        counter.Counter
	done        counter.Counter
	err         counter.Counter
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
		started:     timex.NanoTime(),
		workers:     workers,
		profile:     false,
		workersMask: int32(numWorkers - 1),
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

func (b *BlockingPool) Checkpoint() {
	for b.jobs.Load() > b.done.Load() {
		time.Sleep(time.Millisecond * 100)
	}
}

func (b *BlockingPool) Close() error {
	b.cancel()
	for _, worker := range b.workers {
		_ = worker.Close()
	}
	return nil
}

func (b *BlockingPool) Enqueue(fn func()) bool {
	var (
		idx    = fastrand.Uint32n(uint32(len(b.workers)))
		worker = b.workers[idx]
		count  = 0
		fnp    = runtimex.FuncToPointer(fn)
	)

	for !worker.queue.EnqueueUnsafe(fnp) {
		if count%len(b.workers) == 0 {
			runtime.Gosched()
		}
		idx++
		count++
		if idx >= uint32(len(b.workers)) {
			idx = 0
		}
		worker = b.workers[idx]
	}

	//worker := b.workers[gid.ProcessorID()&b.workersMask]
	//if !worker.queue.EnqueueUnsafeTimeout(runtimex.FuncToPointer(fn), time.Second*10) {
	//	//if !worker.queue.EnqueueUnsafe(runtimex.FuncToPointer(fn)) {
	//	b.done.Incr()
	//	b.err.Incr()
	//	return false
	//}
	return true
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
	var (
		queue     = w.queue
		queueWake = queue.Wake()
		done      = w.ctx.Done()
		begin     = timex.NanoTime()
		elapsed   = begin
	)
	onTask := func(task func()) {
		w.jobs.Incr()
		defer w.pool.done.Incr()
		if w.pool.profile {
			begin = timex.NanoTime()
		}
		w.invoke(task)
		if w.pool.profile {
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
			//if n == 0 {
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
