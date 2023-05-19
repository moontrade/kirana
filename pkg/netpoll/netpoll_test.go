package netpoll

import (
	"fmt"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/mpsc"
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

	queue := mpsc.NewBounded[func()](int64(b.N)*2, nil)

	flushTasks := func(fn *func()) {
		(*fn)()
	}

	onEvent := func(index, count, fd int, filter int16, conn *Task) error {
		if fd == 0 {
			queue.PopMany(math.MaxUint32, flushTasks)
			return nil
		}
		return nil
	}

	onLoop := func(count int) (time.Duration, error) {
		queue.PopMany(math.MaxUint32, flushTasks)

		//end = timex.NanoTime()
		if count == 0 {
			//fmt.Println("timeout")

		}
		return time.Second, nil
	}

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		//queue.PopMany(math.MaxUint32, func(fn *func()) bool {
		//	(*fn)()
		//	return true
		//})

		err := poll.Wait(time.Second, onEvent, onLoop)
		panic(err)
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Push(&fn)
		_ = poll.Wake()
		if i%1000000 == 0 {
			//_ = poll.Wake()
		}
	}

	_ = poll.Wake()

	for c.Load() <= int64(b.N-1) {
		runtime.Gosched()
		//fmt.Println(c.Load(), b.N-1, queue.Len())
	}
}
