package logger

import (
	"github.com/moontrade/unsafe/memory"
	"reflect"
	"strings"
	"testing"
	"unsafe"
)

func zero(b []byte) []byte {
	memory.Zero(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&b)).Data), uintptr(cap(b)))
	return b
}

func TestAppenderShard(t *testing.T) {
	//waker := make(chan int, 1)
	waker := newWriterQueue(16)
	rb := newWriterShard(0, 512, waker)
	data := []byte(strings.Repeat("a", 259))
	buf := make([]byte, len(rb.buf))
	data[len(data)-1] = 'z'
	if _, err := rb.write(data); err != nil {
		t.Fatal(err)
	}
	n, err := rb.read(buf)
	if err != nil {
		t.Fatal(err)
	}
	buf = buf[:n]
	if buf[len(data)-1] != 'z' {
		t.Fatal("expected 'z'")
	}
	buf = buf[:]
	zero(buf)
	if _, err := rb.write(data); err != nil {
		t.Fatal(err)
	}
	n, err = rb.read(buf)
	if err != nil {
		t.Fatal(err)
	}
	buf = buf[:n]
	if buf[len(data)-1] != 'z' {
		t.Fatal("expected 'z'")
	}
}

func BenchmarkWriterShard_AsyncRead(b *testing.B) {
	wr := newWriter(1, 1024*256)
	rb := wr.shards[0]
	data := []byte(strings.Repeat("a", 128))
	//buf := make([]byte, len(rb.buf))
	//
	//go func() {
	//	n := 0
	//	for {
	//		n, _ = rb.read(buf)
	//		if n == 0 {
	//			runtime.Gosched()
	//		}
	//	}
	//}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := rb.write(data); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.Log("writes", b.N, "wakes", rb.wakes, "spins", rb.spins)
}

func BenchmarkWriterShard_AsyncWrite(b *testing.B) {
	//waker := make(chan int, 1)
	waker := newWriterQueue(16)
	_ = waker
	rb := newWriterShard(0, 1024*1024, waker)
	data := []byte(strings.Repeat("a", 256))
	buf := make([]byte, len(rb.buf))

	go func() {
		for i := 0; i < b.N; i++ {
			rb.write(data)
		}
	}()

	size := 0
	n := 0
	var err error
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, err = rb.read(buf)
		//if n == 0 {
		//	runtime.Gosched()
		//}
		if err != nil {
			b.Fatal(err)
		}
		size += n
	}
	b.StopTimer()
	b.Log("N", b.N, "size", size)
}
