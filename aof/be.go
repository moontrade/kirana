//go:build arm64be || armbe || mips || mips64 || ppc || ppc64 || s390 || s390x || sparc || sparc64

package wyhash

import (
	"math/bits"
	"sync/atomic"
	"unsafe"
)

func write64LE(b unsafe.Pointer, v uint64) {
	*(*uint64)(b) = bits.ReverseBytes64(v)
}

func read32(b unsafe.Pointer) uint64 {
	return bits.ReverseBytes64(uint64(*(*uint32)(b)))
}

func read64(p unsafe.Pointer) uint64 {
	return bits.ReverseBytes64(*(*uint64)(p))
}

func storeInt64LE(b unsafe.Pointer, v int64) {
	atomic.StoreInt64((*int64)(b), v)
}

func loadInt64LE(p unsafe.Pointer) int64 {
	r := bits.ReverseBytes64(atomic.LoadUint64((*uint64)(p)))
	return (*int64)(unsafe.Pointer(&r))
}

func load64LE(p unsafe.Pointer) uint64 {
	r := bits.ReverseBytes64(atomic.LoadUint64((*uint64)(p)))
	return (*int64)(unsafe.Pointer(&r))
}
