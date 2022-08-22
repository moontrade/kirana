// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package atomicx_test

import (
	"github.com/moontrade/wormhole/pkg/atomicx"
	"sync/atomic"
	"testing"
)

var sink any

func BenchmarkAtomicLoad64(b *testing.B) {
	b.Run("LoadAcq64", func(b *testing.B) {
		var x uint64
		for i := 0; i < b.N; i++ {
			_ = atomicx.LoadAcq64(&x)
		}
	})
	b.Run("atomicx.Load64", func(b *testing.B) {
		var x uint64
		for i := 0; i < b.N; i++ {
			_ = atomicx.Load64(&x)
		}
	})
	b.Run("atomic.LoadUint64", func(b *testing.B) {
		var x uint64
		for i := 0; i < b.N; i++ {
			_ = atomic.LoadUint64(&x)
		}
	})
	b.Run("atomicx.Store64", func(b *testing.B) {
		var x uint64
		for i := 0; i < b.N; i++ {
			atomicx.Store64(&x, 1)
		}
	})
	b.Run("atomic.StoreUint64", func(b *testing.B) {
		var x uint64
		for i := 0; i < b.N; i++ {
			atomic.StoreUint64(&x, 1)
		}
	})
	b.Run("atomicx.Xchg64", func(b *testing.B) {
		var x uint64
		for i := 0; i < b.N; i++ {
			_ = atomicx.Xchg64(&x, 1)
		}
	})
	b.Run("atomic.SwapUint64", func(b *testing.B) {
		var x uint64
		for i := 0; i < b.N; i++ {
			_ = atomic.SwapUint64(&x, 1)
		}
	})
	b.Run("atomicx.Cas64", func(b *testing.B) {
		var x uint64
		for i := 0; i < b.N; i++ {
			_ = atomicx.Cas64(&x, 0, 1)
		}
	})
	b.Run("atomic.CompareAndSwapUint64", func(b *testing.B) {
		var x uint64
		for i := 0; i < b.N; i++ {
			_ = atomic.CompareAndSwapUint64(&x, 0, 1)
		}
	})
}

func BenchmarkAtomicLoad64Std(b *testing.B) {
	var x int64
	sink = &x
	for i := 0; i < b.N; i++ {
		//_ = sync_atomic.LoadInt64(&x)
		//sync_atomic.StoreUint64(&x, 0)
		//sync_atomic.CompareAndSwapInt64(&x, 0, 1)
		atomic.SwapInt64(&x, 2)
	}
}

func BenchmarkAtomicStore64(b *testing.B) {
	var x uint64
	sink = &x
	for i := 0; i < b.N; i++ {
		atomicx.Store64(&x, 0)
	}
}

func BenchmarkAtomicAdd32(b *testing.B) {
	var x uint32
	sink = &x
	for i := 0; i < b.N; i++ {
		atomicx.Xadd(&x, 1)
		//atomic.Store(&x, 1)
	}
}

func BenchmarkAtomicAdd32Std(b *testing.B) {
	var x uint32
	sink = &x
	for i := 0; i < b.N; i++ {
		//sync_atomic.StoreUint32(&x, 1)
		atomic.AddUint32(&x, 1)
	}
}

func BenchmarkAtomicLoad(b *testing.B) {
	var x uint32
	sink = &x
	for i := 0; i < b.N; i++ {
		_ = atomicx.Load(&x)
	}
}

func BenchmarkAtomicStore(b *testing.B) {
	var x uint32
	sink = &x
	for i := 0; i < b.N; i++ {
		atomicx.Store(&x, 0)
	}
}

func BenchmarkAnd8(b *testing.B) {
	var x [512]uint8 // give byte its own cache line
	sink = &x
	for i := 0; i < b.N; i++ {
		atomicx.And8(&x[255], uint8(i))
	}
}

func BenchmarkAnd(b *testing.B) {
	var x [128]uint32 // give x its own cache line
	sink = &x
	for i := 0; i < b.N; i++ {
		atomicx.And(&x[63], uint32(i))
	}
}

func BenchmarkAnd8Parallel(b *testing.B) {
	var x [512]uint8 // give byte its own cache line
	sink = &x
	b.RunParallel(func(pb *testing.PB) {
		i := uint8(0)
		for pb.Next() {
			atomicx.And8(&x[255], i)
			i++
		}
	})
}

func BenchmarkAndParallel(b *testing.B) {
	var x [128]uint32 // give x its own cache line
	sink = &x
	b.RunParallel(func(pb *testing.PB) {
		i := uint32(0)
		for pb.Next() {
			atomicx.And(&x[63], i)
			i++
		}
	})
}

func BenchmarkOr8(b *testing.B) {
	var x [512]uint8 // give byte its own cache line
	sink = &x
	for i := 0; i < b.N; i++ {
		atomicx.Or8(&x[255], uint8(i))
	}
}

func BenchmarkOr(b *testing.B) {
	var x [128]uint32 // give x its own cache line
	sink = &x
	for i := 0; i < b.N; i++ {
		atomicx.Or(&x[63], uint32(i))
	}
}

func BenchmarkOr8Parallel(b *testing.B) {
	var x [512]uint8 // give byte its own cache line
	sink = &x
	b.RunParallel(func(pb *testing.PB) {
		i := uint8(0)
		for pb.Next() {
			atomicx.Or8(&x[255], i)
			i++
		}
	})
}

func BenchmarkOrParallel(b *testing.B) {
	var x [128]uint32 // give x its own cache line
	sink = &x
	b.RunParallel(func(pb *testing.PB) {
		i := uint32(0)
		for pb.Next() {
			atomicx.Or(&x[63], i)
			i++
		}
	})
}

func BenchmarkXadd(b *testing.B) {
	var x uint32
	ptr := &x
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomicx.Xadd(ptr, 1)
		}
	})
}

func BenchmarkXadd64(b *testing.B) {
	var x uint64
	ptr := &x
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomicx.Xadd64(ptr, 1)
		}
	})
}

func BenchmarkCas(b *testing.B) {
	var x uint32
	x = 1
	ptr := &x
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomicx.Cas(ptr, 1, 0)
			atomicx.Cas(ptr, 0, 1)
		}
	})
}

func BenchmarkCasRel(b *testing.B) {
	var x uint32
	x = 1
	ptr := &x
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomicx.CasRel(ptr, 1, 0)
			atomicx.CasRel(ptr, 0, 1)
		}
	})
}

func BenchmarkCas64(b *testing.B) {
	var x uint64
	x = 1
	ptr := &x
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomicx.Cas64(ptr, 1, 0)
			atomicx.Cas64(ptr, 0, 1)
		}
	})
}
func BenchmarkCas64Std(b *testing.B) {
	var x uint64
	x = 1
	ptr := &x
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.CompareAndSwapUint64(ptr, 1, 0)
			atomic.CompareAndSwapUint64(ptr, 0, 1)
		}
	})
}
func BenchmarkXchg(b *testing.B) {
	var x uint32
	x = 1
	ptr := &x
	b.RunParallel(func(pb *testing.PB) {
		var y uint32
		y = 1
		for pb.Next() {
			y = atomicx.Xchg(ptr, y)
			y += 1
		}
	})
}

func BenchmarkXchg64(b *testing.B) {
	var x uint64
	x = 1
	ptr := &x
	b.RunParallel(func(pb *testing.PB) {
		var y uint64
		y = 1
		for pb.Next() {
			y = atomicx.Xchg64(ptr, y)
			y += 1
		}
	})
}
