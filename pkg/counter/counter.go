package counter

import (
	"github.com/moontrade/wormhole/pkg/atomicx"
	"github.com/moontrade/wormhole/pkg/timex"
	"sync/atomic"
	"time"
)

type Counter int64

func (c *Counter) Load() int64 {
	return atomic.LoadInt64((*int64)(c))
}

func (c *Counter) Incr() int64 {
	return atomicx.Xaddint64((*int64)(c), 1)
}

func (c *Counter) Decr() int64 {
	return atomicx.Xaddint64((*int64)(c), -1)
}

func (c *Counter) Add(count int64) {
	atomicx.Xaddint64((*int64)(c), count)
}

func (c *Counter) Cas(old, new int64) bool {
	return atomicx.Casint64((*int64)(c), old, new)
}

func (c *Counter) Sub(count int64) {
	if count > 0 {
		count = -count
	}
	atomicx.Xaddint64((*int64)(c), count)
}

func (c *Counter) Store(value int64) {
	atomic.StoreInt64((*int64)(c), value)
}

type TimeCounter int64

func (c *TimeCounter) Load() int64 {
	return atomic.LoadInt64((*int64)(c))
}

func (c *TimeCounter) Since(s timex.StopWatch) {
	atomicx.Xaddint64((*int64)(c), s.Stop())
}

func (c *TimeCounter) Store(count int64) {
	atomic.StoreInt64((*int64)(c), count)
}

func (c *TimeCounter) Add(count int64) {
	atomicx.Xaddint64((*int64)(c), count)
}

func (c *TimeCounter) Plus(counter TimeCounter) {
	atomicx.Xaddint64((*int64)(c), int64(counter))
}

func (c *TimeCounter) Duration() time.Duration {
	return time.Duration(*c)
}

func (c *TimeCounter) Cas(old, new int64) bool {
	return atomicx.Casint64((*int64)(c), old, new)
}

func (c *TimeCounter) CasDuration(old, new time.Duration) bool {
	return atomicx.Casint64((*int64)(c), int64(old), int64(new))
}
