//go:build tinygo.wasm || 386 || amd64 || amd64p32 || arm || arm64 || loong64 || mips64le || mips64p32 || mips64p32le || mipsle || ppc64le || riscv || riscv64 || wasm

package message

import (
	"math/bits"
	"unsafe"
)

///////////////////////////////////////////////////////////////////////////////////////////////
// Int16 Little Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsInt16LE() int16 {
	return *(*int16)(p.p)
}

func Int16LE(p unsafe.Pointer) int16 {
	return *(*int16)(p)
}

func (p Pointer) Int16LE(offset int) int16 {
	return *(*int16)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetInt16LE(offset int, v int16) {
	*(*int16)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Int16 Big Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsInt16BE() int16 {
	return int16(bits.ReverseBytes16(*(*uint16)(p.p)))
}

func (p Pointer) Int16BE(offset int) int16 {
	return int16(bits.ReverseBytes16(*(*uint16)(unsafe.Add(p.p, offset))))
}

func (p Pointer) SetInt16BE(offset int, v int16) {
	*(*int16)(unsafe.Add(p.p, offset)) = int16(bits.ReverseBytes16(uint16(v)))
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Uint16 Little Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsUint16LE() uint16 {
	return *(*uint16)(p.p)
}

func Uint16LE(p unsafe.Pointer) uint16 {
	return *(*uint16)(p)
}

func (p Pointer) Uint16LE(offset int) uint16 {
	return *(*uint16)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetUint16LE(offset int, v uint16) {
	*(*uint16)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Uint16 Big Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsUint16BE() uint16 {
	return bits.ReverseBytes16(*(*uint16)(p.p))
}

func (p Pointer) Uint16BE(offset int) uint16 {
	return bits.ReverseBytes16(*(*uint16)(unsafe.Add(p.p, offset)))
}

func (p Pointer) SetUint16BE(offset int, v uint16) {
	*(*uint16)(unsafe.Add(p.p, offset)) = bits.ReverseBytes16(v)
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Int32 Little Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsInt32LE() int32 {
	return *(*int32)(p.p)
}

func (p Pointer) Int32LE(offset int) int32 {
	return *(*int32)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetInt32LE(offset int, v int32) {
	*(*int32)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Int32 Big Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsInt32BE() int32 {
	return int32(bits.ReverseBytes32(*(*uint32)(p.p)))
}

func (p Pointer) Int32BE(offset int) int32 {
	return int32(bits.ReverseBytes32(*(*uint32)(unsafe.Add(p.p, offset))))
}

func (p Pointer) SetInt32BE(offset int, v int32) {
	*(*int32)(unsafe.Add(p.p, offset)) = int32(bits.ReverseBytes32(uint32(v)))
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Uint32 Little Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsUint32LE() uint32 {
	return *(*uint32)(p.p)
}

func (p Pointer) Uint32LE(offset int) uint32 {
	return *(*uint32)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetUint32LE(offset int, v uint32) {
	*(*uint32)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Uint32 Big Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsUint32BE() uint32 {
	return bits.ReverseBytes32(*(*uint32)(p.p))
}

func (p Pointer) Uint32BE(offset int) uint32 {
	return bits.ReverseBytes32(*(*uint32)(unsafe.Add(p.p, offset)))
}

func (p Pointer) SetUint32BE(offset int, v uint32) {
	*(*uint32)(unsafe.Add(p.p, offset)) = bits.ReverseBytes32(v)
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Int64 Little Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsInt64LE() int64 {
	return *(*int64)(p.p)
}

func (p Pointer) Int64LE(offset int) int64 {
	return *(*int64)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetInt64LE(offset int, v int64) {
	*(*int64)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Int64 Big Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsInt64BE() int64 {
	return int64(bits.ReverseBytes64(*(*uint64)(p.p)))
}

func (p Pointer) Int64BE(offset int) int64 {
	return int64(bits.ReverseBytes64(*(*uint64)(unsafe.Add(p.p, offset))))
}

func (p Pointer) SetInt64BE(offset int, v int64) {
	*(*int64)(unsafe.Add(p.p, offset)) = int64(bits.ReverseBytes64(uint64(v)))
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Uint64 Little Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsUint64LE() uint64 {
	return *(*uint64)(p.p)
}

func (p Pointer) Uint64LE(offset int) uint64 {
	return *(*uint64)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetUint64LE(offset int, v uint64) {
	*(*uint64)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Uint64 Big Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsUint64BE() uint64 {
	return bits.ReverseBytes64(*(*uint64)(p.p))
}

func (p Pointer) Uint64BE(offset int) uint64 {
	return bits.ReverseBytes64(*(*uint64)(unsafe.Add(p.p, offset)))
}

func (p Pointer) SetUint64BE(offset int, v uint64) {
	*(*uint64)(unsafe.Add(p.p, offset)) = bits.ReverseBytes64(v)
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Float32 Little Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsFloat32LE() float32 {
	return *(*float32)(p.p)
}

func (p Pointer) Float32LE(offset int) float32 {
	return *(*float32)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetFloat32LE(offset int, v float32) {
	*(*float32)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Float32 Big Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsFloat32BE() float32 {
	return float32(bits.ReverseBytes32(*(*uint32)(p.p)))
}

func (p Pointer) Float32BE(offset int) float32 {
	return float32(bits.ReverseBytes32(*(*uint32)(unsafe.Add(p.p, offset))))
}

func (p Pointer) SetFloat32BE(offset int, v float32) {
	*(*float32)(unsafe.Add(p.p, offset)) = float32(bits.ReverseBytes32(uint32(v)))
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Float64 Little Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsFloat64LE() float64 {
	return *(*float64)(p.p)
}

func (p Pointer) Float64LE(offset int) float64 {
	return *(*float64)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetFloat64LE(offset int, v float64) {
	*(*float64)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Float64 Big Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) AsFloat64BE() float64 {
	return float64(bits.ReverseBytes64(*(*uint64)(p.p)))
}

func (p Pointer) Float64BE(offset int) float64 {
	return float64(bits.ReverseBytes64(*(*uint64)(unsafe.Add(p.p, offset))))
}

func (p Pointer) SetFloat64BE(offset int, v float64) {
	*(*float64)(unsafe.Add(p.p, offset)) = float64(bits.ReverseBytes64(uint64(v)))
}
