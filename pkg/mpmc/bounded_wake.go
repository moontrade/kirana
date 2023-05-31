package mpmc

import (
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/timex"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

// BoundedWake implements a bounded MPMC queue that supports automatic goroutine
// wake mechanism when a new message is enqueued that changes the length from 0 to 1.
type BoundedWake[T any] struct {
	head          int64
	_             [CacheLinePad - 8]byte
	tail          int64
	_             [CacheLinePad - 8]byte
	nodes         []node[T]
	mask          int64
	_             [CacheLinePad - unsafe.Sizeof(reflect.SliceHeader{}) - 8]byte
	wake          int64
	_             [CacheLinePad - 8]byte
	wakeCh        chan int64
	wakeCount     counter.Counter
	wakeFull      counter.Counter
	overflowCount counter.Counter
}

// NewBoundedWake returns the RingBuffer object
func NewBoundedWake[T any](capacity int64, wake chan int64) *BoundedWake[T] {
	if wake == nil {
		wake = make(chan int64, 1)
	}
	if capacity <= 32 {
		capacity = 32
	}
	capacity = int64(pmath.CeilToPowerOf2(int(capacity)))
	nodes := make([]node[T], capacity)
	for i := 0; i < len(nodes); i++ {
		n := &nodes[i]
		atomic.StoreInt64(&n.seq, int64(i))
	}
	return &BoundedWake[T]{
		head:   0,
		tail:   0,
		mask:   capacity - 1,
		wakeCh: wake,
		nodes:  nodes,
	}
}

func (b *BoundedWake[T]) WakeCount() int64 {
	return b.wakeCount.Load()
}

func (b *BoundedWake[T]) WakeChanFullCount() int64 {
	return b.wakeFull.Load()
}

func (b *BoundedWake[T]) Wake() <-chan int64 {
	return b.wakeCh
}

func (b *BoundedWake[T]) Len() int {
	return int(atomic.LoadInt64(&b.tail) - atomic.LoadInt64(&b.head))
}

func (b *BoundedWake[T]) Cap() int {
	return len(b.nodes)
}

func (b *BoundedWake[T]) IsFull() bool {
	return atomic.LoadInt64(&b.tail)-atomic.LoadInt64(&b.head) >= (b.mask)
}

func (b *BoundedWake[T]) IsEmpty() bool {
	return atomic.LoadInt64(&b.tail)-atomic.LoadInt64(&b.head) == 0
}

func (b *BoundedWake[T]) Processed() int {
	return int(atomic.LoadInt64(&b.tail))
}

func (b *BoundedWake[T]) Enqueue(data *T) bool {
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
			runtime.Gosched()
			if b.IsFull() {
				return false
			}
		} else {
			pos = atomic.LoadInt64(&b.tail)
		}
	}

	atomic.StorePointer(&cell.data, unsafe.Pointer(data))
	atomic.StoreInt64(&cell.seq, pos+1)

	if b.wakeCh != nil && pos-atomic.LoadInt64(&b.head) == 0 {
		//if b.wakeCh != nil && atomic.LoadInt64(&b.wake) == 0 {
		if atomic.CompareAndSwapInt64(&b.wake, 0, 1) {
			b.wakeCount.Incr()
			select {
			case b.wakeCh <- 0:
			default:
				b.wakeFull.Incr()
			}
		}
	}

	return true
}

func (b *BoundedWake[T]) EnqueueUnsafeTimeout(data unsafe.Pointer, timeout time.Duration) bool {
	if b.EnqueueUnsafe(data) {
		return true
	}
	var (
		begin = timex.NanoTime()
		count = 0
	)
	runtime.Gosched()
	for {
		if b.EnqueueUnsafe(data) {
			return true
		}
		if timex.NanoTime()-begin >= int64(timeout) {
			return false
		}
		count++
		if count%10 == 0 {
			time.Sleep(time.Millisecond)
		} else {
			runtime.Gosched()
		}
	}
}

func (b *BoundedWake[T]) EnqueueUnsafe(data unsafe.Pointer) bool {
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
			runtime.Gosched()
			if b.IsFull() {
				return false
			}
		} else {
			pos = atomic.LoadInt64(&b.tail)
		}
	}

	atomic.StorePointer(&cell.data, data)
	atomic.StoreInt64(&cell.seq, pos+1)

	if b.wakeCh != nil && pos-atomic.LoadInt64(&b.head) == 0 {
		//if b.wakeCh != nil && atomic.LoadInt64(&b.wake) == 0 {
		if atomic.CompareAndSwapInt64(&b.wake, 0, 1) {
			b.wakeCount.Incr()
			select {
			case b.wakeCh <- 0:
			default:
				b.wakeFull.Incr()
			}
		}
	}

	return true
}

func (b *BoundedWake[T]) Dequeue() *T {
	//b.wake = 0
	atomic.StoreInt64(&b.wake, 0)
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
			runtime.Gosched()
			if b.IsEmpty() {
				return nil
			}
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

func (b *BoundedWake[T]) DequeueUnsafe() unsafe.Pointer {
	//b.wake = 0
	atomic.StoreInt64(&b.wake, 0)
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
			runtime.Gosched()
			if b.IsEmpty() {
				return nil
			}
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

func (b *BoundedWake[T]) DequeueDeref() (res T) {
	//b.wake = 0
	atomic.StoreInt64(&b.wake, 0)
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
			runtime.Gosched()
			if b.IsEmpty() {
				return
			}
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

func (b *BoundedWake[T]) DequeueMany(maxCount int, consumer func(*T)) (count int) {
	atomic.StoreInt64(&b.wake, 0)
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

func (b *BoundedWake[T]) DequeueManyUnsafe(maxCount int, consumer func(pointer unsafe.Pointer)) (count int) {
	atomic.StoreInt64(&b.wake, 0)
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

func (b *BoundedWake[T]) DequeueManyDeref(maxCount int, consumer func(T)) (count int) {
	atomic.StoreInt64(&b.wake, 0)
	var (
		cell   *node[T]
		result unsafe.Pointer
		pos    = atomic.LoadInt64(&b.head)
	)
	for {
		cell = &b.nodes[pos&b.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - (pos + 1)
		if diff == 0 {
			result = atomic.LoadPointer(&cell.data)
			//result = atomic.SwapPointer(&cell.data, nil)
			// Is the data there yet?
			if result == nil {
				atomic.StoreInt64(&b.head, pos)
				return
			}
			atomic.StorePointer(&cell.data, nil)
			atomic.StoreInt64(&cell.seq, b.head+b.mask+1)
			consumer(*(*T)(unsafe.Pointer(&result)))
			count++
			pos++
			b.head++

			if count >= maxCount {
				atomic.StoreInt64(&b.head, pos)
				return
			}
		} else if diff < 0 {
			atomic.StoreInt64(&b.head, pos)
			return
		} else {
			pos = atomic.LoadInt64(&b.head)
		}
	}
}

func ReduceWake[T any, R any](
	b *BoundedWake[T],
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

func (b *BoundedWake[T]) Iterate(consumer func(*T) bool) (count int) {
	var (
		next = atomic.LoadInt64(&b.head)
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
	}
	return
}

func (b *BoundedWake[T]) IterateDesc(consumer func(*T) bool) (count int) {
	var (
		next = atomic.LoadInt64(&b.tail)
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
		next--
		if next < 0 {
			return
		}
	}
	return
}
