package timex

import (
	"testing"
	"time"
)

func BenchmarkNanoTime(b *testing.B) {
	b.Run("nanotime", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NanoTime()
		}
	})
	b.Run("time.Now().UnixNano()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			time.Now().UnixNano()
		}
	})
}
