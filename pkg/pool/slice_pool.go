package pool

import (
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/pmath"
	"reflect"
	"unsafe"
)

// SizeClasses are power of 2 up to 64kb
// Pool is unlikely the ideal choice for sizes larger than 8kb. Best use cases
// are a lot of smallish allocations and frees with high contention.
type SizeClasses struct {
	Size8    SlabConfig
	Size16   SlabConfig
	Size32   SlabConfig
	Size64   SlabConfig
	Size128  SlabConfig
	Size256  SlabConfig
	Size512  SlabConfig
	Size1KB  SlabConfig
	Size2KB  SlabConfig
	Size4KB  SlabConfig
	Size8KB  SlabConfig
	Size16KB SlabConfig
	Size32KB SlabConfig
	Size64KB SlabConfig
}

// SlabConfig configures the maximum number of pages and the size of each page.
type SlabConfig struct {
	PageSize int64
	MaxPages int64
}

func (sc *SlabConfig) IsActive() bool {
	return sc.PageSize > 0 && sc.MaxPages > 0
}

type Slices[T any] struct {
	pools     []*Pool[[]T]
	sizeOf    int
	ptrData   uintptr
	missCount counter.Counter
}

func newSlicePool[T any](sizeClass int, shardConfig SlabConfig) *Pool[[]T] {
	return NewPool[[]T](Config[[]T]{
		SizeClass:     sizeClass,
		PageSize:      shardConfig.PageSize,
		PagesPerShard: shardConfig.MaxPages,
		AllocFunc: func() unsafe.Pointer {
			a := make([]T, sizeClass)
			return unsafe.Pointer(&a[0])
		},
	})
}

func NewSlices[T any](config *Config[[]T], sizes *SizeClasses) *Slices[T] {
	if config == nil {
		config = &Config[[]T]{}
	}
	config.defaults()
	var v T
	sizeOf := int(unsafe.Sizeof(v))
	ptrData := ptrdataOf(v)
	if sizes == nil {
		sizes = DefaultSizeClasses()
	}
	pools := make([]*Pool[[]T], 65)
	if sizes.Size8.IsActive() {
		pools[0] = newSlicePool[T](8, sizes.Size8)
		pools[1] = pools[0]
		pools[2] = pools[0]
		pools[3] = pools[0]
	}
	if sizes.Size16.IsActive() {
		pools[4] = newSlicePool[T](16, sizes.Size16)
	}
	if sizes.Size32.IsActive() {
		pools[5] = newSlicePool[T](32, sizes.Size32)
	}
	if sizes.Size64.IsActive() {
		pools[6] = newSlicePool[T](64, sizes.Size64)
	}
	if sizes.Size128.IsActive() {
		pools[7] = newSlicePool[T](128, sizes.Size128)
	}
	if sizes.Size256.IsActive() {
		pools[8] = newSlicePool[T](256, sizes.Size256)
	}
	if sizes.Size512.IsActive() {
		pools[9] = newSlicePool[T](512, sizes.Size512)
	}
	if sizes.Size1KB.IsActive() {
		pools[10] = newSlicePool[T](1024, sizes.Size1KB)
	}
	if sizes.Size2KB.IsActive() {
		pools[11] = newSlicePool[T](2048, sizes.Size2KB)
	}
	if sizes.Size4KB.IsActive() {
		pools[12] = newSlicePool[T](4096, sizes.Size4KB)
	}
	if sizes.Size8KB.IsActive() {
		pools[13] = newSlicePool[T](8192, sizes.Size8KB)
	}
	if sizes.Size16KB.IsActive() {
		pools[14] = newSlicePool[T](16384, sizes.Size16KB)
	}
	if sizes.Size32KB.IsActive() {
		pools[15] = newSlicePool[T](32768, sizes.Size32KB)
	}
	if sizes.Size64KB.IsActive() {
		pools[16] = newSlicePool[T](65536, sizes.Size64KB)
	}
	if sizes.Size64KB.IsActive() {
		pools[17] = newSlicePool[T](65536*2, sizes.Size64KB)
	}
	return &Slices[T]{
		pools:   pools,
		sizeOf:  sizeOf,
		ptrData: ptrData,
	}
}

func (s *Slices[T]) PoolOf(size int) *Pool[[]T] {
	return s.pools[pmath.PowerOf2Index(size)]
}

func (s *Slices[T]) Slab(size int) SlicesSlab[T] {
	pool := s.PoolOf(size)
	if pool == nil {
		return SlicesSlab[T]{
			p: s,
		}
	}
	return SlicesSlab[T]{
		p: s,
		s: pool.Shard(),
	}
}

func (s *Slices[T]) Get(size int) []T {
	p := s.PoolOf(size)
	if p == nil {
		s.missCount.Incr()
		return make([]T, size)
	}
	item := p.GetUnsafe()
	if item == nil {
		return nil
	}
	return *(*[]T)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(item),
		Len:  p.config.SizeClass,
		Cap:  p.config.SizeClass,
	}))
}

func (s *Slices[T]) Put(item []T) {
	if item == nil {
		return
	}
	p := s.PoolOf(cap(item))
	if p == nil {
		return
	}
	p.PutUnsafe(unsafe.Pointer(&item[0]))
}

func (s *Slices[T]) PutZeroed(item []T) {
	if cap(item) == 0 {
		return
	}
	if s.ptrData > 0 {
		memclrHasPointers(unsafe.Pointer(&item[0]), uintptr(cap(item)*s.sizeOf))
	} else {
		memclrNoHeapPointers(unsafe.Pointer(&item[0]), uintptr(cap(item)*s.sizeOf))
	}
	p := s.PoolOf(cap(item))
	if p == nil {
		return
	}
	p.PutUnsafe(unsafe.Pointer(&item[0]))
}

type SlicesSlab[T any] struct {
	p *Slices[T]
	s *Shard[[]T]
}

func (s SlicesSlab[T]) Get() []T {
	if s.s == nil {
		return make([]T, s.s.config.SizeClass)
	}
	slice := s.s.GetUnsafe()
	if slice == nil {
		return make([]T, s.s.config.SizeClass)
	}
	return *(*[]T)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(slice),
		Len:  s.s.config.SizeClass,
		Cap:  s.s.config.SizeClass,
	}))
}

func (s SlicesSlab[T]) Put(data []T) {
	if s.s == nil || cap(data) != s.s.config.SizeClass {
		s.p.Put(data)
		return
	}
	s.s.PutUnsafe(unsafe.Pointer(&data[0]))
}

func (s SlicesSlab[T]) PutZeroed(data []T) {
	if cap(data) == 0 {
		return
	}
	if s.p.ptrData > 0 {
		memclrHasPointers(unsafe.Pointer(&data[0]), uintptr(cap(data)*s.p.sizeOf))
	} else {
		memclrNoHeapPointers(unsafe.Pointer(&data[0]), uintptr(cap(data)*s.p.sizeOf))
	}
	if s.s == nil || cap(data) != s.s.config.SizeClass {
		s.p.Put(data)
		return
	}
	s.s.PutUnsafe(unsafe.Pointer(&data[0]))
}
