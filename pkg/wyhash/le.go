//go:build tinygo.wasm || 386 || amd64 || amd64p32 || arm || arm64 || loong64 || mips64le || mips64p32 || mips64p32le || mipsle || ppc64le || riscv || riscv64 || wasm

package wyhash

import (
	"math"
	"math/bits"
	"unsafe"
)

func read32(b unsafe.Pointer) uint64 {
	//s := *(*[4]byte)(b)
	//return uint64(binary.LittleEndian.Uint32(s[:]))
	return uint64(*(*uint32)(b))

	//return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func read64(p unsafe.Pointer) uint64 {
	//s := *(*[8]byte)(p)
	//return binary.LittleEndian.Uint64(s[:])
	return *(*uint64)(p)
}

func readUpTo24(p unsafe.Pointer, l uint64) uint64 {
	return uint64(*(*byte)(p))<<16 |
		uint64(*(*byte)(unsafe.Add(p, l>>1)))<<8 |
		uint64(*(*byte)(unsafe.Add(p, l-1)))
}

//const three = 3 - 4 - ((3 >> 3) << 2)
//const three2 = (3 >> 3) << 2
//const four = 4 - 4 - ((4 >> 3) << 2)
//const four2 = (4 >> 3) << 2
//const eight = (8 >> 3) << 2
//const eight2 = 8 - 4 - ((8 >> 3) << 2)

func I8(v int8) uint64 {
	return U8(uint8(v))
}

func U8(v uint8) uint64 {
	var (
		a = uint64(v)<<16 |
			uint64(v)<<8 |
			uint64(v)
	)
	return wymix(s1^1, wymix(a^s1, 0^defaultSeedInit))
}

func I16(v int16) uint64 {
	return U16(*(*uint16)(unsafe.Pointer(&v)))
}

func U16(v uint16) uint64 {
	var (
		//a = uint64(*(*byte)(unsafe.Pointer(&v)))<<16 |
		//	uint64(*(*byte)(unsafe.Add(unsafe.Pointer(&v), 1)))<<8 |
		//	uint64(*(*byte)(unsafe.Add(unsafe.Pointer(&v), 1)))

		a = uint64(byte(v))<<16 |
			uint64(byte(v>>8))<<8 |
			uint64(byte(v>>8))
	)
	return wymix(s1^2, wymix(a^s1, 0^defaultSeedInit))
}

func F32(v float32) uint64 {
	return U32(math.Float32bits(v))
}

func I32(v int32) uint64 {
	return U32(uint32(v))
}

func U32(v uint32) uint64 {
	var (
		vv = uint64(v)
		a  = vv<<32 | vv
	)
	return wymix(s1^4, wymix(a^s1, a^defaultSeedInit))
}

func F64(v float64) uint64 {
	return U64(math.Float64bits(v))
}

func Int(v int) uint64 {
	return U64(uint64(v))
}

func I64(v int64) uint64 {
	return U64(*(*uint64)(unsafe.Pointer(&v)))
}

func Int64(v int64) uint64 {
	return U64(*(*uint64)(unsafe.Pointer(&v)))
}

func Uint64(v uint64) uint64 {
	return U64(v)
}

func U64(v uint64) uint64 {
	var (
		a = v<<32 | v>>32
		b = ((v >> 32) << 32) | ((v << 32) >> 32)
	)
	return wymix(s1^8, wymix(a^s1, b^defaultSeedInit))
}

func wymum0(a, b uint64) uint64 {
	a, b = bits.Mul64(a, b)
	return a ^ b
}

func Uintptr(v uintptr) uint64 {
	return U64(uint64(v))
}
