package mpmc

import (
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/spinlock"
	"sync/atomic"
)

// Paged
type Paged[T any] struct {
	head      atomic.Pointer[Page[T]]
	tail      atomic.Pointer[Page[T]]
	pageSize  int64
	pageCount int64
	mu        spinlock.Mutex
}

func NewPaged[T any](pageSize int) *Paged[T] {
	if pageSize < 2 {
		pageSize = 2
	}
	pageSize = pmath.CeilToPowerOf2(pageSize)
	pd := &Paged[T]{pageSize: int64(pageSize), pageCount: 1}
	pd.tail.Store(pd.newPage())
	pd.head.Store(pd.tail.Load())
	pd.head.Load().right.Store(pd.tail.Load())
	pd.tail.Load().left.Store(pd.head.Load())
	return pd
}

func (pd *Paged[T]) newPage() *Page[T] {
	nodes := make([]node[T], pd.pageSize)
	for i := 0; i < len(nodes); i++ {
		atomic.StoreInt64(&nodes[i].seq, int64(i))
	}
	return &Page[T]{
		Bounded: Bounded[T]{
			head:  0,
			tail:  0,
			mask:  pd.pageSize - 1,
			nodes: nodes,
		},
	}
}

func (pd *Paged[T]) Dequeue() *T {
	head := pd.head.Load()
	v := head.Dequeue()
	if v != nil {
		return v
	}
	return nil
}

//func (pd *Paged[T]) Enqueue(value *T) bool {
//	head := pd.head.Load()
//	if head.Enqueue(value) {
//		return true
//	}
//	next := head.right.Load()
//	if tail != head && tail.Enqueue(value) {
//		return true
//	}
//
//	pd.mu.Lock()
//	currentHead := pd.head.Load()
//	currentTail := pd.tail.Load()
//	if currentTail != tail {
//		if currentTail.Enqueue(value) {
//			return true
//		}
//		currentTail = pd.newPage()
//		currentTail.Enqueue(value)
//		pd.tail
//	} else {
//		currentTail =
//	}
//	pd.mu.Unlock()
//	return true
//}

type Page[T any] struct {
	Bounded[T]
	left  atomic.Pointer[Page[T]]
	right atomic.Pointer[Page[T]]
}
