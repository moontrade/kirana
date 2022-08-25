package reactor

import (
	"fmt"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/timex"
	logger "github.com/moontrade/log"
	"github.com/panjf2000/ants"
	"runtime"
	"testing"
	"time"
)

func TestWorker(t *testing.T) {
	count := 10000
	preload := int(5000)
	queueSize := int(count * 2)
	if queueSize < preload {
		queueSize = preload * 2
	}

	wp := NewWorkerPool(0, preload, queueSize, queueSize, time.Second, time.Second)

	failed := new(counter.Counter)
	c := new(counter.Counter)
	fn := func() {
		c.Incr()
	}

	for i := 0; i < count; i++ {
		if wp.GoRef(&fn) != nil {
			c.Incr()
			failed.Incr()
		}
	}
	for c.Load() < int64(count-1) {
		runtime.Gosched()
	}
	c.Store(0)

	fmt.Println("loaded")
	start := timex.NanoTime()

	for i := 0; i < count; i++ {
		wp.GoRef(&fn)
	}

	for c.Load() != int64(count-1) {
		runtime.Gosched()
		//Time.Sleep(Time.Microsecond * 50)
	}

	elapsed := timex.NanoTime() - start
	logger.Warn("dur", time.Duration(elapsed), "dur_per", time.Duration(elapsed)/time.Duration(count))

	time.Sleep(time.Hour)
}

func BenchmarkWorker(b *testing.B) {
	preload := int(8192)
	queueSize := int(b.N + 8)
	if queueSize < preload {
		queueSize = preload * 2
	}
	wp := NewWorkerPool(8192, 8192, queueSize, queueSize, time.Second, time.Second)

	failed := new(counter.Counter)
	c := new(counter.Counter)
	fn := func() {
		c.Incr()
	}

	for i := 0; i < preload; i++ {
		if wp.GoRef(&fn) != nil {
			c.Incr()
			failed.Incr()
		}
	}
	for c.Load() < int64(preload-1) {
		runtime.Gosched()
	}
	c.Store(0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if wp.GoRef(&fn) != nil {
			c.Incr()
			failed.Incr()
		}
	}

	for c.Load() < int64(b.N)/2 {
		runtime.Gosched()
		//Time.Sleep(Time.Millisecond * 500)
		//fmt.Println("PROGRESS: failed count", failed.Load(), "iterations", c.Load(), "of", b.N)
	}

	b.StopTimer()
	fmt.Println("failed count", failed.Load(), "iterations", c.Load(), "of", b.N)
	_ = wp.Close()
}

func BenchmarkGo(b *testing.B) {
	preload := int(b.N)
	c := new(counter.Counter)

	fn := func() {
		c.Incr()
	}
	b.ReportAllocs()
	for i := 0; i < preload; i++ {
		//go func() {
		//	c.Incr()
		//}()
		go fn()
	}
	for c.Load() < int64(preload-1) {
		runtime.Gosched()
	}
	c.Store(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//go func() {
		//	c.Incr()
		//}()

		go fn()
	}

	for c.Load() < int64(b.N)-1 {
		runtime.Gosched()
	}

	b.StopTimer()
}

func BenchmarkAnts(b *testing.B) {
	preload := int(8192)
	pool, _ := ants.NewPool(preload)
	c := new(counter.Counter)

	fn := func() {
		c.Incr()
	}
	for i := 0; i < preload; i++ {
		pool.Submit(fn)
	}
	for c.Load() < int64(preload-1) {
		runtime.Gosched()
	}
	c.Store(0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//go func() {
		//	c.Incr()
		//}()
		pool.Submit(fn)
	}

	for c.Load() < int64(b.N)-1 {
		runtime.Gosched()
	}

	b.StopTimer()
}

func BenchmarkGopool(b *testing.B) {
	preload := int(4096)
	//wp := gopool.NewPool("pool", int32(preload), nil)

	c := new(counter.Counter)
	fn := func() {
		c.Incr()
	}

	for i := 0; i < preload; i++ {
		gopool.Go(fn)
	}
	for c.Load() < int64(preload-1) {
		runtime.Gosched()
	}
	c.Store(0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gopool.Go(fn)
	}

	for c.Load() < int64(b.N)-1 {
		runtime.Gosched()
	}

	b.StopTimer()
}
