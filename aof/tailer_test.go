package aof

import (
	"encoding/json"
	"fmt"
	"github.com/moontrade/kirana/pkg/timex"
	"github.com/moontrade/kirana/reactor"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"testing"
	"time"
)

func init() {
	reactor.Init(0, reactor.Millis250, 8192*8, 8192)
}

var timer int64

func TestTailer(t *testing.T) {
	debug.SetMemoryLimit(1024 * 1024 * 128)

	//runtime.LockOSThread()

	m, err := NewManager("testdata", 0755, 0444)
	if err != nil {
		t.Fatal(err)
	}

	buf := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '0', '1', '2', '3', '4', '5'}

	for x := 0; x < 32; x++ {
		time.Sleep(time.Millisecond * 100)
		name := fmt.Sprintf("db-%d.txt", x)
		os.Remove("testdata/" + name)

		const SIZE = 1024 * 1024 * 1
		f, err := m.Open(name, *CreateFile().WithSizeNow(SIZE * 2), RecoveryDefault)
		if err != nil {
			t.Fatal(err)
		}
		reader := &Reader{AOF: f, log: make([]int64, 0, 32768)}
		tailer, err := f.Subscribe(reader)
		if err != nil {
			t.Fatal(err)
		}
		_ = tailer
		f.Wake()
		//time.Sleep(time.Millisecond)
		wakeAttempts := 0
		if f.readOnly {
			//time.Sleep(time.Second)
		} else {
			last := timex.NanoTime()
			atomic.StoreInt64(&timer, last)
			for {
				//time.Sleep(time.Microsecond * 50)
				//next := timex.NanoTime()
				//atomic.StoreInt64(&timer, next)
				atomic.StoreInt64(&timer, timex.NanoTime())
				_, _ = f.Write(buf)
				wakeAttempts++

				if f.size >= SIZE-32 {
					//runtime.Gosched()
					break
				}

				//runtime.Gosched()

				if f.size%(1024*1024) == 0 {
					//runtime.Gosched()
					//f.Flush()
					//f.Sync()
					//fmt.Println(toJson(m.stats))
				}
			}
		}

		atomic.CompareAndSwapInt64(&timer, 0, timex.NanoTime())

		for len(reader.log) == 0 {
			runtime.Gosched()
		}
		//f.Close()
		//fmt.Println("done writing", atomic.LoadInt64(&f.size))

		//time.Sleep(time.Millisecond * 500)
		if x > 0 {
			printLatency(wakeAttempts, reader.log)
		}
	}
}

func printLatency(attempts int, l []int64) {
	var (
		sum int64
		min int64 = math.MaxInt
		max int64
	)
	for _, latency := range l {
		sum += latency
		if latency < min {
			min = latency
		}
		if latency > max {
			max = latency
		}
	}
	fmt.Println("Wake Attempts", attempts, "Wakes", len(l), "Min", time.Duration(min), "Max", time.Duration(max), "Avg", div(time.Duration(sum), time.Duration(len(l))))
}

func div(a, b time.Duration) time.Duration {
	if a == 0 || b == 0 {
		return 0
	}
	return a / b
}

type Reader struct {
	*AOF
	Consumer
	last int64
	log  []int64
}

func (r *Reader) PollRead(event ReadEvent) (int64, error) {
	started := atomic.SwapInt64(&timer, timex.NanoTime())
	r.last = started
	next := timex.NanoTime()
	//for started == 0 {
	//	started = atomicx.Xchgint64(&timer, 0)
	//}
	elapsed := next - started
	//elapsed := event.Time - started
	r.log = append(r.log, elapsed)
	//fmt.Println("GID", gid.GID(), "PID", gid.PID(), "read:", event.Begin, event.End, event.EOF, " since write:", time.Duration(elapsed))
	if event.EOF {
		return 0, reactor.ErrStop
	}
	return event.End, nil
}

func (r *Reader) PollReadClosed(reason error) {
	fmt.Println("reader closed", reason)
}

func toJson(v any) string {
	d, err := json.MarshalIndent(v, "", "   ")
	if err != nil {
		panic(err)
	}
	return string(d)
}
