package mpmc

import (
	"fmt"
	"github.com/moontrade/kirana/pkg/counter"
	"testing"
)

func BenchmarkMPMCWake(b *testing.B) {
	//rb := New[Task](1024, nil)
	//rb := NewTask(1024, nil)
	rb := NewBoundedWake[Task](1024, nil)
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

func TestMPMCWake(t *testing.T) {
	//rb := New[Task](1024, nil)
	//rb := NewTask(1024, nil)
	rb := NewBoundedWake[Task](32, nil)
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

	fmt.Println("Wake Count", rb.wakeCount.Load())
	fmt.Println("Wake Full Count", rb.wakeFull.Load())
	fmt.Println(rb.wakeCount.Load())
}
