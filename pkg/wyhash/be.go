//go:build arm64be || armbe || mips || mips64 || ppc || ppc64 || s390 || s390x || sparc || sparc64

package wyhash

import "math/bits"

func read32(b unsafe.Pointer) uint64 {
	return bits.ReverseBytes64(uint64(*(*uint32)(b)))
}

func read64(p unsafe.Pointer) uint64 {
	return bits.ReverseBytes64(*(*uint64)(p))
}

//func ReadUint64LE(p unsafe.Pointer) uint64 {
//	return bits.ReverseBytes64(*(*uint64)(p))
//}
//
//func ReadUint64BE(p unsafe.Pointer) uint64 {
//	return *(*uint64)(p)
//}
