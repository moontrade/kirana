package runtimex

import (
	"testing"
)

func doTracebackCaller() {
	f := getCallerFunc()
	println(f.String())

	func() {
		f := getCallerFunc()
		println(f.String())

		func() {
			f := getCallerFunc()
			println(f.String())

			x := func() {
				f := getCallerFunc()
				println(f.String())
			}
			x()

			func() {
				f := getCallerFunc()
				println(f.String())
			}()
		}()
	}()
}

func doTraceback() {
	println(VisitCaller(func(f *FuncInfo[any]) {
		println(f.String())
	}).String())

	func() {
		println(VisitCaller(func(f *FuncInfo[any]) {
			println(f.String())
		}).String())

		func() {
			println(VisitCaller(func(f *FuncInfo[any]) {
				println(f.String())
			}).String())

			x := func() {
				println(VisitCaller(func(f *FuncInfo[any]) {
					println(f.String())
				}).String())
			}
			x()

			func() {
				println(VisitCaller(func(f *FuncInfo[any]) {
					println(f.String())
				}).String())
			}()
		}()
	}()
}

func TestTraceback(t *testing.T) {
	println(VisitCaller(func(f *FuncInfo[any]) {
		println(f.String())
	}).String())
	doTraceback()
}

func BenchmarkCallerFunc(b *testing.B) {
	b.Run("visit", func(b *testing.B) {
		cb := func(f *FuncInfo[any]) {}
		VisitCaller(cb)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			VisitCaller(cb)
		}
	})
	b.Run("visit parallel", func(b *testing.B) {
		cb := func(f *FuncInfo[any]) {}
		VisitCaller(cb)

		b.SetParallelism(128)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				VisitCaller(cb)
			}
		})
	})
	b.Run("VisitCallerArgs Parallel", func(b *testing.B) {
		args := VisitArgs{}
		cb := func(args *VisitArgs) {}
		VisitCallerArgs(&args, cb)

		b.SetParallelism(128)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				VisitCallerArgs(&args, cb)
			}
		})
	})

	defer func() {
		println("count", count.Load())
	}()
	b.Run("get parallel", func(b *testing.B) {
		getCallerFunc()

		b.SetParallelism(32)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				getCallerFunc()
			}
		})
	})
	b.Run("VisitCallerArgs", func(b *testing.B) {
		args := VisitArgs{}
		cb := func(args *VisitArgs) {}
		VisitCallerArgs(&args, cb)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			VisitCallerArgs(&args, cb)
		}
	})
	b.Run("no callback", func(b *testing.B) {
		getCallerFunc()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			getCallerFunc()
		}
	})
}
