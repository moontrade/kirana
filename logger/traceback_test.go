package logger

import (
	"fmt"
	"testing"
)

func doTracebackCaller() {
	f := getCallerFunc()
	fmt.Println(f.String())

	func() {
		f := getCallerFunc()
		fmt.Println(f.String())

		func() {
			f := getCallerFunc()
			fmt.Println(f.String())

			x := func() {
				f := getCallerFunc()
				fmt.Println(f.String())
			}
			x()

			func() {
				f := getCallerFunc()
				fmt.Println(f.String())
			}()
		}()
	}()
}

func TestTraceback(t *testing.T) {
	doTracebackCaller()
}

func BenchmarkCallerFunc(b *testing.B) {
	defer func() {
		println("count", count.Load())
	}()
	b.Run("parallel", func(b *testing.B) {
		getCallerFunc()
		getCallerFunc()

		b.SetParallelism(2048)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				getCallerFunc()
			}
		})
	})
	b.Run("single thread", func(b *testing.B) {
		getCallerFunc()
		getCallerFunc()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			getCallerFunc()
		}
	})
}
