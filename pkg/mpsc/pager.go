package mpsc

import (
	"github.com/moontrade/wormhole/pkg/atomicx"
	"github.com/moontrade/wormhole/pkg/counter"
	"github.com/moontrade/wormhole/pkg/pmath"
	"github.com/moontrade/wormhole/pkg/pool"
)

type Pager[T any] struct {
	pageSize  int64
	pop       atomicx.Pointer[Bounded[T]]
	push      atomicx.Pointer[Bounded[T]]
	backlog   *Bounded[Bounded[T]]
	pool      *pool.Pool[Bounded[T]]
	allocates counter.Counter
	wake      chan int
}

func NewPager[T any](pageSize, maxPages int64, wake chan int) *Pager[T] {
	if pageSize < 4 {
		pageSize = 4
	}
	pageSize = int64(pmath.CeilToPowerOf2(int(pageSize)))
	if maxPages < 2 {
		maxPages = 2
	}
	maxPages = int64(pmath.CeilToPowerOf2(int(maxPages)))
	p := &Pager[T]{
		pageSize: pageSize,
		wake:     wake,
		backlog:  NewBounded[Bounded[T]](maxPages, wake),
	}
	p.pop.Store(p.alloc())
	p.push.Store(p.pop.Load())
	return p
}

func sumLength[T any](item *Bounded[T], sum int) int {
	return sum + item.Len()
}

func (p *Pager[T]) Len() int {
	var (
		pop    = p.pop.Load()
		push   = p.push.Load()
		length = pop.Len()
	)
	if pop != push {
		length += push.Len()
	}
	if p.backlog.Len() > 0 {
		_, length = Reduce[Bounded[T], int](p.backlog, length, sumLength[T])
	}
	return length
}

func (p *Pager[T]) alloc() *Bounded[T] {
	if p.pool != nil {
		return p.pool.Get()
	}
	p.allocates.Incr()
	return NewBounded[T](p.pageSize, p.wake)
}

func (p *Pager[T]) dealloc(b *Bounded[T]) {
	if p.pool != nil {
		p.pool.Put(b)
	}
}

func (p *Pager[T]) Pop() *T {
	pop := p.pop.Get()
	v := pop.Pop()
	if v != nil {
		return v
	}

	push := p.push.Load()
	// Is empty?
	if pop == push {
		return nil
	}

	next := p.backlog.Pop()
	for next != nil {
		v = next.Pop()
		if v != nil {
			p.pop.Store(next)
			return v
		}
		p.dealloc(next)
		next = p.backlog.Pop()
	}

	if !p.push.CAS(push, pop) {
		next = p.backlog.Pop()
		if next == nil {
			return nil
		}
		p.pop.Store(next)
		return next.Pop()
	}

	return push.Pop()
}

func (p *Pager[T]) Push(item *T) bool {
	push := p.push.Load()
	if push.Push(item) {
		return true
	}
	if p.backlog.IsFull() {
		return false
	}
	next := p.alloc()
	next.Push(item)
	if !p.push.CAS(push, next) {
		next.Pop()
		p.dealloc(next)
		push = p.push.Load()
	} else {
		if !p.backlog.Push(push) {
			panic("dropped")
		}
		push = next
	}
	return push.Push(item)
}
