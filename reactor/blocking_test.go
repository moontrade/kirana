package reactor

import (
	"fmt"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/runtimex"
	"github.com/moontrade/kirana/pkg/timex"
	logger "github.com/moontrade/log"
	"github.com/panjf2000/ants/v2"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func GetFunctionName(i interface{}) string {
	p := reflect.ValueOf(i).Pointer()
	return runtime.FuncForPC(p).Name()
}

// emptyInterface is the header for an interface{} value.
type emptyInterface struct {
	typ  unsafe.Pointer
	word unsafe.Pointer
}

func TestBlockingPool(t *testing.T) {
	fn := func() {
		pc, file, line, ok := runtime.Caller(1)
		_, _, _, _ = pc, file, line, ok
		f := runtime.FuncForPC(pc)
		_ = f
		n := f.Name()
		_ = n
		f.Entry()
		fmt.Println("hi")
	}
	info := runtimex.GetFuncInfo(fn)
	_ = info

	var task = &OneShot{}
	tt := reflect.TypeOf(task.init)
	_ = tt
	taskType := reflect.TypeOf(task)
	for i := 0; i < taskType.NumMethod(); i++ {
		fmt.Println(taskType.Method(i).Name)
	}
	method, _ := taskType.MethodByName("Dequeue")
	_ = method

	taskInfo := runtimex.GetMethodSlow(task, uintptr(PollToPollFnPointer(task)), "Dequeue")
	_ = taskInfo

	bp := NewBlockingPool(2, 1024)
	var wg sync.WaitGroup
	wg.Add(1)
	bp.Enqueue(func() {
		defer wg.Done()
		logger.Warn("invoked")
	})
	wg.Wait()
}

func TestBlockingConcurrent(t *testing.T) {
	//bp := blocking
	bp := NewBlockingPool(32, 1024*64)

	wg := new(sync.WaitGroup)
	numThreads := runtime.GOMAXPROCS(0) * 2
	numThreads = 2
	iterations := 10000000
	totalIterations := (numThreads * iterations) + iterations
	finalCount := int64(totalIterations - 1)
	c := new(counter.Counter)
	dispatched := new(counter.Counter)
	overflowCount := new(counter.Counter)
	fn := func() {
		c.Incr()
	}

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for x := 0; x < iterations; x++ {
				dispatched.Incr()
				if !bp.Enqueue(fn) {
					c.Incr()
					overflowCount.Incr()
				}
			}
		}()
	}

	start := timex.NewStopWatch()

	for x := 0; x < iterations; x++ {
		dispatched.Incr()
		if !bp.Enqueue(fn) {
			c.Incr()
			overflowCount.Incr()
		}
	}

	for c.Load() < finalCount {
		runtime.Gosched()
	}
	elapsed := start.ElapsedDur()

	t.Log("final count", c.Load(), "overflow", overflowCount.Load(), "duration", elapsed.String(), "per op", (elapsed / time.Duration(c.Load())).String())
}

func BenchmarkBlockingPool(b *testing.B) {
	//Init(1, Millis500, 1000000, 100000)
	//bp := blocking

	const parallelism = 64
	const maxWorkers = 8192

	fmt.Println("MAX PROCS", runtime.GOMAXPROCS(0))

	b.Run("kirana - Sleep 1ms", func(b *testing.B) {
		wg := &sync.WaitGroup{}
		bp := NewBlockingPool(maxWorkers, 256)

		c := new(counter.Counter)
		errs := new(counter.Counter)
		fn := func() {
			time.Sleep(time.Millisecond)
			wg.Done()
		}

		c.Store(0)
		b.SetParallelism(parallelism)

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				wg.Add(1)
				if !bp.Enqueue(fn) {
					wg.Done()
					c.Incr()
					errs.Incr()
				}
			}
		})

		wg.Wait()
		b.StopTimer()

		fmt.Println("Kirana Errors", errs.Load())
	})

	b.Run("ants - Sleep 1ms", func(b *testing.B) {
		wg := &sync.WaitGroup{}
		c := new(counter.Counter)
		errs := new(counter.Counter)
		fn := func() {
			time.Sleep(time.Millisecond)
			wg.Done()
		}

		wp, err := ants.NewPool(maxWorkers)
		if err != nil {
			b.Fatal(err)
		}

		c.Store(0)
		b.SetParallelism(parallelism)

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				wg.Add(1)
				if wp.Submit(fn) != nil {
					wg.Done()
					c.Incr()
					errs.Incr()
				}
			}
		})

		//for i := 0; i < b.N; i++ {
		//	if !bp.Invoke(fn) {
		//		c.Incr()
		//		panic("could not invoke")
		//		//runtime.Gosched()
		//		//for !bp.Invoke(fn) {
		//		//	runtime.Gosched()
		//		//}
		//	}
		//}

		wg.Wait()

		//for c.Load() < int64(b.N-1) {
		//	runtime.Gosched()
		//	//fmt.Println(c.Load(), b.N-1)
		//}

		b.StopTimer()

		fmt.Println("Ants Errors", errs.Load())
	})

	b.Run("kirana - Sleep 50ms", func(b *testing.B) {
		wg := &sync.WaitGroup{}
		bp := NewBlockingPool(maxWorkers, 256)

		c := new(counter.Counter)
		errs := new(counter.Counter)
		fn := func() {
			time.Sleep(time.Millisecond * 50)
			wg.Done()
		}

		c.Store(0)
		b.SetParallelism(parallelism)

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				wg.Add(1)
				if !bp.Enqueue(fn) {
					wg.Done()
					c.Incr()
					errs.Incr()
				}
			}
		})

		wg.Wait()
		b.StopTimer()

		fmt.Println("Kirana Errors", errs.Load())
	})

	b.Run("ants - Sleep 50ms", func(b *testing.B) {
		wg := &sync.WaitGroup{}
		c := new(counter.Counter)
		errs := new(counter.Counter)
		fn := func() {
			time.Sleep(time.Millisecond * 50)
			wg.Done()
		}

		wp, err := ants.NewPool(maxWorkers)
		if err != nil {
			b.Fatal(err)
		}

		c.Store(0)
		b.SetParallelism(parallelism)

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				wg.Add(1)
				if wp.Submit(fn) != nil {
					wg.Done()
					c.Incr()
					errs.Incr()
				}
			}
		})

		//for i := 0; i < b.N; i++ {
		//	if !bp.Invoke(fn) {
		//		c.Incr()
		//		panic("could not invoke")
		//		//runtime.Gosched()
		//		//for !bp.Invoke(fn) {
		//		//	runtime.Gosched()
		//		//}
		//	}
		//}

		wg.Wait()

		//for c.Load() < int64(b.N-1) {
		//	runtime.Gosched()
		//	//fmt.Println(c.Load(), b.N-1)
		//}

		b.StopTimer()

		fmt.Println("Ants Errors", errs.Load())
	})

	b.Run("kirana - no sleep", func(b *testing.B) {
		wg := &sync.WaitGroup{}
		bp := NewBlockingPool(runtime.GOMAXPROCS(0)*512, 256)

		c := new(counter.Counter)
		errs := new(counter.Counter)
		fn := func() {
			wg.Done()
		}
		//for i := 0; i < b.N; i++ {
		//	bp.Invoke(fn)
		//}
		//
		//for c.Load() < int64(b.N-1) {
		//	runtime.Gosched()
		//	//fmt.Println(c.Load(), b.N-1)
		//}

		c.Store(0)
		b.SetParallelism(parallelism)

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				wg.Add(1)
				if !bp.Enqueue(fn) {
					wg.Done()
					c.Incr()
					errs.Incr()
				}
			}
		})

		//for i := 0; i < b.N; i++ {
		//	if !bp.Invoke(fn) {
		//		c.Incr()
		//		panic("could not invoke")
		//		//runtime.Gosched()
		//		//for !bp.Invoke(fn) {
		//		//	runtime.Gosched()
		//		//}
		//	}
		//}

		wg.Wait()
		//for c.Load()+errs.Load() < int64(b.N-1) {
		//	runtime.Gosched()
		//	//fmt.Println(c.Load(), b.N-1)
		//}

		b.StopTimer()

		fmt.Println("Kirana Errors", errs.Load())
	})

	b.Run("ants - no sleep", func(b *testing.B) {
		wg := &sync.WaitGroup{}
		c := new(counter.Counter)
		errs := new(counter.Counter)
		fn := func() {
			wg.Done()
		}

		c.Store(0)
		b.SetParallelism(parallelism)

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				wg.Add(1)
				if ants.Submit(fn) != nil {
					wg.Done()
					c.Incr()
					errs.Incr()
				}
			}
		})

		//for i := 0; i < b.N; i++ {
		//	if !bp.Invoke(fn) {
		//		c.Incr()
		//		panic("could not invoke")
		//		//runtime.Gosched()
		//		//for !bp.Invoke(fn) {
		//		//	runtime.Gosched()
		//		//}
		//	}
		//}

		wg.Wait()

		//for c.Load() < int64(b.N-1) {
		//	runtime.Gosched()
		//	//fmt.Println(c.Load(), b.N-1)
		//}

		b.StopTimer()

		fmt.Println("Ants Errors", errs.Load())
	})

	//b.Run("ultrapool", func(b *testing.B) {
	//	wg := &sync.WaitGroup{}
	//	c := new(counter.Counter)
	//	errs := new(counter.Counter)
	//	fn := func() {
	//		time.Sleep(time.Millisecond)
	//		wg.Done()
	//	}
	//
	//	wp := ultrapool.NewWorkerPool[func()](func(task func()) {
	//		task()
	//	})
	//
	//	c.Store(0)
	//	b.SetParallelism(parallelism)
	//
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	b.RunParallel(func(pb *testing.PB) {
	//		for pb.Next() {
	//			wg.Add(1)
	//			if wp.AddTask(fn) != nil {
	//				wg.Done()
	//				time.Sleep(time.Millisecond)
	//				c.Incr()
	//				errs.Incr()
	//			}
	//		}
	//	})
	//
	//	wg.Wait()
	//	b.StopTimer()
	//
	//	fmt.Println("Ultrapool Errors", errs.Load())
	//})

	//fmt.Println("workers", len(bp.workers), " wakes", bp.workers[0].WakeCount(), " wake chan full count", bp.queue.WakeChanFullCount())
}
