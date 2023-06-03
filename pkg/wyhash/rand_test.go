package wyhash

import (
	"github.com/moontrade/kirana/pkg/fastrand"
	"testing"
)

func BenchmarkAtomicRand_Next(b *testing.B) {
	const parallelism = 16
	b.Run("wyhash AtomicRand", func(b *testing.B) {
		r := NewAtomicRand()
		b.SetParallelism(parallelism)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				r.Next()
			}
		})
	})

	b.Run("wyhash AtomicRand Sharded", func(b *testing.B) {
		b.SetParallelism(parallelism)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				Next()
			}
		})
	})

	b.Run("fastrand", func(b *testing.B) {
		b.SetParallelism(parallelism)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				fastrand.Int63()
			}
		})
	})

	//b.Run("go exp/rand", func(b *testing.B) {
	//	b.SetParallelism(parallelism)
	//	b.RunParallel(func(pb *testing.PB) {
	//		for pb.Next() {
	//			rand2.Uint64()
	//		}
	//	})
	//})
	//
	//b.Run("go math/rand", func(b *testing.B) {
	//	b.SetParallelism(parallelism)
	//	b.RunParallel(func(pb *testing.PB) {
	//		for pb.Next() {
	//			rand.Int63()
	//		}
	//	})
	//})
}
