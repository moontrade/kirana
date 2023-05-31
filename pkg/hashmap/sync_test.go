package hashmap

import (
	"github.com/moontrade/kirana/pkg/fastrand"
	"sync"
	"testing"
)

const initsize = 1024

func BenchmarkMap(b *testing.B) {
	b.Run("hashmap.SyncMap", func(b *testing.B) {
		m := NewSyncMap[int, int](0, 1024, HashInt)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.Put(i, i)
			m.Get(i)
			m.Delete(i)
		}
	})

	b.Run("sync.Map", func(b *testing.B) {
		m := new(sync.Map)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.Store(i, i)
			m.Load(i)
			m.Delete(i)
		}
	})
}

func BenchmarkConcurrent(b *testing.B) {
	b.Run("hashmap.SyncMap 30% Delete 70% Store", func(b *testing.B) {
		m := NewSyncMap[int, int](0, 1024, HashInt)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if fastrand.Intn(10) < 3 {
					m.Delete(1)
				} else {
					m.Put(1, 1)
				}
			}
		})
	})

	b.Run("sync.Map 30% Delete 70% Store", func(b *testing.B) {
		m := new(sync.Map)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if fastrand.Intn(10) < 3 {
					m.Delete(1)
				} else {
					m.Store(1, 1)
				}
			}
		})
	})

	b.Run("hashmap.SyncMap 30% Delete 30% Store 40% Get", func(b *testing.B) {
		m := NewSyncMap[int, int](0, 1024, HashInt)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				r := fastrand.Intn(10)
				if r < 3 {
					m.Delete(1)
				} else if r < 6 {
					m.Store(1, 1)
				} else {
					m.Get(1)
				}
			}
		})
	})

	b.Run("sync.Map 30% Delete 30% Store 40% Get", func(b *testing.B) {
		m := new(sync.Map)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				r := fastrand.Intn(10)
				if r < 3 {
					m.Delete(1)
				} else if r < 6 {
					m.Store(1, 1)
				} else {
					m.Load(1)
				}
			}
		})
	})

	b.Run("hashmap.SyncMap 20% Store 80% Get", func(b *testing.B) {
		m := NewSyncMap[int, int](0, 1024, HashInt)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				r := fastrand.Intn(10)
				if r < 2 {
					m.Store(1, 1)
				} else {
					m.Get(1)
				}
			}
		})
	})

	b.Run("sync.Map 20% Store 80% Get", func(b *testing.B) {
		m := new(sync.Map)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				r := fastrand.Intn(10)
				if r < 2 {
					m.Store(1, 1)
				} else {
					m.Load(1)
				}
			}
		})
	})
}

func BenchmarkLoad(b *testing.B) {
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		for i := 0; i < initsize; i++ {
			l.Store(int64(i), nil)
		}
		var ok bool
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, ok = l.Load(int64(fastrand.Uint32n(initsize)))
				if !ok {
					b.Fatal("not found")
				}
			}
		})
	})
	b.Run("hashmap.Map Get", func(b *testing.B) {
		var m = New[int64, struct{}](1024*2, HashInt64)
		for i := 0; i < initsize; i++ {
			m.Set(int64(i), struct{}{})
		}
		var ok bool
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, ok = m.Get(int64(fastrand.Uint32n(initsize)))
				if !ok {
					b.Fatal("not found")
				}
			}
		})
	})
	b.Run("hashmap.SyncMap Get", func(b *testing.B) {
		var m = NewSyncMap[int64, struct{}](512, 1024, HashInt64)
		for i := 0; i < initsize; i++ {
			m.Put(int64(i), struct{}{})
		}
		var ok bool
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, ok = m.Get(int64(fastrand.Uint32n(initsize)))
				if !ok {
					b.Fatal("not found")
				}
			}
		})
	})
	b.Run("hashmap.SyncMap GetOrLoad", func(b *testing.B) {
		var m = NewSyncMap[int64, struct{}](512, 1024, HashInt64)
		for i := 0; i < initsize; i++ {
			m.Put(int64(i), struct{}{})
		}
		var ok bool
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, ok = m.GetOrLoad(int64(fastrand.Uint32n(initsize)))
				if !ok {
					b.Fatal("not found")
				}
			}
		})
	})
	b.Run("hashmap.SyncMap Load", func(b *testing.B) {
		var m = NewSyncMap[int64, struct{}](512, 1024, HashInt64)
		for i := 0; i < initsize; i++ {
			m.Put(int64(i), struct{}{})
		}
		var ok bool
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, ok = m.Load(int64(fastrand.Uint32n(initsize)))
				if !ok {
					b.Fatal("not found")
				}
			}
		})
	})
}
