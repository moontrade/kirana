package ringbuf

import (
	"runtime"
	"strings"
	"testing"
)

func Benchmark_AsyncRead(b *testing.B) {
	rb := New(1024 * 1024)
	data := []byte(strings.Repeat("a", 256))
	buf := make([]byte, 1024*1024)

	go func() {
		n := 0
		for {
			n, _ = rb.Read(buf)
			if n == 0 {
				runtime.Gosched()
			}
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Write(data)
	}
}

func Benchmark_AsyncWrite(b *testing.B) {
	rb := New(1024 * 1024)
	data := []byte(strings.Repeat("a", 256))
	buf := make([]byte, len(rb.buf))

	go func() {
		for {
			rb.Write(data)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Read(buf)
	}
}
