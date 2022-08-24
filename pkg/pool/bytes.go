package pool

import (
	"unsafe"
)

var defaultBytes = NewByteSlices(DefaultSizeClasses())

func DefaultBytes() *ByteSlices {
	return defaultBytes
}

func DefaultSizeClasses() *SizeClasses {
	return &SizeClasses{
		Size8:    SlabConfig{4096, 256},
		Size16:   SlabConfig{4096, 256},
		Size32:   SlabConfig{4096, 256},
		Size64:   SlabConfig{4096, 256},
		Size128:  SlabConfig{4096, 256},
		Size256:  SlabConfig{4096, 1024},
		Size512:  SlabConfig{4096, 1024},
		Size1KB:  SlabConfig{1024, 64},
		Size2KB:  SlabConfig{256, 64},
		Size4KB:  SlabConfig{256, 64},
		Size8KB:  SlabConfig{256, 64},
		Size16KB: SlabConfig{64, 64},
		Size32KB: SlabConfig{64, 64},
		Size64KB: SlabConfig{64, 64},
	}
}

//func DefaultSizeClasses() *SizeClasses {
//	return &SizeClasses{
//		Size8:     SlabConfig{1024 * 1024, 1024},
//		Size16:    SlabConfig{1024 * 1024, 1024},
//		Size32:    SlabConfig{1024 * 1024, 1024},
//		Size64:    SlabConfig{1024 * 1024, 1024},
//		Size128:   SlabConfig{1024 * 1024, 1024},
//		Size256:   SlabConfig{1024, 1024},
//		Size512:   SlabConfig{1024, 1024},
//		Size1KB:  SlabConfig{256, 512},
//		Size2KB:  SlabConfig{256, 256},
//		Size4KB:  SlabConfig{256, 256},
//		Size8KB:  SlabConfig{256, 256},
//		Size16KB: SlabConfig{128, 128},
//		Size32KB: SlabConfig{128, 128},
//		Size64KB: SlabConfig{64, 64},
//	}
//}

type ByteSlices struct {
	s *Slices[byte]
}

func NewByteSlices(sizes *SizeClasses) *ByteSlices {
	return &ByteSlices{s: NewSlices[byte](&Config[[]byte]{
		//ShardFunc: ShardByGoroutineID,
	}, sizes)}
}

func (b *ByteSlices) Alloc(size int) []byte {
	return b.s.Get(size)
}

func (b *ByteSlices) AllocCap(size, capacity int) []byte {
	data := b.s.Get(capacity)
	if cap(data) == 0 {
		return nil
	}
	return data[0:size]
}

func (b *ByteSlices) AllocZeroed(size int) []byte {
	data := b.s.Get(size)
	if cap(data) == 0 {
		return nil
	}
	memclrNoHeapPointers(unsafe.Pointer(&data[0]), uintptr(cap(data)))
	return data[0:size]
}

func (b *ByteSlices) AllocZeroedCap(size, capacity int) []byte {
	data := b.s.Get(capacity)
	if cap(data) == 0 {
		return nil
	}
	memclrNoHeapPointers(unsafe.Pointer(&data[0]), uintptr(cap(data)))
	return data[0:size]
}

func (b *ByteSlices) Free(data []byte) {
	b.s.Put(data)
}

func Alloc(size int) []byte {
	return defaultBytes.Alloc(size)
}

func AllocCap(size, capacity int) []byte {
	return defaultBytes.AllocCap(size, capacity)
}

func AllocZeroed(size int) []byte {
	return defaultBytes.AllocZeroed(size)
}

func AllocZeroedCap(size, capacity int) []byte {
	return defaultBytes.AllocZeroedCap(size, capacity)
}

func Free(b []byte) {
	defaultBytes.Free(b)
}
