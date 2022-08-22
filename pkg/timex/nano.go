package timex

import (
	"time"
	_ "unsafe"
)

const NanoTimeCost = time.Nanosecond * 20

//go:noescape
//go:linkname NanoTime runtime.nanotime
func NanoTime() int64

func Since(start int64) int64 {
	return NanoTime() - start
}

func SinceDur(start int64) time.Duration {
	return time.Duration(NanoTime() - start)
}

type StopWatch int64

func NewStopWatch() StopWatch {
	return StopWatch(NanoTime())
}

func (s *StopWatch) Start() {
	*s = StopWatch(NanoTime())
}

func (s *StopWatch) Stop() int64 {
	o := int64(*s)
	n := NanoTime()
	*s = StopWatch(n)
	return n - o
}

func (s *StopWatch) Elapsed() int64 {
	return NanoTime() - int64(*s)
}

func (s *StopWatch) ElapsedMicros() int64 {
	nanos := NanoTime() - int64(*s)
	if nanos <= 0 {
		return 0
	}
	return nanos / 1000
}

func (s *StopWatch) ElapsedMillis() int64 {
	nanos := NanoTime() - int64(*s)
	if nanos <= 0 {
		return 0
	}
	return nanos / 1000000
}

func (s *StopWatch) ElapsedDur() time.Duration {
	return time.Duration(NanoTime() - int64(*s))
}

func (s *StopWatch) StopDur() time.Duration {
	o := int64(*s)
	n := NanoTime()
	*s = StopWatch(n)
	return time.Duration(n - o)
}

//func (s *StopWatch) FramedSleep(frame time.Duration) time.Duration {
//	o := int64(*s)
//	n := NanoTime()
//	expected := o + int64(frame)
//	if n > expected {
//
//	}
//	*s = StopWatch(o + int64(frame))
//
//}
