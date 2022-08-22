//go:build tinygo.wasm || 386 || amd64 || amd64p32 || arm || arm64 || loong64 || mips64le || mips64p32 || mips64p32le || mipsle || ppc64le || riscv || riscv64 || wasm

package wyhash

import (
	"unsafe"
)

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
