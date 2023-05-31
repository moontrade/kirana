package reactor

import (
	logger "github.com/moontrade/log"
	"runtime"
	"testing"
	"time"
)

func TestHFTicker(t *testing.T) {
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		tick := time.Millisecond * 10
		//tick = time.Millisecond * 500
		//next := timex.NanoTime() + int64(tick)

		for {
			//begin := timex.NanoTime()
			//time.Sleep(time.Duration(next - begin))
			//next += int64(tick)
			park(tick)
			//time.Sleep(tick)
			//cgo.NonBlocking((*byte)(cgo2.Sleep), uintptr(tick), 0)
			//fmt.Println(timex.NanoTime())
			//end := timex.NanoTime()
		}
	}()

	time.Sleep(time.Hour)
}

func TestTicker(t *testing.T) {
	ticker := StartTicker(time.Microsecond * 25000)
	//var ms runtime.MemStats
	run := func(dur time.Duration) {
		ch := make(chan int64, 1)
		ln, err := ticker.Register(dur, nil, ch)
		if err != nil {
			t.Fatal(err)
		}
		for v := range ln.Chan() {
			_ = v
			tick := ln.next
			logger.Debug("tick", tick.Tick, "dur", tick.Dur)
			//if tick.Dur == time.Second {
			//	logger.Debug(dur, "Avg Per Tick", time.Duration(ticker.ticksDur)/time.Duration(ticker.ticks.Load()))
			//printMemStat(ms)
			//}
		}
	}

	go run(time.Millisecond * 250)
	go run(time.Millisecond * 500)
	go run(time.Second)

	time.Sleep(time.Hour)
}
