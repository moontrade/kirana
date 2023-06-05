package pool

import (
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/pmath"
	"reflect"
	"unsafe"
)

// SizeClasses are power of 2 up to 1G
// Pool is unlikely the ideal choice for sizes larger than 8kb. Best use cases
// are a lot of smallish allocations and frees with high contention.
type SizeClasses struct {
	Size8    SizeClass
	Size16   SizeClass
	Size32   SizeClass
	Size64   SizeClass
	Size128  SizeClass
	Size256  SizeClass
	Size512  SizeClass
	Size1K   SizeClass
	Size2K   SizeClass
	Size4K   SizeClass
	Size8K   SizeClass
	Size16K  SizeClass
	Size32K  SizeClass
	Size64K  SizeClass
	Size128K SizeClass
	Size256K SizeClass
	Size512K SizeClass
	Size1M   SizeClass
	Size2M   SizeClass
	Size4M   SizeClass
	Size8M   SizeClass
	Size16M  SizeClass
	Size32M  SizeClass
	Size64M  SizeClass
	Size128M SizeClass
	Size256M SizeClass
	Size512M SizeClass
	Size1G   SizeClass
}

// SizeClass configures the maximum number of pages and the size of each page.
type SizeClass struct {
	Size     int64
	PageSize int64
}

func (sc *SizeClass) IsActive() bool {
	return sc.PageSize > 0
}

type Slices[T any] struct {
	pools     []*Pool[[]T]
	sizeOf    int
	ptrData   uintptr
	missCount counter.Counter
}

func newSlicePool[T any](sizeClass int, numShards int, shardConfig SizeClass) *Pool[[]T] {
	return NewPool[[]T](Config[[]T]{
		SizeClass: sizeClass,
		NumShards: numShards,
		PageSize:  shardConfig.PageSize,
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
		pools[0] = newSlicePool[T](8, config.NumShards, sizes.Size8)
		pools[1] = pools[0]
		pools[2] = pools[0]
		pools[3] = pools[0]
	}
	if sizes.Size16.IsActive() {
		pools[4] = newSlicePool[T](16, config.NumShards, sizes.Size16)
	}
	if sizes.Size32.IsActive() {
		pools[5] = newSlicePool[T](32, config.NumShards, sizes.Size32)
	}
	if sizes.Size64.IsActive() {
		pools[6] = newSlicePool[T](64, config.NumShards, sizes.Size64)
	}
	if sizes.Size128.IsActive() {
		pools[7] = newSlicePool[T](128, config.NumShards, sizes.Size128)
	}
	if sizes.Size256.IsActive() {
		pools[8] = newSlicePool[T](256, config.NumShards, sizes.Size256)
	}
	if sizes.Size512.IsActive() {
		pools[9] = newSlicePool[T](512, config.NumShards, sizes.Size512)
	}
	if sizes.Size1K.IsActive() {
		pools[10] = newSlicePool[T](1024, config.NumShards, sizes.Size1K)
	}
	if sizes.Size2K.IsActive() {
		pools[11] = newSlicePool[T](1024*2, config.NumShards, sizes.Size2K)
	}
	if sizes.Size4K.IsActive() {
		pools[12] = newSlicePool[T](1024*4, config.NumShards, sizes.Size4K)
	}
	if sizes.Size8K.IsActive() {
		pools[13] = newSlicePool[T](1024*8, config.NumShards, sizes.Size8K)
	}
	if sizes.Size16K.IsActive() {
		pools[14] = newSlicePool[T](1024*16, config.NumShards, sizes.Size16K)
	}
	if sizes.Size32K.IsActive() {
		pools[15] = newSlicePool[T](1024*32, config.NumShards, sizes.Size32K)
	}
	if sizes.Size64K.IsActive() {
		pools[16] = newSlicePool[T](1024*64, config.NumShards, sizes.Size64K)
	}
	if sizes.Size128K.IsActive() {
		pools[17] = newSlicePool[T](1024*128, config.NumShards, sizes.Size128K)
	}
	if sizes.Size256K.IsActive() {
		pools[18] = newSlicePool[T](1024*256, config.NumShards, sizes.Size256K)
	}
	if sizes.Size512K.IsActive() {
		pools[19] = newSlicePool[T](1024*512, config.NumShards, sizes.Size512K)
	}
	if sizes.Size1M.IsActive() {
		pools[20] = newSlicePool[T](1024*1024, config.NumShards, sizes.Size1M)
	}
	if sizes.Size2M.IsActive() {
		pools[21] = newSlicePool[T](1024*1024*2, config.NumShards, sizes.Size2M)
	}
	if sizes.Size4M.IsActive() {
		pools[22] = newSlicePool[T](1024*1024*4, config.NumShards, sizes.Size4M)
	}
	if sizes.Size8M.IsActive() {
		pools[23] = newSlicePool[T](1024*1024*8, config.NumShards, sizes.Size8M)
	}
	if sizes.Size16M.IsActive() {
		pools[24] = newSlicePool[T](1024*1024*16, config.NumShards, sizes.Size16M)
	}
	if sizes.Size32M.IsActive() {
		pools[25] = newSlicePool[T](1024*1024*32, config.NumShards, sizes.Size32M)
	}
	if sizes.Size64M.IsActive() {
		pools[26] = newSlicePool[T](1024*1024*64, config.NumShards, sizes.Size64M)
	}
	if sizes.Size128M.IsActive() {
		pools[27] = newSlicePool[T](1024*1024*128, config.NumShards, sizes.Size128M)
	}
	if sizes.Size128M.IsActive() {
		pools[28] = newSlicePool[T](1024*1024*256, config.NumShards, sizes.Size256M)
	}
	if sizes.Size512M.IsActive() {
		pools[29] = newSlicePool[T](1024*1024*512, config.NumShards, sizes.Size512M)
	}
	if sizes.Size1G.IsActive() {
		pools[30] = newSlicePool[T](1024*1024*1024, config.NumShards, sizes.Size1G)
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
	item = item[:cap(item)]
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
	data = data[:cap(data)]
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
