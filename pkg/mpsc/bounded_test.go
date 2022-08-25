package mpsc

import (
	"github.com/moontrade/kirana/pkg/counter"
	"testing"
)

type Task struct {
	c *counter.Counter
}

func BenchmarkMPSC(b *testing.B) {
	//rb := New[Task](1024, nil)
	//rb := NewTask(1024, nil)
	rb := NewBounded[Task](1024, nil)
	//rb := New[*Task](1024, nil)

	c := counter.Counter(0)
	task := &Task{c: &c}
	//ptr := unsafe.Pointer(task)

	//go func() {
	//	<-rb.wakeCh
	//	c.Incr()
	//	if c.Load()%1000 == 0 {
	//		fmt.Println(c.Load())
	//	}
	//}()
	rb.Push(task)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Push(task)
		rb.Pop()
		//<-rb.wakeCh
		//if t != task {
		//	b.Fatal("bad")
		//}
		//t, ok = rb.Dequeue()
		//if ok {
		//
		//}
		//rb.PopMany(func(v int) {})
	}
	b.StopTimer()
	//fmt.Println("Wake Count", rb.wakeCount.Load())
	//fmt.Println("Wake Full Count", rb.wakeFull.Load())
	//fmt.Println(rb.wakeCount.Load())
}

func TestMPSC(t *testing.T) {
	//rb := New[Task](1024, nil)
	//rb := NewTask(1024, nil)
	rb := NewBounded[Task](32, nil)
	//rb := New[*Task](1024, nil)

	c := counter.Counter(0)
	task := &Task{c: &c}
	//ptr := unsafe.Pointer(task)

	//go func() {
	//	<-rb.wakeCh
	//	c.Incr()
	//	if c.Load()%1000 == 0 {
	//		fmt.Println(c.Load())
	//	}
	//}()
	rb.Push(task)

	for i := 0; i < 64; i++ {
		rb.Push(task)
		rb.Pop()
		//<-rb.wakeCh
		//if t != task {
		//	b.Fatal("bad")
		//}
		//t, ok = rb.Dequeue()
		//if ok {
		//
		//}
		//rb.PopMany(func(v int) {})
	}

	//fmt.Println("Wake Count", rb.wakeCount.Load())
	//fmt.Println("Wake Full Count", rb.wakeFull.Load())
	//fmt.Println(rb.wakeCount.Load())
}
