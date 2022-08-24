//go:build tinygo.wasm || 386 || amd64 || amd64p32 || arm || arm64 || loong64 || mips64le || mips64p32 || mips64p32le || mipsle || ppc64le || riscv || riscv64 || wasm

package aof

import (
	"sync/atomic"
	"unsafe"
)

func write64LE(b unsafe.Pointer, v uint64) {
	*(*uint64)(b) = v
}

func read32(b unsafe.Pointer) uint64 {
	return uint64(*(*uint32)(b))
}

func read64(p unsafe.Pointer) uint64 {
	return *(*uint64)(p)
}

func readUpTo24(p unsafe.Pointer, l uint64) uint64 {
	return uint64(*(*byte)(p))<<16 |
		uint64(*(*byte)(unsafe.Add(p, l>>1)))<<8 |
		uint64(*(*byte)(unsafe.Add(p, l-1)))
}

func storeInt64LE(b unsafe.Pointer, v int64) {
	atomic.StoreInt64((*int64)(b), v)
}

func loadInt64LE(p unsafe.Pointer) int64 {
	return atomic.LoadInt64((*int64)(p))
}

func lastByteUint64LE(v uint64) byte {
	return (*(*[8]byte)(unsafe.Pointer(&v)))[7]
}
