package reactor

import (
	"fmt"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/timex"
	"runtime"
	"sync"
	"testing"
	"time"
)

var Started = timex.NanoTime()

func BenchmarkInvoke(b *testing.B) {
	w, err := NewReactor(Config{Level1Wheel: NewWheel(Millis250), InvokeQSize: 10000000, LockOSThread: false})
	if err != nil {
		b.Fatal(err)
	}
	w.Start()

	var c = counter.Counter(0)
	var overflow = counter.Counter(0)
	c.Store(0)

	t, err := w.Spawn(&SimpleTask{c: &c})
	_ = t

	var fn = func() {
		c.Incr()
	}

	ffn := func() {
		c.Incr()
	}
	_ = fn
	_ = ffn

	//ptr := unsafe.Pointer(&onEntryFn)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//w.Wake(next)
		//w.Invoke(ffn)
		//w.Invoke(func() {
		//	c.Incr()
		//})

		if !w.Invoke(ffn) {
			c.Incr()
			overflow.Incr()
		}
		//w.InvokeRef(&ffn)
		//w.InvokeRef(FuncToPointer(ffn))

		//w.InvokePtr(unsafe.Pointer(&onEntryFn))
		//w.InvokeRef(onEntryFn)
		//w.Invoke(simpleFn)
		//w.Invoke(func() {
		//	c.Incr()
		//})
	}

	fmt.Println(c.Load(), b.N)

	for c.Load() < int64(b.N-1) {
		runtime.Gosched()
	}
	fmt.Println(c.Load(), b.N-1, "overflow", overflow.Load())
}

func TestLoopInvokeConcurrent(t *testing.T) {
	w, err := NewReactor(Config{Level1Wheel: NewWheel(Millis250), InvokeQSize: 10000000, LockOSThread: false})
	if err != nil {
		t.Fatal(err)
	}
	w.Start()
	wg := new(sync.WaitGroup)
	numThreads := runtime.GOMAXPROCS(0) * 2
	numThreads = 32
	iterations := 1000000
	totalIterations := numThreads * iterations
	finalCount := int64(totalIterations - 1)
	c := new(counter.Counter)
	dispatches := new(counter.Counter)
	overflowCount := new(counter.Counter)
	fn := func() {
		c.Incr()
	}
	startWg := new(sync.WaitGroup)

	for i := 0; i < numThreads-1; i++ {
		wg.Add(1)
		startWg.Add(1)
		go func() {
			defer wg.Done()
			startWg.Done()
			//runtime.LockOSThread()
			//defer runtime.UnlockOSThread()

			for x := 0; x < iterations; x++ {
				dispatches.Incr()
				if !w.Invoke(fn) {
					overflowCount.Incr()
					c.Incr()
				}
			}
		}()
	}

	startWg.Wait()
	start := timex.NewStopWatch()

	for x := 0; x < iterations; x++ {
		dispatches.Incr()
		if !w.Invoke(fn) {
			overflowCount.Incr()
			c.Incr()
		}
	}

	for c.Load() < finalCount {
		runtime.Gosched()
	}
	elapsed := start.ElapsedDur()

	t.Log("final count", c.Load(), "overflow count", overflowCount.Load(),
		"wakes", w.invokeQ.WakeCount(), "wake miss", w.invokeQ.WakeChanFullCount(),
		"duration", elapsed.String(), "per op", (elapsed / time.Duration(c.Load())).String())
}

func TestReactor(t *testing.T) {
	tasks := 500
	slotSize := 500
	w, err := NewReactor(Config{Level1Wheel: NewWheel(Millis25)})
	if err != nil {
		t.Fatal(err)
	}
	w.Start()

	c := new(counter.Counter)

	for i := 0; i < tasks; i++ {
		if i%slotSize == 0 {
			time.Sleep(w.tickDur + (time.Microsecond * 25))
		}
		_, err = w.SpawnInterval(&SimpleTask{c: c}, time.Millisecond*100)
		if err != nil {
			t.Fatal(err)
		}
	}

	go func() {
		var ms runtime.MemStats
		for {
			time.Sleep(time.Second * 3)
			//w.Print()
			printMemStat(&ms)
			fmt.Println(c.Load())
			//w.Spawn(&OneShot{c: c})
			//onEntryFn := func() {
			//
			//}
			//w.Invoke(onEntryFn)
			//w.Invoke(func() {
			//	fmt.Println("invoked")
			//})

			//w.Invoke(func() {
			//	fmt.Println("invoked")
			//})
			//w.Invoke(&OneShot{c: c})
			//w.Invoke(&OneShot{c: c})
			//w.Invoke(&OneShot{c: c})
			//w.Invoke(&OneShot{c: c})
		}
	}()

	time.Sleep(time.Hour)
}

func printMemStat(ms *runtime.MemStats) {
	runtime.ReadMemStats(ms)
	//fmt.Println("--------------------------------------")
	//fmt.Println("Memory Statistics Reporting Time: 	", Time.Now())
	//fmt.Println("--------------------------------------")
	fmt.Println()
	fmt.Println("Uptime:									", time.Duration(timex.NanoTime()-Started))
	//fmt.Println("Uptime:									", time.Duration(timex.NanoTime()-Started).Truncate(time.Second)%5)
	fmt.Println("Bytes of allocated heap objects: 		", ms.Alloc)
	fmt.Println("Total bytes of Heap object: 			", ms.TotalAlloc)
	fmt.Println("Bytes of memory obtained from OS: 		", ms.Sys)
	fmt.Println("Count of heap objects: 					", ms.Mallocs)
	fmt.Println("Count of heap objects freed: 			", ms.Frees)
	fmt.Println("Count of live heap objects:				", ms.Mallocs-ms.Frees)
	fmt.Println("Number of GC cycles:					", ms.NumGC)
	fmt.Println("Number of GC CPU:						", ms.GCCPUFraction)
	//fmt.Println("Number of GC Pause:			", time.Duration(ms.PauseTotalNs))
	//fmt.Println("Number of Next GC:			", ms.NextGC)
	fmt.Println("Number of GC Pause / GC:				", time.Duration(float64(ms.PauseTotalNs)/float64(ms.NumGC)))
	//fmt.Println("--------------------------------------")
	fmt.Println()
	//if (time.Duration(timex.NanoTime()-Started).Truncate(time.Second))%(time.Second*5) == 0 {
	//	runtime.GC()
	//}
	runtime.GC()
}
