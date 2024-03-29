package netpoll

import (
	"fmt"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/mpmc"
	"math"
	"runtime"
	"testing"
	"time"
)

type Task struct {
	fd int
}

func TestNetpoll(t *testing.T) {
	poll := OpenPoll[Task]()

	onEvent := func(index, count, fd int, filter int16, conn *Task) error {
		if index == -1 {

			return nil
		}
		if index == count-1 {

		}
		return nil
	}

	onLoop := func(count int) (time.Duration, error) {
		//end = timex.NanoTime()
		if count == 0 {
			fmt.Println("timeout")
		}
		return time.Second, nil
	}

	go func() {
		if err := poll.Wait(time.Second, onEvent, onLoop); err != nil {
			panic(err)
		}
		fmt.Println("test")
	}()

	go func() {
		for {
			time.Sleep(time.Second * 2)
			_ = poll.Wake()
		}
	}()

	time.Sleep(time.Hour)
}

func BenchmarkWake(b *testing.B) {
	poll := OpenPoll[Task]()
	c := new(counter.Counter)
	fn := func() {
		c.Incr()
	}

	ch := make(chan int64, 1)

	queue := mpmc.NewBoundedWake[func()](int64(b.N)*2, ch)

	flushTasks := func(fn *func()) {
		(*fn)()
	}

	onEvent := func(index, count, fd int, filter int16, conn *Task) error {
		if fd == 0 {
			queue.DequeueMany(math.MaxUint32, flushTasks)
			return nil
		}
		return nil
	}

	onLoop := func(count int) (time.Duration, error) {
		queue.DequeueMany(math.MaxUint32, flushTasks)

		//end = timex.NanoTime()
		if count == 0 {
			//fmt.Println("timeout")

		}
		return time.Second, nil
	}

	go func() {
		for {
			select {
			case v, ok := <-ch:
				//fmt.Println("woke")
				if !ok {
					return
				}
				_ = v
				if err := poll.Wake(); err != nil {
					b.Fatal(err)
				}
			}
		}
	}()

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		//queue.DequeueMany(math.MaxUint32, func(fn *func()) bool {
		//	(*fn)()
		//	return true
		//})

		err := poll.Wait(time.Hour, onEvent, onLoop)
		if err != nil {
			b.Fatal(err)
		}
		//panic(err)
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Enqueue(&fn)
		//_ = poll.Wake()
	}

	//_ = poll.Wake()

	for c.Load() <= int64(b.N-1) {
		runtime.Gosched()
		//fmt.Println(c.Load(), b.N-1, queue.Len())
	}
}
