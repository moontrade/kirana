package timex

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkNanoTime(b *testing.B) {
	b.Run("nanotime", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NanoTime()
		}
	})
	b.Run("walltime", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			WallTime()
		}
	})
	b.Run("now", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Now()
		}
	})
	b.Run("time.Now().UnixNano()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			time.Now().UnixNano()
		}
	})
}

func TestWallTime(t *testing.T) {
	fmt.Println(WallTime())
	fmt.Println(time.UnixMicro(Now() / 1000).String())
}
