package pool

import (
	"errors"
	"fmt"
	"github.com/moontrade/kirana/pkg/atomicx"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/gid"
	"github.com/moontrade/kirana/pkg/mpmc"
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/runtimex"
	"github.com/moontrade/kirana/pkg/spinlock"
	"math"
	"runtime"
	"unsafe"
)

var (
	ErrNeedAllocFunc = errors.New("need alloc func")
)

type ShardFunc func() int

func ShardByProcessor() int { return int(gid.PID()) }

func ShardByGoroutineID() int { return int(gid.GID()) }

type AllocFunc[T any] func() unsafe.Pointer

type DeallocFunc[T any] func(pointer unsafe.Pointer)

type InitFunc[T any] func(pointer unsafe.Pointer)

type DeInitFunc[T any] func(pointer unsafe.Pointer)

type Config[T any] struct {
	SizeClass, NumShards    int
	PageSize, PagesPerShard int64
	ShardFunc
	AllocFunc[T]
	DeallocFunc[T]
	InitFunc[T]
	DeInitFunc[T]
}

func (c *Config[T]) Validate() {
	if c.AllocFunc == nil {
		c.AllocFunc = func() unsafe.Pointer {
			return unsafe.Pointer(new(T))
		}
	}
	if c.NumShards < 1 {
		c.NumShards = runtime.GOMAXPROCS(0)
	}
	c.NumShards = pmath.CeilToPowerOf2(c.NumShards)
	if c.PageSize < 2 {
		c.PageSize = 1024
	}
	c.PageSize = int64(pmath.CeilToPowerOf2(int(c.PageSize)))
	if c.PagesPerShard < 1 {
		c.PagesPerShard = 1
	} else {
		c.PagesPerShard = int64(pmath.CeilToPowerOf2(int(c.PagesPerShard)))
	}
	if c.ShardFunc == nil {
		c.ShardFunc = ShardByProcessor
	}
}

func (c *Config[T]) defaults() {
	if c.NumShards < 1 {
		c.NumShards = runtime.GOMAXPROCS(0)
	}
	c.NumShards = pmath.CeilToPowerOf2(c.NumShards)
	if c.PageSize < 2 {
		c.PageSize = 1024
	}
	c.PageSize = int64(pmath.CeilToPowerOf2(int(c.PageSize)))
	if c.PagesPerShard < 1 {
		c.PagesPerShard = 1
	} else {
		c.PagesPerShard = int64(pmath.CeilToPowerOf2(int(c.PagesPerShard)))
	}
	if c.ShardFunc == nil {
		c.ShardFunc = ShardByProcessor
	}
}

type Stats struct {
	Allocs            counter.Counter
	Deallocs          counter.Counter
	PageAllocs        counter.Counter
	PageAllocAttempts counter.Counter
	PageDeallocs      counter.Counter
}

type Pool[T any] struct {
	Stats
	shards   []Shard[T]
	mask     int
	config   Config[T]
	lastMiss *Shard[T]
}

func (p *Pool[T]) SizeClass() int { return p.config.SizeClass }

func NewPool[T any](
	config Config[T],
) *Pool[T] {
	config.Validate()
	pool := &Pool[T]{
		config: config,
	}
	pool.shards = make([]Shard[T], config.NumShards)
	pool.mask = config.NumShards - 1
	for i := 0; i < len(pool.shards); i++ {
		pool.shards[i] = Shard[T]{
			pool:   pool,
			config: config,
			full:   mpmc.NewBounded[mpmc.Bounded[T]](config.PagesPerShard),
			free:   mpmc.NewBounded[mpmc.Bounded[T]](config.PagesPerShard),
		}
		pool.shards[i].pop.Store(mpmc.NewBounded[T](config.PageSize))
		pool.shards[i].push.Store(mpmc.NewBounded[T](config.PageSize))
	}
	pool.lastMiss = &pool.shards[0]
	return pool
}

func (p *Pool[T]) Shards() []Shard[T] {
	return p.shards
}

func (p *Pool[T]) Shard() *Shard[T] {
	pid := runtimex.Pid()
	if len(p.shards) <= pid {
		return &p.shards[pid]
	} else {
		return &p.shards[pid&p.mask]
	}
}

func (p *Pool[T]) Get() *T {
	return (*T)(p.Shard().GetUnsafe())
}

func (p *Pool[T]) GetUnsafe() unsafe.Pointer {
	pid := runtimex.Pid()
	if pid < len(p.shards) {
		shard := &p.shards[pid]
		v := shard.GetUnsafe()
		return v
	} else {
		shard := &p.shards[pid%len(p.shards)]
		v := shard.GetUnsafe()
		return v
	}
}

func (p *Pool[T]) Put(data *T) {
	//p.lastMiss.PutUnsafe0(unsafe.Pointer(data))
	//p.Shard().PutUnsafe(unsafe.Pointer(data))
	if p.lastMiss.TryPutUnsafe(unsafe.Pointer(data)) {
		return
	}
	pid := runtimex.Pid()
	if len(p.shards) <= pid {
		shard := &p.shards[pid]
		shard.PutUnsafe(unsafe.Pointer(data))
	} else {
		shard := &p.shards[pid%len(p.shards)]
		shard.PutUnsafe(unsafe.Pointer(data))
	}
}

func (p *Pool[T]) PutUnsafe(data unsafe.Pointer) {
	if p.lastMiss.TryPutUnsafe(data) {
		return
	}
	pid := runtimex.Pid()
	if len(p.shards) <= pid {
		shard := &p.shards[pid]
		shard.PutUnsafe(data)
	} else {
		shard := &p.shards[pid%len(p.shards)]
		shard.PutUnsafe(data)
	}
}

type ShardStats struct {
	Allocates   counter.Counter
	Deallocates counter.Counter
}

type Shard[T any] struct {
	ShardStats
	pool   *Pool[T]
	pid    int
	config Config[T]
	pop    atomicx.Pointer[mpmc.Bounded[T]]
	push   atomicx.Pointer[mpmc.Bounded[T]]
	full   *mpmc.Bounded[mpmc.Bounded[T]]
	free   *mpmc.Bounded[mpmc.Bounded[T]]
	mu     spinlock.Mutex
}

func (s *Shard[T]) Pool() *Pool[T] { return s.pool }

func (s *Shard[T]) SizeClass() int { return s.config.SizeClass }

func (s *Shard[T]) Len() int {
	var (
		pop    = s.pop.Load()
		push   = s.push.Load()
		length = pop.Len()
	)
	if pop != push {
		length += push.Len()
	}
	if s.full.Len() > 0 {
		_, length = mpmc.Reduce[mpmc.Bounded[T], int](s.full, length, sumLength[T])
	}
	return length
}

func (s *Shard[T]) FreePages() int {
	return s.free.Len()
}

func sumLength[T any](item *mpmc.Bounded[T], sum int) int {
	return sum + item.Len()
}

func (s *Shard[T]) allocate() unsafe.Pointer {
	//s.Allocates.Incr()
	s.pool.Allocs.Incr()
	s.pool.lastMiss = s
	value := s.config.AllocFunc()
	if value == nil {
		return nil
	}
	if s.config.InitFunc != nil {
		s.config.InitFunc(value)
	}
	return value
}

func (s *Shard[T]) deallocate(item unsafe.Pointer) {
	s.pool.Deallocs.Incr()
	//s.Deallocates.Incr()
	if s.config.DeallocFunc != nil {
		s.config.DeallocFunc(item)
	}
}

func (s *Shard[T]) deallocatePage(page *mpmc.Bounded[T]) {
	s.pool.PageDeallocs.Incr()
	page.PopManyUnsafe(math.MaxInt, s.deallocate)
}

func (s *Shard[T]) GetUnsafe0() unsafe.Pointer {
	result := s.pop.Get().PopUnsafe()
	if result != nil {
		return result
	}
	return s.allocate()
}

func (s *Shard[T]) GetUnsafe() unsafe.Pointer {
	v := s.getUnsafe()
	return v
}

func (s *Shard[T]) getUnsafe() unsafe.Pointer {
	pop := s.pop.Get()
	v := pop.PopUnsafe()
	// Empty?
	if v != nil {
		if s.config.InitFunc != nil {
			s.config.InitFunc(v)
		}
		return v
	}

	//pop = s.pop.Load()
	push := s.push.Load()
	v = push.PopUnsafe()
	if v != nil {
		if s.config.InitFunc != nil {
			s.config.InitFunc(v)
		}
		return v
	}
	//if pop == push {
	//	return s.allocate()
	//}

	// Try to get the next full list
	next := s.full.Pop()

	// Is the full list empty?
	if next == nil {
		// Somehow emptied before we could pop 1 out. Unlikely to impossible!
		return s.allocate()
	}

	if !s.pop.CAS(pop, next) {
		// Push back into full
		if !s.full.Push(next) {
			s.deallocatePage(next)
		}
		pop = s.pop.Load()
		v = pop.PopUnsafe()
		if v == nil {
			v = s.allocate()
		}
		return v
	}

	// Push into free list
	s.free.Push(pop)

	v = next.PopUnsafe()
	if v == nil {
		v = s.allocate()
	}
	return v
}

func (s *Shard[T]) PutUnsafe0(data unsafe.Pointer) {
	s.pop.Get().PushUnsafe(data)
}

func (s *Shard[T]) TryPutUnsafe(data unsafe.Pointer) bool {
	return s.pop.Get().PushUnsafe(data)
}

func (s *Shard[T]) PutUnsafe(data unsafe.Pointer) {
	//_ = runtimex.Pin()
	s.putUnsafe(data)
}

func (s *Shard[T]) putUnsafe(data unsafe.Pointer) {
	if data == nil {
		return
	}
	if s.config.DeInitFunc != nil {
		s.config.DeInitFunc(data)
	}
	push := s.push.Load()
	if push.PushUnsafe(data) {
		return
	}

	// Get pop
	pop := s.pop.Load()
	if pop.PushUnsafe(data) {
		return
	}
	// Are push and pop the same?
	//if push == pop {
	//	goto NextFree
	//}

	//// Try setting pop to the now full push
	//if s.pop.CAS(pop, push) {
	//	// Race?
	//	if !s.push.CAS(push, pop) {
	//		// Atomically load and try to push
	//		push = s.push.Load()
	//		if !push.PushUnsafe(data) {
	//			goto NextFree
	//		}
	//	} else {
	//		// Push into the new ring which is now the active push
	//		if !pop.PushUnsafe(data) {
	//			goto NextFree
	//		}
	//	}
	//	return
	//} else {
	//	// Atomically load and try to push
	//	push = s.push.Load()
	//	if !push.PushUnsafe(data) {
	//		goto NextFree
	//	}
	//	return
	//}

NextFree:
	// Get the next free
	next := s.free.Pop()

	// Anything in the free list?
	if next == nil {
		s.pool.PageAllocAttempts.Incr()
		// Allocate new page
		next = mpmc.NewBounded[T](s.config.PageSize)
		// Race?
		if !s.push.CAS(push, next) {
			push = s.push.Load()
			// Atomically load and try to push
			if !s.push.Load().PushUnsafe(data) {
				goto NextFree
				//s.deallocate(data)
			}
		} else {
			s.pool.PageAllocs.Incr()
			// Push into the new ring which is now the active push
			if !next.PushUnsafe(data) {
				s.deallocate(data)
			}
			// Add push to full list
			if !s.full.Push(push) {
				s.deallocatePage(push)
			}
		}
	} else {
		if !next.IsEmpty() {
			if !s.full.Push(next) {
				s.deallocatePage(next)
			}
			fmt.Println("free list had non empty with size", next.Len())
			goto NextFree
		}
		// Atomically set to the new push
		if !s.push.CAS(push, next) {
			push = s.push.Load()
			// Put back into free list if the CAS failed
			s.free.Push(next)
			// Atomically load and try to push
			if !s.push.Load().PushUnsafe(data) {
				goto NextFree
				//s.deallocate(data)
			}
		} else {
			// Push into the new active push
			if !next.PushUnsafe(data) {
				s.deallocate(data)
			}
			// Add push to full list
			if !s.full.Push(push) {
				s.deallocatePage(push)
			}
		}
	}
}
