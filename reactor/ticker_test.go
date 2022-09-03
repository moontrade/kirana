package reactor

import (
	logger "github.com/moontrade/log"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	ticker := StartTicker(time.Millisecond * 250)
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
