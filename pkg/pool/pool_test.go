package pool

import (
	"bytes"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/runtimex"
	"github.com/moontrade/kirana/pkg/syncx"
	"github.com/moontrade/kirana/pkg/timex"
	"runtime"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func TestPtrData(t *testing.T) {
	t.Log("struct with pointers", uint(ptrdataOf(struct {
		b []byte
		s string
		z struct {
			y struct {
				x struct {
					a *bytes.Buffer
					b bytes.Buffer
				}
				s string
			}
		}
		x struct{ x, y int }
		y *bytes.Buffer
		w unsafe.Pointer
		u uintptr
		t unsafe.Pointer
	}{})))
	t.Log("[]*bytes.Buffer", uint(ptrdataOf([]*bytes.Buffer{})))
	t.Log("[]bytes.Buffer", uint(ptrdataOf([]bytes.Buffer{})))
	t.Log("*bytes.Buffer", uint(ptrdataOf(&bytes.Buffer{})))
	t.Log("bytes.Buffer", uint(ptrdataOf(bytes.Buffer{})))
}

func BenchmarkConcurrentPool(b *testing.B) {
	value := make([]byte, 128)
	allocate := func() unsafe.Pointer {
		//value := make([]byte, 128)
		return unsafe.Pointer(&value[0])
	}
	//b.Run("sync.Pool", func(b *testing.B) {
	//	pool := &sync.Pool{New: func() interface{} {
	//		return value
	//	}}
	//
	//	b.ResetTimer()
	//	b.ReportAllocs()
	//	for i := 0; i < b.N; i++ {
	//		pool.Put(pool.Get())
	//	}
	//})

	b.Run("sync.Pool", func(b *testing.B) {
		runtime.GC()
		allocs := new(counter.Counter)
		pool := &sync.Pool{New: func() interface{} {
			allocs.Incr()
			return allocate()
		}}

		go func() {
			for x := 0; x < 1; x++ {
				for i := 0; i < b.N; i++ {
					pool.Put(allocate())
				}
			}
		}()
		runtime.Gosched()
		time.Sleep(time.Millisecond)

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			//if i%2 == 0 {
			//	pool.Put(unsafe.Pointer(&value[0]))
			//}
			//if i%4 == 0 {
			//	pool.Get()
			//	pool.Get()
			//}
			//if i%8 == 0 {
			//	pool.Put(unsafe.Pointer(&value[0]))
			//	pool.Put(unsafe.Pointer(&value[0]))
			//	pool.Put(unsafe.Pointer(&value[0]))
			//	pool.Put(unsafe.Pointer(&value[0]))
			//	pool.Put(unsafe.Pointer(&value[0]))
			//	pool.Put(unsafe.Pointer(&value[0]))
			//}
			//pool.Get()
			pool.Get()
			//pool.Put(pool.Get())
		}
		b.StopTimer()
		b.Log("N", b.N, "Allocs", allocs.Load())
	})

	b.Run("Pool", func(b *testing.B) {
		pool := NewPool[[]byte](Config[[]byte]{
			PageSize:      65535,
			PagesPerShard: 8192,
			AllocFunc: func() unsafe.Pointer {
				return allocate()
			},
		})
		for i := 0; i < 10000; i++ {
			//shard.PutUnsafe(allocate())
			pool.PutUnsafe(allocate())
		}
		//shard := pool.Shard()
		go func() {
			for x := 0; x < 1; x++ {
				for i := 0; i < b.N; i++ {
					//shard.PutUnsafe(allocate())
					pool.PutUnsafe(allocate())
				}
			}
			//for i := 0; i < b.N; i++ {
			//	pool.PutUnsafe(allocate())
			//}
		}()
		runtime.Gosched()
		time.Sleep(time.Millisecond)

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			//if i%2 == 0 {
			//	pool.PutUnsafe(unsafe.Pointer(&value[0]))
			//}
			//if i%4 == 0 {
			//	pool.GetUnsafe()
			//	pool.GetUnsafe()
			//}
			//if i%8 == 0 {
			//	pool.PutUnsafe(unsafe.Pointer(&value[0]))
			//	pool.PutUnsafe(unsafe.Pointer(&value[0]))
			//	pool.PutUnsafe(unsafe.Pointer(&value[0]))
			//	pool.PutUnsafe(unsafe.Pointer(&value[0]))
			//	pool.PutUnsafe(unsafe.Pointer(&value[0]))
			//	pool.PutUnsafe(unsafe.Pointer(&value[0]))
			//}
			//shard.GetUnsafe()
			pool.GetUnsafe()
			//pool.PutUnsafe(pool.GetUnsafe())
		}
		b.StopTimer()
		b.Log("N", b.N, " PageAllocs", pool.PageAllocs, " PageDeallocs", pool.PageDeallocs, " Allocs", pool.Allocs, " Deallocs", pool.Deallocs)
	})

	b.Run("syncx.Pool", func(b *testing.B) {
		runtime.GC()
		allocs := new(counter.Counter)
		pool := &syncx.Pool{NoGC: true, New: func() interface{} {
			allocs.Incr()
			return allocate()
		}}

		go func() {
			for x := 0; x < 1; x++ {
				for i := 0; i < b.N; i++ {
					pool.Put(allocate())
				}
			}
		}()

		runtime.Gosched()
		time.Sleep(time.Millisecond)

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			//if i%2 == 0 {
			//	pool.Put(unsafe.Pointer(&value[0]))
			//}
			//if i%4 == 0 {
			//	pool.Get()
			//	pool.Get()
			//}
			//if i%8 == 0 {
			//	pool.Put(unsafe.Pointer(&value[0]))
			//	pool.Put(unsafe.Pointer(&value[0]))
			//	pool.Put(unsafe.Pointer(&value[0]))
			//	pool.Put(unsafe.Pointer(&value[0]))
			//	pool.Put(unsafe.Pointer(&value[0]))
			//	pool.Put(unsafe.Pointer(&value[0]))
			//}
			pool.Get()
			//pool.Put(pool.Get())
			//pool.Put(pool.Get())
		}
		b.StopTimer()
		b.Log("N", b.N, "Allocs", allocs.Load())
	})
	//b.Run("sync.Pool Concurrent", func(b *testing.B) {
	//	pool := &sync.Pool{New: func() interface{} {
	//		return value
	//	}}
	//
	//	b.RunParallel(func(pb *testing.PB) {
	//		for pb.Next() {
	//			pool.Put(pool.Get())
	//		}
	//	})
	//})
	//b.Run("syncx.Pool Concurrent", func(b *testing.B) {
	//	pool := &syncx.Pool{New: func() interface{} {
	//		return value
	//	}}
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	b.RunParallel(func(pb *testing.PB) {
	//		for pb.Next() {
	//			pool.Put(pool.Get())
	//		}
	//	})
	//})
	//b.Run("syncx.Pool Concurrent Get", func(b *testing.B) {
	//	pool := &syncx.Pool{New: func() interface{} {
	//		return value
	//	}}
	//
	//	go func() {
	//		for i := 0; i < b.N; i++ {
	//			pool.Put(value)
	//		}
	//	}()
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	b.RunParallel(func(pb *testing.PB) {
	//		for pb.Next() {
	//			pool.Get()
	//			//pool.Put(pool.Get())
	//		}
	//	})
	//})
	//
	//b.Run("Pool Concurrent", func(b *testing.B) {
	//	pool := NewPool[func()](Config[func()]{
	//		AllocFunc: func() unsafe.Pointer {
	//			return value
	//		},
	//	})
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	b.RunParallel(func(pb *testing.PB) {
	//		for pb.Next() {
	//			pool.PutUnsafe(pool.GetUnsafe())
	//		}
	//	})
	//})
	//b.Run("Pool Concurrent Get", func(b *testing.B) {
	//	pool := NewPool[func()](Config[func()]{
	//		AllocFunc: func() unsafe.Pointer {
	//			return value
	//		},
	//	})
	//	go func() {
	//		for i := 0; i < b.N; i++ {
	//			pool.PutUnsafe(value)
	//		}
	//	}()
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	b.RunParallel(func(pb *testing.PB) {
	//		for pb.Next() {
	//			pool.Get()
	//		}
	//	})
	//})
}

func TestConcurrentSyncPool(t *testing.T) {
	c := new(counter.Counter)
	dispatched := new(counter.Counter)
	overflowCount := new(counter.Counter)

	fn := func() {
		c.Incr()
	}
	//fnp := runtimex.FuncToPointer(fn)
	pool := &sync.Pool{New: func() interface{} {
		return fn
	}}

	wg := new(sync.WaitGroup)
	numConsumers := 4
	numProducers := runtime.GOMAXPROCS(0) * 2
	numProducers = 1
	iterations := 10000000
	totalIterations := numProducers * iterations
	finalCount := int64(totalIterations - 1)

	start := timex.NewStopWatch()

	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for x := 0; x < iterations; x++ {
				dispatched.Incr()
				pool.Put(fn)
			}
		}()
	}

	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				next := pool.Get().(func())
				next()
				pool.Put(next)
			}
		}()
	}

	for c.Load() < finalCount {
		runtime.Gosched()
	}
	elapsed := start.ElapsedDur()

	t.Log("final count", c.Load(), "overflow", overflowCount.Load(), "duration", elapsed.String(), "per op", (elapsed / time.Duration(c.Load())).String())
}

func TestConcurrentPool(t *testing.T) {
	iterations := 10000000
	numConsumers := 8
	totalIterations := numConsumers * iterations
	finalCount := int64(totalIterations - 1)

	t.Run("syncx.Pool", func(t *testing.T) {
		c := new(counter.Counter)
		//dispatched := new(counter.Counter)
		overflowCount := new(counter.Counter)

		fn := func() {
			c.Incr()
		}
		//fnp := runtimex.FuncToPointer(fn)
		pool := &syncx.Pool{New: func() interface{} {
			return fn
		}}

		wg := new(sync.WaitGroup)

		start := timex.NewStopWatch()

		//for i := 0; i < numProducers; i++ {
		//	wg.Add(1)
		//	go func() {
		//		defer wg.Done()
		//
		//		for x := 0; x < iterations; x++ {
		//			dispatched.Incr()
		//			pool.Put(fn)
		//		}
		//	}()
		//}

		for i := 0; i < numConsumers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for x := 0; x < iterations; x++ {
					next := pool.Get().(func())
					next()
					pool.Put(next)
				}
			}()
		}

		for c.Load() < finalCount {
			runtime.Gosched()
		}
		elapsed := start.ElapsedDur()
		if elapsed <= 0 {
			elapsed = 1
		}

		count := time.Duration(c.Load())
		perOp := time.Duration(0)
		if count > 0 {
			perOp = elapsed / count
		}

		t.Log("sync.Pool", "final count", c.Load(), "overflow", overflowCount.Load(), "duration", elapsed.String(), "per op", perOp.String())
	})
	t.Run("Pool", func(t *testing.T) {
		c := new(counter.Counter)
		//dispatched := new(counter.Counter)
		overflowCount := new(counter.Counter)

		fn := func() {
			c.Incr()
		}
		//fnp := runtimex.FuncToPointer(fn)
		pool := NewPool[func()](Config[func()]{
			PageSize:      8192,
			PagesPerShard: 8192,
			AllocFunc: func() unsafe.Pointer {
				return runtimex.FuncToPointer(fn)
			},
		})

		wg := new(sync.WaitGroup)
		//numProducers := runtime.GOMAXPROCS(0) * 2
		//numProducers = 0

		start := timex.NewStopWatch()

		//for i := 0; i < numProducers; i++ {
		//	wg.Add(1)
		//	go func() {
		//		defer wg.Done()
		//
		//		for x := 0; x < iterations; x++ {
		//			dispatched.Incr()
		//			pool.PutUnsafe(runtimex.FuncToPointer(fn))
		//		}
		//	}()
		//}

		for i := 0; i < numConsumers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				shard := pool.Shard()

				for x := 0; x < iterations; x++ {
					next := runtimex.FuncFromPointer(shard.GetUnsafe())
					//next()
					c.Incr()
					pool.PutUnsafe(runtimex.FuncToPointer(next))
				}
			}()
		}

		for c.Load() < finalCount {
			runtime.Gosched()
		}
		elapsed := start.ElapsedDur()
		if elapsed <= 0 {
			elapsed = 1
		}

		count := time.Duration(c.Load())
		perOp := time.Duration(0)
		if count > 0 {
			perOp = elapsed / count
		}

		t.Log("Pool", "final count", c.Load(), "overflow", overflowCount.Load(), "duration", elapsed.String(), "per op", perOp.String())
	})
}

func TestPool(t *testing.T) {
	allocs := counter.Counter(0)
	//fn := func() {}

	pool := NewPool[bytes.Buffer](
		Config[bytes.Buffer]{
			SizeClass: 64,
			AllocFunc: func() unsafe.Pointer {
				allocs.Incr()
				return unsafe.Pointer(bytes.NewBuffer(make([]byte, 64)))
			},
		},
	)

	v1 := pool.Get()
	v2 := pool.Get()
	v3 := pool.Get()
	v4 := pool.Get()
	v5 := pool.Get()
	v6 := pool.Get()
	v7 := pool.Get()
	v8 := pool.Get()
	v9 := pool.Get()

	pool.Put(v9)
	pool.Put(v8)
	pool.Put(v7)
	pool.Put(v6)
	pool.Put(v5)

	v9 = pool.Get()

	pool.Put(v1)
	pool.Put(v2)
	pool.Put(v3)
	pool.Put(v4)
	pool.Put(v9)

	wg := new(sync.WaitGroup)
	startWg := new(sync.WaitGroup)
	numWorkers := runtime.GOMAXPROCS(0) * 2
	numWorkers = 2
	iterations := 10000000
	totalIterations := numWorkers * iterations
	//finalCount := int64(totalIterations - 1)

	//t.Log("pool size", pool.shard().Len())

	started := timex.NewStopWatch()

	for i := 0; i < numWorkers; i++ {
		s := pool.shards[i]
		_ = s
		wg.Add(1)
		startWg.Add(1)
		go func() {
			defer wg.Done()
			startWg.Done()
			for x := 0; x < iterations; x++ {
				pool.PutUnsafe(s.GetUnsafe())
				//pool.Put(pool.Get())
			}
		}()
	}

	startWg.Wait()

	wg.Wait()

	elapsed := started.ElapsedDur()

	//for i := 0; i < 10; i++ {
	//	pool.Put(pool.Get())
	//	t.Log("pool size", pool.shard().Len())
	//}
	t.Log("allocs", allocs.Load(), "elapsed", elapsed, "ns/op", (elapsed / time.Duration(totalIterations)).String())
}

var allocs = counter.Counter(0)

var tlsPool = NewPool[bytes.Buffer](Config[bytes.Buffer]{SizeClass: 64, PageSize: 8192, PagesPerShard: 1024,
	AllocFunc: func() unsafe.Pointer {
		allocs.Incr()
		return unsafe.Pointer(bytes.NewBuffer(make([]byte, 64)))
	}})

func BenchmarkTLSPool(b *testing.B) {
	b.SetParallelism(8)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		//shard := pool.shard()
		for pb.Next() {
			tlsPool.Put(tlsPool.Get())
			//shard.Put(shard.Get())
		}
	})
	b.Log("allocs", allocs.Load())
}
