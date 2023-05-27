package pool

import (
	"errors"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/gid"
	"github.com/moontrade/kirana/pkg/mpmc"
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/spinlock"
	"math"
	"runtime"
	"sync"
	"unsafe"
)

var (
	ErrNeedAllocFunc = errors.New("need alloc func")
)

type AllocFunc[T any] func() unsafe.Pointer

type DeallocFunc[T any] func(pointer unsafe.Pointer)

type InitFunc[T any] func(pointer unsafe.Pointer)

type DeInitFunc[T any] func(pointer unsafe.Pointer)

type Config[T any] struct {
	SizeClass, NumShards    int
	PageSize, PagesPerShard int64
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
}

type Stats struct {
	Allocs            counter.Counter
	Allocs2           counter.Counter
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
			//queue:  NewCircleBuf(int(config.PageSize)),
			queue: mpmc.NewBounded[T](config.PageSize),
		}
	}
	pool.lastMiss = &pool.shards[0]
	return pool
}

func (p *Pool[T]) Shards() []Shard[T] {
	return p.shards
}

func (p *Pool[T]) Len() int {
	r := 0
	for i := 0; i < len(p.shards); i++ {
		r += p.shards[i].Len()
	}
	return r
}

func (p *Pool[T]) Shard() *Shard[T] {
	pid := int(gid.PID())
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
	pid := int(gid.PID())
	if pid < len(p.shards) {
		v := p.shards[pid].GetUnsafe()
		return v
	} else {
		v := p.shards[pid&p.mask].GetUnsafe()
		return v
	}
}

func (p *Pool[T]) Put(data *T) {
	pid := int(gid.PID())
	if len(p.shards) > pid {
		p.shards[pid].PutUnsafe(unsafe.Pointer(data))
	} else {
		p.shards[pid&p.mask].PutUnsafe(unsafe.Pointer(data))
	}
}

func (p *Pool[T]) PutUnsafe(data unsafe.Pointer) {
	pid := int(gid.PID())
	if len(p.shards) > pid {
		p.shards[pid].PutUnsafe(data)
	} else {
		p.shards[pid&p.mask].PutUnsafe(data)
	}
}

type ShardStats struct {
	Allocates   counter.Counter
	Allocates2  counter.Counter
	Deallocates counter.Counter
}

type Shard[T any] struct {
	ShardStats
	pool   *Pool[T]
	pid    int
	config Config[T]
	//queue  *CircleBuf
	queue *mpmc.Bounded[T]
	mu    spinlock.Mutex
	_mu   sync.Mutex
}

func (s *Shard[T]) Pool() *Pool[T] { return s.pool }

func (s *Shard[T]) SizeClass() int { return s.config.SizeClass }

func (s *Shard[T]) Len() int {
	return s.queue.Len()
}

func (s *Shard[T]) fill(pct float64) {
	if pct <= 0 {
		return
	}
	capacity := s.queue.Cap()
	if pct < 1.0 {
		capacity = int(float64(capacity) * pct)
	}
	for i := 0; i < capacity; i++ {
		s.putUnsafe(s.getUnsafe())
	}
}

func sumLength[T any](item *mpmc.Bounded[T], sum int) int {
	return sum + item.Len()
}

func (s *Shard[T]) allocate() unsafe.Pointer {
	//s.Allocates.Incr()
	s.pool.Allocs.Incr()
	s.Allocates.Incr()
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
	s.Deallocates.Incr()
	if s.config.DeallocFunc != nil {
		s.config.DeallocFunc(item)
	}
}

func (s *Shard[T]) deallocatePage(page *mpmc.Bounded[T]) {
	s.pool.PageDeallocs.Incr()
	page.DequeueManyUnsafe(math.MaxInt, s.deallocate)
}

func (s *Shard[T]) GetUnsafe() unsafe.Pointer {
	v := s.getUnsafe()
	return v
	//return nil
}

func (s *Shard[T]) getUnsafe() unsafe.Pointer {
	//s.mu.Lock()
	//defer s.mu.Unlock()

	v := s.queue.DequeueUnsafe()
	// Hit?
	if v != nil {
		if s.config.InitFunc != nil {
			s.config.InitFunc(v)
		}
		return v
	}

	v = s.allocate()
	return v
}

func (s *Shard[T]) TryPutUnsafe(data unsafe.Pointer) bool {
	if data == nil {
		return false
	}

	if s.config.DeInitFunc != nil {
		s.config.DeInitFunc(data)
	}

	//s.mu.Lock()
	//defer s.mu.Unlock()

	if s.queue.EnqueueUnsafe(data) {
		return true
	}

	s.deallocate(data)

	return false
}

func (s *Shard[T]) PutUnsafe(data unsafe.Pointer) {
	s.putUnsafe(data)
}

func (s *Shard[T]) putUnsafe(data unsafe.Pointer) {
	if data == nil {
		return
	}

	if s.config.DeInitFunc != nil {
		s.config.DeInitFunc(data)
	}

	//s.mu.Lock()
	//defer s.mu.Unlock()

	if s.queue.EnqueueUnsafe(data) {
		return
	}

	s.deallocate(data)
}

type CircleBuf struct {
	nodes []unsafe.Pointer
	mask  int64
	head  int64
	tail  int64
}

func NewCircleBuf(capacity int) *CircleBuf {
	if capacity < 4 {
		capacity = 4
	}
	capacity = int(pmath.CeilToPowerOf2(int(capacity)))
	return &CircleBuf{
		nodes: make([]unsafe.Pointer, capacity),
		mask:  int64(capacity - 1),
		head:  0,
		tail:  0,
	}
}

func (c *CircleBuf) Enqueue(v unsafe.Pointer) bool {
	if c.tail-c.head >= c.mask {
		return false
	}
	c.nodes[c.tail&c.mask] = v
	c.tail++
	return true
}

func (c *CircleBuf) EnqueueUnsafe(v unsafe.Pointer) bool {
	if c.tail-c.head >= c.mask {
		return false
	}
	c.nodes[c.tail&c.mask] = v
	c.tail++
	return true
}

func (c *CircleBuf) Dequeue() unsafe.Pointer {
	if c.tail-c.head == 0 {
		return nil
	}
	r := c.nodes[c.head&c.mask]
	c.head++
	return r
}

func (c *CircleBuf) DequeueUnsafe() unsafe.Pointer {
	if c.tail-c.head == 0 {
		return nil
	}
	r := c.nodes[c.head&c.mask]
	c.head++
	return r
}

func (c *CircleBuf) Len() int64 {
	return c.tail - c.head
}

func (c *CircleBuf) Cap() int {
	return len(c.nodes)
}
