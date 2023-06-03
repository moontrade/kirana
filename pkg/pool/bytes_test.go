package pool

import (
	"fmt"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/fastrand"
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/syncx"
	"github.com/moontrade/kirana/pkg/util"
	"reflect"
	"sync"
	"testing"
	"unsafe"
)

func TestBytesPool_AllocRC(t *testing.T) {
	t.Log(pmath.CeilToPowerOf2(4096))
	t.Log(pmath.CeilToPowerOf2(8192))
	t.Log(pmath.PowerOf2Index(8))
	t.Log(pmath.PowerOf2Index(16))
	t.Log(pmath.PowerOf2Index(32))
	t.Log(pmath.PowerOf2Index(64))
	t.Log(pmath.PowerOf2Index(128))
	t.Log(pmath.PowerOf2Index(256))
	t.Log(pmath.PowerOf2Index(512))
	t.Log(pmath.PowerOf2Index(1024))
	t.Log(pmath.PowerOf2Index(2048))
	t.Log(pmath.PowerOf2Index(4096))
	t.Log(pmath.PowerOf2Index(8192))
	t.Log(pmath.PowerOf2Index(16384))
	t.Log(pmath.PowerOf2Index(32768))
	t.Log(pmath.PowerOf2Index(65536))
	t.Log(pmath.PowerOf2Index(65536 * 2))

	//t.Logf("%d", bits.TrailingZeros64(uint64(pmath.CeilToPowerOf2(1))))
}

func printPoolStats(p *Pool[[]byte]) {
	fmt.Println("allocs", p.Allocs.Load())
	fmt.Println("allocs2", p.Allocs2.Load())
	fmt.Println("deallocs", p.Deallocs.Load())
	fmt.Println("pageAllocs", p.PageAllocs.Load())
	fmt.Println("pageAllocAttempts", p.PageAllocAttempts.Load())
	fmt.Println("count", p.Len())
	fmt.Println()
}

func TestAlloc(t *testing.T) {
	var p8 = defaultBytes.s.pools[0]
	pop := p8.shards[0].queue
	_ = pop
	defer func() {
		fmt.Println(pop.Len())
		printPoolStats(p8)
	}()

	printPoolStats(p8)

	const parallelism = 128
	const iterations = 5000000
	const sizeClass = 8

	var wg sync.WaitGroup

	warmup := make([][]byte, 0, 1000)
	for i := 0; i < 1000; i++ {
		warmup = append(warmup, Alloc(sizeClass))
	}
	for i := 0; i < 1000; i++ {
		Free(warmup[0])
	}

	wg.Add(parallelism)
	for i := 0; i < parallelism/2; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				Free(Alloc(sizeClass))
			}
		}()
	}
	for i := 0; i < parallelism-(parallelism/2); i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				a1 := Alloc(sizeClass)
				a2 := Alloc(sizeClass)
				Free(a2)
				Free(a1)
			}
		}()
	}

	wg.Wait()
}

func BenchmarkRandomSizes(b *testing.B) {
	var (
		//min, max    = 36, 8092
		min, max    = 16, 256
		parallelism = 8192
		//maxAllocs = 16384
		//runTLSF     = true
		//showGCStats = false
	)

	const sizes = 1024
	const mask = sizes - 1

	var counter counter.Counter

	randomRangeSizes := make([]int, 0, sizes)
	for i := 0; i < sizes; i++ {
		randomRangeSizes = append(randomRangeSizes, int(randomPowerOf2Range(min, max)))
	}

	b.Run("Sync", func(b *testing.B) {
		pool := NewSyncSlicePool()
		//for i := 0; i < b.N; i++ {
		//	size := randomRangeSizes[i&mask]
		//	pool.Put(pool.Get(size))
		//}
		//clearList := make([][]byte, 0, maxAllocs)

		b.SetParallelism(parallelism)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				i++
				size := randomRangeSizes[i&mask]
				pool.Put(pool.Get(size))
			}
		})
	})

	b.Run("Sync Slab", func(b *testing.B) {
		pool := NewSyncSlicePool()
		//for i := 0; i < b.N; i++ {
		//	size := randomRangeSizes[i&mask]
		//	pool.Put(pool.Get(size))
		//}
		//clearList := make([][]byte, 0, maxAllocs)

		b.SetParallelism(parallelism)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			size := randomRangeSizes[counter.Incr()&mask]
			s := pool.pools[pmath.PowerOf2Index(size)]
			for pb.Next() {
				s.Put(s.Get())
			}
		})
	})

	b.Run("Syncx", func(b *testing.B) {
		pool := NewSyncxSlicePool()
		//for i := 0; i < b.N; i++ {
		//	size := randomRangeSizes[i&mask]
		//	pool.Put(pool.Get(size))
		//}
		//clearList := make([][]byte, 0, maxAllocs)

		b.SetParallelism(parallelism)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				i++
				size := randomRangeSizes[i&mask]
				pool.Put(pool.Get(size))
			}
		})
	})

	b.Run("Syncx Slab", func(b *testing.B) {
		pool := NewSyncSlicePool()
		//for i := 0; i < b.N; i++ {
		//	size := randomRangeSizes[i&mask]
		//	pool.Put(pool.Get(size))
		//}
		//clearList := make([][]byte, 0, maxAllocs)

		b.SetParallelism(parallelism)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			size := randomRangeSizes[counter.Incr()&mask]
			s := pool.pools[pmath.PowerOf2Index(size)]
			for pb.Next() {
				s.Put(s.Get())
			}
		})
	})

	b.Run("Kirana", func(b *testing.B) {
		//for i := 0; i < b.N; i++ {
		//	size := randomRangeSizes[i&mask]
		//	pool.Put(pool.Get(size))
		//}
		//clearList := make([][]byte, 0, maxAllocs)

		b.SetParallelism(parallelism)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				i++
				size := randomRangeSizes[i&mask]
				Free(Alloc(size))
				//pool.Put(pool.Get(size))
			}
		})
	})

	b.Run("Kirana Slab", func(b *testing.B) {
		//for i := 0; i < b.N; i++ {
		//	size := randomRangeSizes[i&mask]
		//	pool.Put(pool.Get(size))
		//}
		//clearList := make([][]byte, 0, maxAllocs)

		b.SetParallelism(parallelism)
		b.ReportAllocs()
		b.ResetTimer()

		//s := &defaultBytes.s.PoolOf(128).shards[0]

		b.RunParallel(func(pb *testing.PB) {
			size := randomRangeSizes[counter.Incr()&mask]
			s := defaultBytes.s.PoolOf(size).Shard()

			//s := defaultBytes.s.PoolOf(size).shards[0]
			//s := defaultBytes.s.pools[pmath.PowerOf2Index(size)].Shard()
			for pb.Next() {
				s.PutUnsafe(s.GetUnsafe())
				//pool.Put(pool.Get(size))
			}
		})

		//fmt.Println("allocs", s.Allocates.Load(), "deallocs", s.Deallocates.Load(), "pageAllocs", s.ShardStats)
	})
}

func BenchmarkBytesPool_Get(b *testing.B) {
	var (
		//min, max    = 36, 8092
		min, max  = 8, 256
		maxAllocs = 16
		//runTLSF     = true
		//showGCStats = false
	)

	const sizes = 1024
	const mask = sizes - 1
	const parallelism = 16
	const (
		runSyncStd     = true
		runSyncx       = true
		runKirana      = true
		runRandomSizes = true
		runOneSize     = true
	)

	randomRangeSizes := make([]int, 0, sizes)
	for i := 0; i < sizes; i++ {
		randomRangeSizes = append(randomRangeSizes, int(randomPowerOf2Range(min, max)))
	}

	defer func() {
		var p = defaultBytes
		var p8 = p.s.pools[0]
		_ = p8
		fmt.Println("allocs", p8.Allocs.Load())
		fmt.Println("deallocs", p8.Deallocs.Load())
		fmt.Println("pageAllocs", p8.PageAllocs.Load())
		fmt.Println("pageAllocAttempts", p8.PageAllocAttempts.Load())
		fmt.Println("count", p8.Len())
		fmt.Println("ALL DONE!!!")
	}()

	if runSyncStd && runRandomSizes {
		b.Run("sync.Pool Random Sizes Parallel", func(b *testing.B) {
			pool := NewSyncSlicePool()
			//for i := 0; i < b.N; i++ {
			//	size := randomRangeSizes[i&mask]
			//	pool.Put(pool.Get(size))
			//}

			b.ReportAllocs()
			b.ResetTimer()
			b.SetParallelism(parallelism)

			b.RunParallel(func(pb *testing.PB) {
				clearList := make([][]byte, 0, maxAllocs)

				i := 0
				for pb.Next() {
					//size := randomRangeSizes[counter.Incr()&mask]
					size := randomRangeSizes[i&mask]
					i++
					clearList = append(clearList, pool.Get(size))
					if len(clearList) >= maxAllocs {
						for _, el := range clearList {
							pool.Put(el)
						}
						clearList = clearList[:0]
					}
				}
			})
		})
	}

	if runSyncx && runRandomSizes {
		b.Run("syncx.Pool Random Sizes Parallel", func(b *testing.B) {
			pool := NewSyncxSlicePool()
			//for i := 0; i < b.N; i++ {
			//	size := randomRangeSizes[i&mask]
			//	pool.Put(pool.Get(size))
			//}

			b.ReportAllocs()
			b.ResetTimer()
			b.SetParallelism(parallelism)

			b.RunParallel(func(pb *testing.PB) {
				clearList := make([][]byte, 0, maxAllocs)

				i := 0
				for pb.Next() {
					//size := randomRangeSizes[counter.Incr()&mask]
					size := randomRangeSizes[i&mask]
					i++
					clearList = append(clearList, pool.Get(size))
					if len(clearList) >= maxAllocs {
						for _, el := range clearList {
							pool.Put(el)
						}
						clearList = clearList[:0]
					}
				}
			})
		})
	}

	if runKirana && runRandomSizes {
		b.Run("Kirana Pool Random Sizes Parallel", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			b.SetParallelism(parallelism)

			b.RunParallel(func(pb *testing.PB) {
				clearList := make([][]byte, 0, maxAllocs)

				i := 0
				for pb.Next() {
					size := randomRangeSizes[i&mask]
					i++
					//size := randomRangeSizes[counter.Incr()&mask]
					clearList = append(clearList, Alloc(size))
					if len(clearList) >= maxAllocs {
						for _, el := range clearList {
							Free(el)
						}
						clearList = clearList[:0]
					}
				}
			})
		})
	}

	if runSyncStd && runOneSize {
		b.Run("sync.Pool One Size Parallel", func(b *testing.B) {
			pool := NewSyncSlicePool()
			//for i := 0; i < b.N; i++ {
			//	size := randomRangeSizes[i&mask]
			//	pool.Put(pool.Get(size))
			//}

			size := randomRangeSizes[0]
			b.ReportAllocs()
			b.ResetTimer()
			b.SetParallelism(parallelism)

			b.RunParallel(func(pb *testing.PB) {
				clearList := make([][]byte, 0, maxAllocs)

				for pb.Next() {
					clearList = append(clearList, pool.Get(size))
					if len(clearList) >= maxAllocs {
						for _, el := range clearList {
							pool.Put(el)
						}
						clearList = clearList[:0]
					}
				}
			})
		})
	}

	if runSyncx && runOneSize {
		b.Run("syncx.Pool One Size Parallel", func(b *testing.B) {
			pool := NewSyncxSlicePool()
			//for i := 0; i < b.N; i++ {
			//	size := randomRangeSizes[i&mask]
			//	pool.Put(pool.Get(size))
			//}
			size := randomRangeSizes[0]

			b.ReportAllocs()
			b.ResetTimer()
			b.SetParallelism(parallelism)

			b.RunParallel(func(pb *testing.PB) {
				clearList := make([][]byte, 0, maxAllocs)

				for pb.Next() {
					clearList = append(clearList, pool.Get(size))
					if len(clearList) >= maxAllocs {
						for _, el := range clearList {
							pool.Put(el)
						}
						clearList = clearList[:0]
					}
				}
			})
		})
	}

	if runKirana && runOneSize {
		b.Run("Kirana Pool One Size Parallel", func(b *testing.B) {
			//size := randomRangeSizes[0]
			size := 8

			b.ReportAllocs()
			b.ResetTimer()
			b.SetParallelism(parallelism)

			b.RunParallel(func(pb *testing.PB) {
				clearList := make([]unsafe.Pointer, 0, maxAllocs)
				for pb.Next() {
					clearList = append(clearList, bytesToPtr(Alloc(size)))
					if len(clearList) >= maxAllocs {
						for _, el := range clearList {
							Free(ptrToBytes(el, size))
						}
						clearList = clearList[:0]
					}
				}
			})
		})
	}

	//b.Run("sync.Pool Slab", func(b *testing.B) {
	//	size := randomRangeSizes[12&mask]
	//	pool := NewSyncSlicePool()
	//	p := pool.pools[pmath.PowerOf2Index(size)]
	//	p.Put(p.Get())
	//	clearList := make([][]byte, 0, 128)
	//	var v []byte
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		v = p.Get().([]byte)
	//		clearList = append(clearList, v)
	//		if len(clearList) == cap(clearList) {
	//			for _, el := range clearList {
	//				p.Put(el)
	//			}
	//			clearList = clearList[:0]
	//		}
	//		//p.Put(p.Get())
	//	}
	//})
	//b.Run("syncx.Pool Slab", func(b *testing.B) {
	//	size := randomRangeSizes[12&mask]
	//	pool := NewSyncxSlicePool()
	//	p := pool.pools[pmath.PowerOf2Index(size)]
	//	p.Put(p.Get())
	//	clearList := make([][]byte, 0, 128)
	//	var v []byte
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		v = p.Get().([]byte)
	//		clearList = append(clearList, v)
	//		if len(clearList) == cap(clearList) {
	//			for _, el := range clearList {
	//				p.Put(el)
	//			}
	//			clearList = clearList[:0]
	//		}
	//		//p.Put(p.Get())
	//	}
	//})
	//b.Run("Kirana Pool Slab", func(b *testing.B) {
	//	size := randomRangeSizes[12&mask]
	//	p := defaultBytes.s.Slab(size)
	//	p.Put(p.Get())
	//	clearList := make([][]byte, 0, 128)
	//	var v []byte
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		v = p.Get()
	//		clearList = append(clearList, v)
	//		if len(clearList) == cap(clearList) {
	//			for _, el := range clearList {
	//				p.Put(el)
	//			}
	//			clearList = clearList[:0]
	//		}
	//	}
	//})
	//b.Run("TLSF", func(b *testing.B) {
	//	heap := tlsf.NewHeap(4)
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		heap.Free(heap.Alloc(uintptr(randomRangeSizes[i&mask])))
	//	}
	//})
	//b.Run("rpmalloc Heap", func(b *testing.B) {
	//	heap := rpmalloc.AcquireHeap()
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		heap.Free(heap.Alloc(uintptr(randomRangeSizes[i&mask])))
	//	}
	//	b.StopTimer()
	//	heap.FreeAll()
	//})
	//b.Run("rpmalloc", func(b *testing.B) {
	//	heap := rpmalloc.AcquireHeap()
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		memory.Free(memory.Alloc(uintptr(randomRangeSizes[i&mask])))
	//	}
	//	b.StopTimer()
	//	heap.FreeAll()
	//})
}

func bytesToPtr(b []byte) unsafe.Pointer {
	return unsafe.Pointer(&b[0])
}

func ptrToBytes(p unsafe.Pointer, size int) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(p),
		Len:  size,
		Cap:  size,
	}))
}

func randomPowerOf2Range(min, max int) int {
	first := pmath.PowerOf2Index(min)
	last := pmath.PowerOf2Index(max) + 1

	var ret uint64
	bit := fastrand.Intn(last-first) + first
	ret = util.SetBit64(ret, uint64(bit))
	return int(ret)
}

type SyncxSlicePool struct {
	pools [64]*syncx.Pool
}

func NewSyncxSlicePool() *SyncxSlicePool {
	pools := make([]*syncx.Pool, 65)
	newPool := func(size int) *syncx.Pool {
		return &syncx.Pool{
			NoGC: true,
			New: func() interface{} {
				return make([]byte, size)
			},
		}
	}
	pools[0] = newPool(8)
	pools[1] = pools[0]
	pools[2] = pools[0]
	pools[3] = pools[0]
	pools[4] = newPool(16)
	pools[5] = newPool(32)
	pools[6] = newPool(64)
	pools[7] = newPool(128)
	pools[8] = newPool(256)
	pools[9] = newPool(512)
	pools[10] = newPool(1024)
	pools[11] = newPool(1024 * 2)
	pools[12] = newPool(1024 * 4)
	pools[13] = newPool(1024 * 8)
	pools[14] = newPool(1024 * 16)
	pools[15] = newPool(1024 * 32)
	pools[16] = newPool(1024 * 64)
	pools[17] = newPool(1024 * 128)
	pools[18] = newPool(1024 * 256)
	r := &SyncxSlicePool{}
	copy(r.pools[0:cap(r.pools)], pools)
	return r
}

func (s *SyncxSlicePool) Get(size int) []byte {
	p := s.pools[pmath.PowerOf2Index(size)]
	if p == nil {
		return make([]byte, size)
	}
	return p.Get().([]byte)
}

func (s *SyncxSlicePool) Put(b []byte) {
	p := s.pools[pmath.PowerOf2Index(cap(b))]
	if p == nil {
		return
	}
	p.Put(b)
}

type SyncSlicePool struct {
	pools [65]*sync.Pool
}

func NewSyncSlicePool() *SyncSlicePool {
	pools := make([]*sync.Pool, 65)
	newPool := func(size int) *sync.Pool {
		return &sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		}
	}
	pools[0] = newPool(8)
	pools[1] = pools[0]
	pools[2] = pools[0]
	pools[3] = pools[0]
	pools[4] = newPool(16)
	pools[5] = newPool(32)
	pools[6] = newPool(64)
	pools[7] = newPool(128)
	pools[8] = newPool(256)
	pools[9] = newPool(512)
	pools[10] = newPool(1024)
	pools[11] = newPool(1024 * 2)
	pools[12] = newPool(1024 * 4)
	pools[13] = newPool(1024 * 8)
	pools[14] = newPool(1024 * 16)
	pools[15] = newPool(1024 * 32)
	pools[16] = newPool(1024 * 64)
	pools[17] = newPool(1024 * 128)
	pools[18] = newPool(1024 * 256)
	r := &SyncSlicePool{}
	copy(r.pools[0:cap(r.pools)], pools)
	return r
}

func (s *SyncSlicePool) Get(size int) []byte {
	p := s.pools[pmath.PowerOf2Index(size)]
	if p == nil {
		return make([]byte, size)
	}
	return p.Get().([]byte)
}

func (s *SyncSlicePool) Put(b []byte) {
	p := s.pools[pmath.PowerOf2Index(cap(b))]
	if p == nil {
		return
	}
	p.Put(b)
}
