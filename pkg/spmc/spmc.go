package spmc

import (
	"github.com/moontrade/kirana/pkg/mpsc"
	"sync/atomic"
	"unsafe"
)

// SPMC implements a circular buffer. It is a fixed activeSize,
// and new writes will be blocked when spawnQ is full.
type SPMC[T any] struct {
	capacity uint64
	mask     uint64
	_        [mpsc.CacheLinePad - 16]byte
	head     uint64
	_        [mpsc.CacheLinePad - 8]byte
	tail     uint64
	_        [mpsc.CacheLinePad - 8]byte
	wake     int64
	notify   chan int
	null     T
	element  []T
}

// New returns the RingBuffer object
func NewSPMC[T any](capacity uint64, notify chan int) *SPMC[T] {
	if notify == nil {
		notify = make(chan int, 1)
	}
	return &SPMC[T]{
		head:     uint64(0),
		tail:     uint64(0),
		capacity: capacity,
		mask:     capacity - 1,
		notify:   notify,
		element:  make([]T, capacity),
	}
}

func (r *SPMC[T]) Offer(value T) (success bool) {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := atomic.LoadUint64(&r.head)
	if r.isFull(oldTail, oldHead) {
		return false
	}

	isEmpty := r.isEmpty(oldTail, oldHead)

	newTail := oldTail + 1
	tailNode := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newTail&r.mask])))
	// not published yet
	if tailNode != nil {
		return false
	}
	if !atomic.CompareAndSwapUint64(&r.tail, oldTail, newTail) {
		return false
	}

	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newTail&r.mask])), unsafe.Pointer(&value))
	if isEmpty && r.notify != nil {
		select {
		case r.notify <- 1:
		default:
		}
	}
	return true
}

func (r *SPMC[T]) OfferSupplier(valueSupplier func() (v T, finish bool)) {
	oldTail := r.tail
	oldHead := atomic.LoadUint64(&r.head)
	if r.isFull(oldTail, oldHead) {
		return
	}
	isEmpty := r.isEmpty(oldTail, oldHead)
	newTail := oldTail + 1
	for ; newTail-oldHead < r.capacity; newTail++ {
		tailNode := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newTail&r.mask])))
		// not published yet
		if tailNode != nil {
			break
		}

		v, finish := valueSupplier()
		if finish {
			break
		}
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newTail&r.mask])), unsafe.Pointer(&v))
	}
	atomic.StoreUint64(&r.tail, newTail-1)
	if isEmpty && r.notify != nil {
		select {
		case r.notify <- 1:
		default:
		}
	}
}

func (r *SPMC[T]) Poll() (value T, success bool) {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := atomic.LoadUint64(&r.head)
	if r.isEmpty(oldTail, oldHead) {
		return r.null, false
	}

	newHead := oldHead + 1
	headNode := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newHead&r.mask])))
	// not published yet
	if headNode == nil {
		return r.null, false
	}
	if !atomic.CompareAndSwapUint64(&r.head, oldHead, newHead) {
		return r.null, false
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newHead&r.mask])), nil)

	return *(*T)(headNode), true
}

// isFull check whether buffer is full by compare (tail - head).
// Because of none-sync read of tail and head, the tail maybe smaller than head(which is
// never happened in the view of buffer):
//
// Say if the thread read tail=4 at time point one (in this time head=3), then wait to
// get scheduled, after a long wait, at time point two (in this time tail=8), the thread
// read head=7. So at the view in the thread, tail=4 and head=7.
//
// Hence, once tail < head means the tail is far behind the real (which means CAS-tail will
// definitely fail), so we just return full to the Offer caller let it try again.
func (r *SPMC[T]) isFull(tail uint64, head uint64) bool {
	return tail-head >= r.capacity-1
}

// isEmpty check whether buffer is empty by compare (tail - head).
// Same as isFull, the tail also may be smaller than head at thread view, which can be lead
// to wrong result:
//
// consider consumer c1 get tail=3 head=5, but actually the latest tail=5
// (means buffer empty), if we continue Poll, the dirty value can be fetched, maybe nil -
// which is harmless, maybe old value that have not been set to nil yet (by last consumer)
// - which is fatal.
//
// To keep the correctness of ring buffer, we need to return true when tail < head and
// tail == head.
func (r *SPMC[T]) isEmpty(tail uint64, head uint64) bool {
	return (tail < head) || (tail-head == 0)
}
