package mpmc

import (
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/runtimex"
	"github.com/moontrade/kirana/pkg/timex"
	"math"
	"runtime"
	"sync"
	"testing"
	"time"
)

type Task struct {
	c *counter.Counter
}

func BenchmarkMPSC(b *testing.B) {
	//rb := New[Task](1024, nil)
	//rb := NewTask(1024, nil)
	rb := NewBounded[Task](1024)
	//rb := New[*Task](1024, nil)

	c := counter.Counter(0)
	task := &Task{c: &c}
	//ptr := unsafe.Pointer(task)

	//go func() {
	//	<-rb.waker
	//	c.Incr()
	//	if c.Load()%1000 == 0 {
	//		fmt.Println(c.Load())
	//	}
	//}()
	rb.Enqueue(task)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Enqueue(task)
		rb.Dequeue()
		//<-rb.waker
		//if t != task {
		//	b.Fatal("bad")
		//}
		//t, ok = rb.Dequeue()
		//if ok {
		//
		//}
		//rb.DequeueMany(func(v int) {})
	}
	b.StopTimer()
	//fmt.Println("Wake Count", rb.wakeCount.Load())
	//fmt.Println("Wake Full Count", rb.wakeFull.Load())
	//fmt.Println(rb.wakeCount.Load())
}

func TestMPSC(t *testing.T) {
	//rb := New[Task](1024, nil)
	//rb := NewTask(1024, nil)
	rb := NewBounded[Task](32)
	//rb := New[*Task](1024, nil)

	c := counter.Counter(0)
	task := &Task{c: &c}
	//ptr := unsafe.Pointer(task)

	//go func() {
	//	<-rb.waker
	//	c.Incr()
	//	if c.Load()%1000 == 0 {
	//		fmt.Println(c.Load())
	//	}
	//}()
	//rb.Enqueue(task)

	for i := 0; i < 64; i++ {
		rb.Enqueue(task)
		rb.Dequeue()
		//<-rb.waker
		//if t != task {
		//	b.Fatal("bad")
		//}
		//t, ok = rb.Dequeue()
		//if ok {
		//
		//}
		//rb.DequeueMany(func(v int) {})
	}

	//fmt.Println("Wake Count", rb.wakeCount.Load())
	//fmt.Println("Wake Full Count", rb.wakeFull.Load())
	//fmt.Println(rb.wakeCount.Load())
}

func TestConcurrent(t *testing.T) {
	c := new(counter.Counter)
	dispatched := new(counter.Counter)
	overflowCount := new(counter.Counter)
	fn := func() {
		c.Incr()
	}

	bp := NewBounded[func()](100000000)
	//bp := NewSharded[func()](8, 16384)

	wg := new(sync.WaitGroup)
	numConsumers := 1
	numProducers := runtime.GOMAXPROCS(0) * 2
	numProducers = 1
	iterations := 100000000
	totalIterations := numProducers * iterations
	finalCount := int64(totalIterations - 1)

	fnp := runtimex.FuncToPointer(fn)

	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for x := 0; x < iterations; x++ {
				dispatched.Incr()
				for !bp.EnqueueUnsafe(fnp) {
					runtime.Gosched()
				}
				//if !bp.EnqueueUnsafe(fnp) {
				//	c.Incr()
				//	overflowCount.Incr()
				//}
			}
		}()
	}

	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			onTask := func(fn func()) {
				fn()
			}

			for {
				//t := bp.DequeueDeref()
				//if t != nil {
				//	t()
				//} else {
				//	runtime.Gosched()
				//}
				bp.DequeueManyDeref(math.MaxUint32, onTask)
			}
		}()
	}

	start := timex.NewStopWatch()

	//for x := 0; x < iterations; x++ {
	//	dispatched.Incr()
	//	if !bp.EnqueueUnsafe(fnp) {
	//		c.Incr()
	//		overflowCount.Incr()
	//	}
	//}

	for c.Load() < finalCount {
		runtime.Gosched()
	}
	elapsed := start.ElapsedDur()

	t.Log("final count", c.Load(), "overflow", overflowCount.Load(), "duration", elapsed.String(), "per op", (elapsed / time.Duration(c.Load())).String())
}
