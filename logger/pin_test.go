package logger

import (
	"runtime"
	"sync"
	"testing"
)

func TestPin(t *testing.T) {
	n := procPin()
	procUnpin()
	t.Log(n)
	var wg sync.WaitGroup
	for i := 0; i < runtime.GOMAXPROCS(0)*2; i++ {
		wg.Add(1)
		go func() {
			n := procPin()
			procUnpin()
			t.Log(n)
			wg.Done()
		}()
		runtime.Gosched()
	}
	wg.Wait()
}

func BenchmarkPin(b *testing.B) {
	b.Run("single thread", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			procUnpinGp(procPin())
		}
	})

	b.Run("parallel", func(b *testing.B) {
		b.SetParallelism(32)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				gp := procPin()
				_ = gp
				systemstack(func() {
				})
				procUnpin()
			}
		})
	})
}
