//go:build tinygo.wasm || 386 || amd64 || amd64p32 || arm || arm64 || loong64 || mips64le || mips64p32 || mips64p32le || mipsle || ppc64le || riscv || riscv64 || wasm

package pmath

import (
	"math/bits"
)

func PowerOf2Index(size int) int {
	return bits.TrailingZeros64(uint64(CeilToPowerOf2(size)))
}
