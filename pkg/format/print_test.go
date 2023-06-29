package format

import (
	"runtime/debug"
	"testing"
)

func TestAppendf(t *testing.T) {
	info, _ := debug.ReadBuildInfo()
	Println(info)
	b := make([]byte, 0, 512)
	b = Appendf(b, "hi %s, you are %d years old for %.3f", "Joe", 50, 15.516)
	println(string(b))
}

func BenchmarkAppendf(b *testing.B) {
	b.Run("With String", func(b *testing.B) {
		buf := make([]byte, 0, 512)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = Appendf(buf, "hi %s", "Joe")
			buf = buf[:0]
		}
	})
	b.Run("With Value", func(b *testing.B) {
		buf := make([]byte, 0, 512)
		info, _ := debug.ReadBuildInfo()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = Appendf(buf, "hi %v", info)
			buf = buf[:0]
		}
	})
	b.Run("String/Int", func(b *testing.B) {
		buf := make([]byte, 0, 512)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = Appendf(buf, "hi %s %d", "Joe", 50)
			buf = buf[:0]
		}
	})
	b.Run("With Float", func(b *testing.B) {
		buf := make([]byte, 0, 512)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = Appendf(buf, "hi %.2f", 15.516)
			buf = buf[:0]
		}
	})
}
