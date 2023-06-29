package ringbuf

import (
	"errors"
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/timex"
	"golang.org/x/sys/cpu"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	ErrFull      = errors.New("full")
	ErrEmptyData = errors.New("empty data")
	ErrTooBig    = errors.New("too big")
)

const CacheLinePad = unsafe.Sizeof(cpu.CacheLinePad{})

// Buffer implements a circular buffer.
type Buffer struct {
	ridx int64
	_    [CacheLinePad - 8]byte
	widx int64
	_    [CacheLinePad - 8]byte
	cidx int64
	_    [CacheLinePad - 8]byte
	mask int64
	buf  []byte
	_    [CacheLinePad - unsafe.Sizeof(reflect.SliceHeader{}) - 32]byte
}

// New returns the RingBuffer object
func New(capacity int64) *Buffer {
	if capacity < 256 {
		capacity = 256
	}
	capacity = int64(pmath.CeilToPowerOf2(int(capacity)))
	return &Buffer{
		buf:  make([]byte, capacity),
		mask: capacity - 1,
	}
}

func (b *Buffer) Read(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	var (
		r = atomic.LoadInt64(&b.ridx)
		c = atomic.LoadInt64(&b.cidx)
	)
	if r == c {
		return 0, nil
	}
	n := int(c - r)
	if n > len(buf) {
		n = len(buf)
	}
	// do copy
	index := r & b.mask
	to := b.buf[index:]
	if len(to) < len(buf) {
		copy(buf, to)
		remainder := len(buf) - len(to)
		copy(buf[len(to):], b.buf[0:remainder])
	} else {
		copy(buf, to)
	}
	atomic.StoreInt64(&b.ridx, r+int64(n))
	return n, nil
}

func (b *Buffer) Write(buf []byte) (int, error) {
	// Is it empty?
	if len(buf) == 0 {
		return 0, ErrEmptyData
	}
	// Is it bigger than the entire ring buffer?
	if len(buf) > len(b.buf) {
		return 0, ErrTooBig
	}
	var (
		end = atomic.AddInt64(&b.widx, int64(len(buf)))
		i   = end - int64(len(buf))
		r   = atomic.LoadInt64(&b.ridx)
	)
	// Wait until space is available
	if int64(len(b.buf))-(i-r) < int64(len(buf)) {
		runtime.Gosched()
		r = atomic.LoadInt64(&b.ridx)

		var begin int64 = 0
		for int64(len(b.buf))-(i-r) < int64(len(buf)) {
			runtime.Gosched()
			r = atomic.LoadInt64(&b.ridx)

			if begin == 0 {
				begin = timex.NanoTime()
			} else if (timex.NanoTime() - begin) > int64(time.Second) {
				return 0, ErrFull
			}
		}
	}
	// Copy
	index := i & b.mask
	to := b.buf[index:]
	if len(to) < len(buf) {
		copy(to, buf)
		remainder := len(buf) - len(to)
		copy(b.buf[0:remainder], buf[len(to):])
	} else {
		copy(to, buf)
	}
	// Writes must be ordered by waiting (spinning) for previous write to complete
	c := atomic.LoadInt64(&b.cidx)
	for c != i {
		runtime.Gosched()
		c = atomic.LoadInt64(&b.cidx)
	}
	atomic.StoreInt64(&b.cidx, end)
	return len(buf), nil
}
