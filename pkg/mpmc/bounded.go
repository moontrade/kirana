package mpmc

import (
	"github.com/moontrade/kirana/pkg/pmath"
	"golang.org/x/sys/cpu"
	"reflect"
	"runtime"
	"sync/atomic"
	"unsafe"
)

const CacheLinePad = unsafe.Sizeof(cpu.CacheLinePad{})

type node[T any] struct {
	seq  int64
	data unsafe.Pointer
	//_    [CacheLinePad - 16]byte
}

// Bounded implements a circular buffer.
type Bounded[T any] struct {
	head  int64
	_     [CacheLinePad - 8]byte
	tail  int64
	_     [CacheLinePad - 8]byte
	nodes []node[T]
	mask  int64
	_     [CacheLinePad - unsafe.Sizeof(reflect.SliceHeader{}) - 8]byte
	left  *Bounded[T]
	right *Bounded[T]
}

// NewBounded returns the RingBuffer object
func NewBounded[T any](capacity int64) *Bounded[T] {
	if capacity < 4 {
		capacity = 4
	}
	capacity = int64(pmath.CeilToPowerOf2(int(capacity)))
	nodes := make([]node[T], capacity)
	for i := 0; i < len(nodes); i++ {
		//n := &nodes[i]
		atomic.StoreInt64(&nodes[i].seq, int64(i))
	}
	return &Bounded[T]{
		head:  0,
		tail:  0,
		mask:  capacity - 1,
		nodes: nodes,
	}
}

func (b *Bounded[T]) Len() int {
	return int(atomic.LoadInt64(&b.tail) - atomic.LoadInt64(&b.head))
}

func (b *Bounded[T]) Cap() int {
	return len(b.nodes)
}

func (b *Bounded[T]) IsFull() bool {
	return atomic.LoadInt64(&b.tail)-atomic.LoadInt64(&b.head) >= (b.mask)
}

func (b *Bounded[T]) IsEmpty() bool {
	return atomic.LoadInt64(&b.tail)-atomic.LoadInt64(&b.head) == 0
}

func (b *Bounded[T]) Processed() int {
	return int(atomic.LoadInt64(&b.tail))
}

func (b *Bounded[T]) Enqueue(data *T) bool {
	if data == nil {
		return false
	}
	var (
		cell *node[T]
		pos  = atomic.LoadInt64(&b.tail)
	)
	for {
		cell = &b.nodes[pos&b.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - pos
		if diff == 0 {
			//if atomicx.Casint64(&b.tail, pos, pos+1) {
			if atomic.CompareAndSwapInt64(&b.tail, pos, pos+1) {
				break
			}
			//pos = atomic.LoadInt64(&b.tail)
		} else if diff < 0 {
			if b.IsFull() {
				return false
			}
			runtime.Gosched()
		} else {
			pos = atomic.LoadInt64(&b.tail)
		}
	}

	atomic.StorePointer(&cell.data, unsafe.Pointer(data))
	atomic.StoreInt64(&cell.seq, pos+1)
	return true
}

func (b *Bounded[T]) EnqueueUnsafe(data unsafe.Pointer) bool {
	if data == nil {
		return false
	}
	var (
		cell *node[T]
		pos  = atomic.LoadInt64(&b.tail)
	)
	for {
		cell = &b.nodes[pos&b.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - pos
		if diff == 0 {
			if atomic.CompareAndSwapInt64(&b.tail, pos, pos+1) {
				break
			}
		} else if diff < 0 {
			if b.IsFull() {
				return false
			}
			runtime.Gosched()
		} else {
			pos = atomic.LoadInt64(&b.tail)
		}
	}
	atomic.StorePointer(&cell.data, data)
	atomic.StoreInt64(&cell.seq, pos+1)
	return true
}

func (b *Bounded[T]) Dequeue() *T {
	var (
		cell *node[T]
		pos  = atomic.LoadInt64(&b.head)
	)
	for {
		cell = &b.nodes[pos&b.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - (pos + 1)
		if diff == 0 {
			if atomic.CompareAndSwapInt64(&b.head, pos, pos+1) {
				break
			}
		} else if diff < 0 {
			if b.IsEmpty() {
				return nil
			}
			runtime.Gosched()
		} else {
			pos = atomic.LoadInt64(&b.head)
		}
	}

	data := atomic.SwapPointer(&cell.data, nil)
	for data == nil {
		runtime.Gosched()
		data = atomic.SwapPointer(&cell.data, nil)
	}
	atomic.StoreInt64(&cell.seq, pos+b.mask+1)
	return (*T)(data)
}

func (b *Bounded[T]) DequeueUnsafe() unsafe.Pointer {
	var (
		cell *node[T]
		pos  = atomic.LoadInt64(&b.head)
	)
	for {
		cell = &b.nodes[pos&b.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - (pos + 1)
		if diff == 0 {
			if atomic.CompareAndSwapInt64(&b.head, pos, pos+1) {
				break
			}
		} else if diff < 0 {
			if b.IsEmpty() {
				return nil
			}
			runtime.Gosched()
		} else {
			pos = atomic.LoadInt64(&b.head)
		}
	}

	data := atomic.SwapPointer(&cell.data, nil)
	for data == nil {
		runtime.Gosched()
		data = atomic.SwapPointer(&cell.data, nil)
	}
	atomic.StoreInt64(&cell.seq, pos+b.mask+1)
	return data
}

func (b *Bounded[T]) DequeueDeref() (res T) {
	var (
		cell *node[T]
		pos  = atomic.LoadInt64(&b.head)
	)
	for {
		cell = &b.nodes[pos&b.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - (pos + 1)
		if diff == 0 {
			if atomic.CompareAndSwapInt64(&b.head, pos, pos+1) {
				break
			}
		} else if diff < 0 {
			if b.IsEmpty() {
				return
			}
			runtime.Gosched()
		} else {
			pos = atomic.LoadInt64(&b.head)
		}
	}

	data := atomic.SwapPointer(&cell.data, nil)
	for data == nil {
		runtime.Gosched()
		data = atomic.SwapPointer(&cell.data, nil)
	}
	atomic.StoreInt64(&cell.seq, pos+b.mask+1)
	return *(*T)(unsafe.Pointer(&data))
}

func (b *Bounded[T]) DequeueMany(maxCount int, consumer func(*T)) (count int) {
	var (
		cell *node[T]
		pos  = atomic.LoadInt64(&b.head)
	)
	for {
		cell = &b.nodes[pos&b.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - (pos + 1)
		if diff == 0 {
			if atomic.CompareAndSwapInt64(&b.head, pos, pos+1) {
				data := atomic.SwapPointer(&cell.data, nil)
				for data == nil {
					runtime.Gosched()
					data = atomic.SwapPointer(&cell.data, nil)
				}

				atomic.StoreInt64(&cell.seq, pos+b.mask+1)
				consumer((*T)(data))
				count++

				if count >= maxCount {
					return
				}
			}
		} else if diff < 0 {
			if b.IsEmpty() {
				return
			}
			runtime.Gosched()
		} else {
			pos = atomic.LoadInt64(&b.head)
		}
	}
}

func (b *Bounded[T]) DequeueManyUnsafe(maxCount int, consumer func(pointer unsafe.Pointer)) (count int) {
	var (
		cell *node[T]
		pos  = atomic.LoadInt64(&b.head)
	)
	for {
		cell = &b.nodes[pos&b.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - (pos + 1)
		if diff == 0 {
			if atomic.CompareAndSwapInt64(&b.head, pos, pos+1) {
				data := atomic.SwapPointer(&cell.data, nil)
				for data == nil {
					runtime.Gosched()
					data = atomic.SwapPointer(&cell.data, nil)
				}

				atomic.StoreInt64(&cell.seq, pos+b.mask+1)
				consumer(data)
				count++

				if count >= maxCount {
					return
				}
			}
		} else if diff < 0 {
			if b.IsEmpty() {
				return
			}
			runtime.Gosched()
		} else {
			pos = atomic.LoadInt64(&b.head)
		}
	}
}

func (b *Bounded[T]) DequeueManyDeref(maxCount int, consumer func(T)) (count int) {
	var (
		cell *node[T]
		pos  = atomic.LoadInt64(&b.head)
	)
	for {
		cell = &b.nodes[pos&b.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - (pos + 1)
		if diff == 0 {
			if atomic.CompareAndSwapInt64(&b.head, pos, pos+1) {
				data := atomic.SwapPointer(&cell.data, nil)
				for data == nil {
					runtime.Gosched()
					data = atomic.SwapPointer(&cell.data, nil)
				}

				atomic.StoreInt64(&cell.seq, pos+b.mask+1)
				consumer(*(*T)(unsafe.Pointer(&data)))
				count++

				if count >= maxCount {
					return
				}
			}
		} else if diff < 0 {
			if b.IsEmpty() {
				return
			}
			runtime.Gosched()
		} else {
			pos = atomic.LoadInt64(&b.head)
		}
	}
}

func Reduce[T any, R any](
	b *Bounded[T],
	value R,
	reducer func(*T, R) R,
) (int, R) {
	var (
		next  = atomic.LoadInt64(&b.head)
		tail  = atomic.LoadInt64(&b.tail)
		size  = tail - next
		mask  = b.mask
		count = 0
	)
	for i := 0; i < len(b.nodes); i++ {
		slot := &b.nodes[next&mask]
		data := atomic.LoadPointer(&slot.data)
		if data != nil {
			count++
			value = reducer((*T)(data), value)
		}
		next++
		size--
		if size <= 0 {
			return count, value
		}
	}
	return count, value
}

func (b *Bounded[T]) Iterate(consumer func(*T) bool) (count int) {
	var (
		next = atomic.LoadInt64(&b.head)
		tail = atomic.LoadInt64(&b.tail)
		size = tail - next
		mask = b.mask
	)
	for i := 0; i < len(b.nodes); i++ {
		slot := &b.nodes[next&mask]
		data := atomic.LoadPointer(&slot.data)
		if data != nil {
			count++
			if !consumer((*T)(data)) {
				return
			}
		}
		next++
		size--
		if size <= 0 {
			return
		}
	}
	return
}

func (b *Bounded[T]) IterateDesc(consumer func(*T) bool) (count int) {
	var (
		next = atomic.LoadInt64(&b.tail)
		head = atomic.LoadInt64(&b.head)
		mask = b.mask
		size = next - head
	)
	for i := 0; i < len(b.nodes); i++ {
		slot := &b.nodes[next&mask]
		data := atomic.LoadPointer(&slot.data)
		if data != nil {
			count++
			if !consumer((*T)(data)) {
				return
			}
		}
		next--
		if next < 0 {
			return
		}
		size--
		if size <= 0 {
			return
		}
	}
	return
}
