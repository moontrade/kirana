package reactor

import (
	logger "github.com/moontrade/log"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	ticker := StartTicker(time.Millisecond * 20)
	//var ms runtime.MemStats

	run := func(dur time.Duration) {
		ln, err := ticker.Register(dur, nil)
		if err != nil {
			t.Fatal(err)
		}
		for tick := range ln.Chan() {
			//logger.Debug("tick", tick.Tick, "dur", tick.Dur)
			if tick.Dur == time.Second {
				logger.Debug("Avg Per Tick", time.Duration(ticker.ticksDur)/time.Duration(ticker.ticks.Load()))
				//printMemStat(ms)
			}
		}
	}

	go run(time.Millisecond * 250)
	go run(time.Millisecond * 500)
	go run(time.Second)

	time.Sleep(time.Hour)
}
