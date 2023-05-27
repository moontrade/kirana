package pool

import (
	"unsafe"
)

var defaultBytes *ByteSlices

func init() {
	defaultBytes = NewByteSlices(DefaultSizeClasses())
	//pools := defaultBytes.s.pools
	//for _, p := range pools {
	//	if p == nil {
	//		continue
	//	}
	//	if p.SizeClass() <= 1024 {
	//		for i := 0; i < len(p.shards); i++ {
	//			p.shards[i].fillHalf()
	//		}
	//	}
	//}
}

func DefaultBytes() *ByteSlices {
	return defaultBytes
}

func DefaultSizeClasses() *SizeClasses {
	return &SizeClasses{
		Size8:   SizeClass{16384, 1024},
		Size16:  SizeClass{16384, 256},
		Size32:  SizeClass{16384, 256},
		Size64:  SizeClass{16384, 256},
		Size128: SizeClass{16384, 256},
		Size256: SizeClass{1024, 256},
		Size512: SizeClass{64, 256},
		Size1K:  SizeClass{64, 64},
		Size2K:  SizeClass{64, 64},
		Size4K:  SizeClass{64, 64},
		Size8K:  SizeClass{64, 64},
		Size16K: SizeClass{64, 64},
		Size32K: SizeClass{64, 64},
		Size64K: SizeClass{64, 64},
	}
}

//func DefaultSizeClasses() *SizeClasses {
//	return &SizeClasses{
//		Size8:     SizeClass{1024 * 1024, 1024},
//		Size16:    SizeClass{1024 * 1024, 1024},
//		Size32:    SizeClass{1024 * 1024, 1024},
//		Size64:    SizeClass{1024 * 1024, 1024},
//		Size128:   SizeClass{1024 * 1024, 1024},
//		Size256:   SizeClass{1024, 1024},
//		Size512:   SizeClass{1024, 1024},
//		Size1K:  SizeClass{256, 512},
//		Size2K:  SizeClass{256, 256},
//		Size4K:  SizeClass{256, 256},
//		Size8K:  SizeClass{256, 256},
//		Size16K: SizeClass{128, 128},
//		Size32K: SizeClass{128, 128},
//		Size64K: SizeClass{64, 64},
//	}
//}

type ByteSlices struct {
	s *Slices[byte]
}

func NewByteSlices(sizes *SizeClasses) *ByteSlices {
	return &ByteSlices{s: NewSlices[byte](&Config[[]byte]{
		//NumShards: runtime.GOMAXPROCS(0),
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
