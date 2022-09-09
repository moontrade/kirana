package pool

import (
	"github.com/moontrade/kirana/pkg/fastrand"
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/syncx"
	"github.com/moontrade/kirana/pkg/util"
	"sync"
	"testing"
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

func BenchmarkBytesPool_Get(b *testing.B) {
	var (
		//min, max    = 36, 8092
		min, max  = 16, 128
		maxAllocs = 16384
		//runTLSF     = true
		//showGCStats = false
	)

	const sizes = 1024
	const mask = sizes - 1

	randomRangeSizes := make([]int, 0, sizes)
	for i := 0; i < sizes; i++ {
		randomRangeSizes = append(randomRangeSizes, int(randomPowerOf2Range(min, max)))
	}

	//b.Run("FindPool", func(b *testing.B) {
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		pool.PoolOf(randomRangeSizes[i&mask])
	//	}
	//})
	//b.Run("FindPoolSwitch", func(b *testing.B) {
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		pool.poolOfSlow(randomRangeSizes[i&mask])
	//	}
	//})
	//
	b.Run("sync.Pool Random Sizes", func(b *testing.B) {
		pool := NewSyncSlicePool()
		//for i := 0; i < b.N; i++ {
		//	size := randomRangeSizes[i&mask]
		//	pool.Put(pool.Get(size))
		//}
		clearList := make([][]byte, 0, maxAllocs)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			size := randomRangeSizes[i&mask]
			clearList = append(clearList, pool.Get(size))
			if len(clearList) == cap(clearList) {
				for _, el := range clearList {
					pool.Put(el)
				}
				clearList = clearList[:0]
			}
		}
	})
	b.Run("syncx.Pool Random Sizes", func(b *testing.B) {
		pool := NewSyncxSlicePool()
		//for i := 0; i < b.N; i++ {
		//	size := randomRangeSizes[i&mask]
		//	pool.Put(pool.Get(size))
		//}
		clearList := make([][]byte, 0, maxAllocs)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			size := randomRangeSizes[i&mask]
			clearList = append(clearList, pool.Get(size))
			if len(clearList) == cap(clearList) {
				for _, el := range clearList {
					pool.Put(el)
				}
				clearList = clearList[:0]
			}
		}
	})
	b.Run("Kirana Pool Random Sizes", func(b *testing.B) {
		//for _, pool := range defaultBytes.pools {
		//	if pool == nil {
		//		continue
		//	}
		//	for i := 0; i < len(pool.shards); i++ {
		//		shard := &pool.shards[i]
		//		for i := 0; i < maxAllocs; i++ {
		//			shard.PutUnsafe(shard.GetUnsafe())
		//		}
		//	}
		//}
		//for i := 0; i < b.N; i++ {
		//	size := randomRangeSizes[i&mask]
		//	Free(Alloc(size))
		//}
		clearList := make([][]byte, 0, maxAllocs)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			size := randomRangeSizes[i&mask]
			clearList = append(clearList, Alloc(size))
			if len(clearList) == cap(clearList) {
				for _, el := range clearList {
					Free(el)
				}
				clearList = clearList[:0]
			}
		}
		b.StopTimer()
		for _, pool := range defaultBytes.s.pools {
			if pool == nil {
				continue
			}
			if pool.Stats.Allocs.Load() == 0 {
				continue
			}
			b.Log("N", b.N, "Size Class", pool.config.SizeClass, " PageAllocs", pool.PageAllocs, " PageDeallocs", pool.PageDeallocs, " Allocs", pool.Allocs, " Deallocs", pool.Deallocs)
		}
	})

	b.Run("sync.Pool Slab", func(b *testing.B) {
		size := randomRangeSizes[12&mask]
		pool := NewSyncSlicePool()
		p := pool.pools[pmath.PowerOf2Index(size)]
		p.Put(p.Get())
		clearList := make([][]byte, 0, 128)
		var v []byte
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			v = p.Get().([]byte)
			clearList = append(clearList, v)
			if len(clearList) == cap(clearList) {
				for _, el := range clearList {
					p.Put(el)
				}
				clearList = clearList[:0]
			}
			//p.Put(p.Get())
		}
	})
	b.Run("syncx.Pool Slab", func(b *testing.B) {
		size := randomRangeSizes[12&mask]
		pool := NewSyncxSlicePool()
		p := pool.pools[pmath.PowerOf2Index(size)]
		p.Put(p.Get())
		clearList := make([][]byte, 0, 128)
		var v []byte
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			v = p.Get().([]byte)
			clearList = append(clearList, v)
			if len(clearList) == cap(clearList) {
				for _, el := range clearList {
					p.Put(el)
				}
				clearList = clearList[:0]
			}
			//p.Put(p.Get())
		}
	})
	b.Run("Kirana Pool Slab", func(b *testing.B) {
		size := randomRangeSizes[12&mask]
		p := defaultBytes.s.Slab(size)
		p.Put(p.Get())
		clearList := make([][]byte, 0, 128)
		var v []byte
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			v = p.Get()
			clearList = append(clearList, v)
			if len(clearList) == cap(clearList) {
				for _, el := range clearList {
					p.Put(el)
				}
				clearList = clearList[:0]
			}
		}
	})
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
	pools[11] = newPool(2048)
	pools[12] = newPool(4096)
	pools[13] = newPool(8192)
	pools[14] = newPool(16384)
	pools[15] = newPool(32768)
	pools[16] = newPool(65536)
	pools[17] = newPool(65536 * 2)
	pools[18] = newPool(65536 * 4)
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
	pools [64]*sync.Pool
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
	pools[11] = newPool(2048)
	pools[12] = newPool(4096)
	pools[13] = newPool(8192)
	pools[14] = newPool(16384)
	pools[15] = newPool(32768)
	pools[16] = newPool(65536)
	pools[17] = newPool(65536 * 2)
	pools[18] = newPool(65536 * 4)
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
