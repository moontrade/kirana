package mpsc

import (
	"github.com/moontrade/kirana/pkg/atomicx"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/pmath"
	"golang.org/x/sys/cpu"
	"reflect"
	"sync/atomic"
	"unsafe"
)

const CacheLinePad = unsafe.Sizeof(cpu.CacheLinePad{})

type node[T any] struct {
	seq  int64
	data unsafe.Pointer
	//_    [CacheLinePad - 16]byte
}

// Bounded implements a circular buffer. It is a fixed activeSize,
// and new writes will be blocked when spawnQ is backlog.
type Bounded[T any] struct {
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

// NewBounded returns the RingBuffer object
func NewBounded[T any](capacity int64, wake chan int64) *Bounded[T] {
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
	return &Bounded[T]{
		head:   0,
		tail:   0,
		mask:   capacity - 1,
		wakeCh: wake,
		nodes:  nodes,
	}
}

func (nr *Bounded[T]) WakeCount() int64 {
	return nr.wakeCount.Load()
}

func (nr *Bounded[T]) WakeChanFullCount() int64 {
	return nr.wakeFull.Load()
}

func (nr *Bounded[T]) Wake() <-chan int64 {
	return nr.wakeCh
}

func (nr *Bounded[T]) Len() int {
	return int(atomic.LoadInt64(&nr.tail) - atomic.LoadInt64(&nr.head))
}

func (nr *Bounded[T]) Cap() int {
	return len(nr.nodes)
}

func (nr *Bounded[T]) IsFull() bool {
	return atomic.LoadInt64(&nr.tail)-atomic.LoadInt64(&nr.head) >= (nr.mask)
}

func (nr *Bounded[T]) IsEmpty() bool {
	return atomic.LoadInt64(&nr.tail)-atomic.LoadInt64(&nr.head) == 0
}

func (nr *Bounded[T]) Processed() int {
	return int(atomic.LoadInt64(&nr.tail))
}

func (nr *Bounded[T]) Push(data *T) bool {
	var (
		cell *node[T]
		pos  = atomic.LoadInt64(&nr.tail)
	)
	for {
		cell = &nr.nodes[pos&nr.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - pos
		if diff == 0 {
			if atomicx.Casint64(&nr.tail, pos, pos+1) {
				//if atomic.CompareAndSwapInt64(&nr.tail, pos, pos+1) {
				break
			}
		} else if diff < 0 {
			return false
		} else {
			pos = atomic.LoadInt64(&nr.tail)
		}
	}

	atomic.StorePointer(&cell.data, unsafe.Pointer(data))
	atomic.StoreInt64(&cell.seq, pos+1)

	if nr.wakeCh != nil && pos-atomic.LoadInt64(&nr.head) == 0 {
		//if nr.wakeCh != nil && atomic.LoadInt64(&nr.wake) == 0 {
		if atomicx.Casint64(&nr.wake, 0, 1) {
			nr.wakeCount.Incr()
			select {
			case nr.wakeCh <- 0:
			default:
				nr.wakeFull.Incr()
			}
		}
	}

	return true
}

func (nr *Bounded[T]) PushUnsafe(data unsafe.Pointer) bool {
	var (
		cell *node[T]
		pos  = atomic.LoadInt64(&nr.tail)
	)
	for {
		cell = &nr.nodes[pos&nr.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - pos
		if diff == 0 {
			if atomicx.Casint64(&nr.tail, pos, pos+1) {
				//if atomic.CompareAndSwapInt64(&nr.tail, pos, pos+1) {
				break
			}
		} else if diff < 0 {
			return false
		} else {
			pos = atomic.LoadInt64(&nr.tail)
		}
	}

	atomic.StorePointer(&cell.data, data)
	atomic.StoreInt64(&cell.seq, pos+1)

	if nr.wakeCh != nil && pos-atomic.LoadInt64(&nr.head) == 0 {
		//if nr.wakeCh != nil && atomic.LoadInt64(&nr.wake) == 0 {
		if atomicx.Casint64(&nr.wake, 0, 1) {
			nr.wakeCount.Incr()
			select {
			case nr.wakeCh <- 0:
			default:
				nr.wakeFull.Incr()
			}
		}
	}

	return true
}

func (nr *Bounded[T]) Pop() *T {
	//nr.wake = 0
	atomic.StoreInt64(&nr.wake, 0)
	var (
		cell   *node[T]
		result *T
		pos    = atomic.LoadInt64(&nr.head)
	)
	for {
		cell = &nr.nodes[pos&nr.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - (pos + 1)
		if diff == 0 {
			if atomicx.Casint64(&nr.head, pos, pos+1) {
				//if atomic.CompareAndSwapInt64(&nr.head, pos, pos+1) {
				break
			}
		} else if diff < 0 {
			return nil
		} else {
			pos = atomic.LoadInt64(&nr.head)
		}
	}

	result = (*T)(atomic.LoadPointer(&cell.data))
	atomic.StorePointer(&cell.data, nil)
	//result = (*T)(atomic.SwapPointer(&cell.data, nil))
	atomic.StoreInt64(&cell.seq, pos+nr.mask+1)
	return result
}

func (nr *Bounded[T]) PopDeref() (res T) {
	//nr.wake = 0
	atomic.StoreInt64(&nr.wake, 0)
	var (
		cell *node[T]
		pos  = nr.head
	)
	for {
		cell = &nr.nodes[pos&nr.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - (pos + 1)
		if diff == 0 {
			if atomicx.Casint64(&nr.head, pos, pos+1) {
				//if atomic.CompareAndSwapInt64(&nr.head, pos, pos+1) {
				break
			}
		} else if diff < 0 {
			return
		} else {
			pos = atomic.LoadInt64(&nr.head)
		}
	}

	result := atomic.LoadPointer(&cell.data)
	atomic.StorePointer(&cell.data, nil)
	//result := atomic.SwapPointer(&cell.data, nil)
	atomic.StoreInt64(&cell.seq, pos+nr.mask+1)
	return *(*T)(unsafe.Pointer(&result))
}

func (nr *Bounded[T]) PopMany(maxCount int, consumer func(*T)) (count int) {
	atomic.StoreInt64(&nr.wake, 0)
	var (
		cell   *node[T]
		result *T
		pos    = atomic.LoadInt64(&nr.head)
	)
	for {
		cell = &nr.nodes[pos&nr.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - (pos + 1)
		if diff == 0 {
			if cell.data == nil {
				atomic.StoreInt64(&nr.head, pos)
				return
			}
			result = (*T)(atomic.LoadPointer(&cell.data))
			//result = (*T)(cell.data)
			atomic.StorePointer(&cell.data, nil)
			cell.data = nil
			//result = (*T)(atomic.SwapPointer(&cell.data, nil))
			atomic.StoreInt64(&cell.seq, pos+nr.mask+1)
			consumer(result)
			count++
			pos++
			nr.head++

			if count >= maxCount {
				atomic.StoreInt64(&nr.head, pos)
				return
			}
			//if atomic.CompareAndSwapInt64(&nr.head, nr.head, nr.head+1) {
			//	result = *(*T)(unsafe.Pointer(&cell.data))
			//	cell.data = nil
			//	atomic.StoreInt64(&cell.seq, nr.head+nr.mask+1)
			//	consumer(result)
			//	count++
			//	//nr.head++
			//
			//	if count >= maxCount {
			//		//atomic.StoreInt64(&nr.head, nr.head)
			//		return
			//	}
			//
			//	goto Start
			//}
		} else if diff < 0 {
			atomic.StoreInt64(&nr.head, pos)
			return
		} else {
			pos = atomic.LoadInt64(&nr.head)
		}
	}
}

func (nr *Bounded[T]) PopManyDeref(maxCount int, consumer func(T)) (count int) {
	atomic.StoreInt64(&nr.wake, 0)
	var (
		cell   *node[T]
		result unsafe.Pointer
		pos    = atomic.LoadInt64(&nr.head)
	)
	for {
		cell = &nr.nodes[pos&nr.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - (pos + 1)
		if diff == 0 {
			result = atomic.LoadPointer(&cell.data)
			//result = atomic.SwapPointer(&cell.data, nil)
			// Is the data there yet?
			if result == nil {
				atomic.StoreInt64(&nr.head, pos)
				return
			}
			atomic.StorePointer(&cell.data, nil)
			atomic.StoreInt64(&cell.seq, nr.head+nr.mask+1)
			consumer(*(*T)(unsafe.Pointer(&result)))
			count++
			pos++
			nr.head++

			if count >= maxCount {
				atomic.StoreInt64(&nr.head, pos)
				return
			}
		} else if diff < 0 {
			atomic.StoreInt64(&nr.head, pos)
			return
		} else {
			pos = atomic.LoadInt64(&nr.head)
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

func (nr *Bounded[T]) Iterate(consumer func(*T) bool) (count int) {
	var (
		next = atomic.LoadInt64(&nr.head)
		mask = nr.mask
	)
	for i := 0; i < len(nr.nodes); i++ {
		slot := &nr.nodes[next&mask]
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

func (nr *Bounded[T]) IterateDesc(consumer func(*T) bool) (count int) {
	var (
		next = atomic.LoadInt64(&nr.tail)
		mask = nr.mask
	)
	for i := 0; i < len(nr.nodes); i++ {
		slot := &nr.nodes[next&mask]
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
