package message

import (
	"io"
	"math"
	"strings"
	"unsafe"
)

const MaxSize = math.MaxUint16 - 4

// Pointer is a message allocated using the Go heap and managed by Go's GC.
type Pointer struct {
	p unsafe.Pointer
}

func PointerOf(b []byte) Pointer {
	return Pointer{p: unsafe.Pointer(&b[0])}
}

func (p Pointer) Size() uint16 {
	return p.AsUint16LE()
}

func (p Pointer) Type() uint16 {
	return p.Uint16LE(2)
}

// Clone
func (p Pointer) Clone(size uintptr) Pointer {
	c := Pointer{Alloc(size)}
	Copy(c.p, p.p, size)
	return c
}

// Zero zeroes out the entire allocation.
func (p Pointer) Zero(size uintptr) {
	Zero(p.p, size)
}

// Move does a memmove
func (p Pointer) Move(offset, size int, to Pointer) {
	Move(to.p, unsafe.Add(p.p, offset), uintptr(size))
}

// Copy does a memcpy
func (p Pointer) Copy(offset, size int, to unsafe.Pointer) {
	Copy(to, unsafe.Add(p.p, offset), uintptr(size))
}

// Equals does a memequal
func (p Pointer) Equals(offset, size int, to Pointer) bool {
	return Equals(to.p, unsafe.Add(p.p, offset), uintptr(size))
}

// Compare does a memcmp
func (p Pointer) Compare(offset, size int, to Pointer) int {
	return Compare(to.p, unsafe.Add(p.p, offset), uintptr(size))
}

func (p Pointer) WriteTo(w io.Writer, size int) error {
	n, err := w.Write(p.Bytes(0, size, size))
	if err != nil {
		return err
	}
	if n != size {
		return io.ErrShortBuffer
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Byte
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) Bool(offset int) bool {
	return *(*int8)(unsafe.Add(p.p, offset)) != 0
}

func (p Pointer) Int8(offset int) int8 {
	return *(*int8)(unsafe.Add(p.p, offset))
}

func (p Pointer) Uint8(offset int) uint8 {
	return *(*uint8)(unsafe.Add(p.p, offset))
}

func (p Pointer) Byte(offset int) byte {
	return *(*byte)(unsafe.Add(p.p, offset))
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Put Byte
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) SetBool(offset int, value bool) {
	if value {
		*(*int8)(unsafe.Add(p.p, offset)) = 1
	} else {
		*(*int8)(unsafe.Add(p.p, offset)) = 0
	}
}

func (p Pointer) SetInt8(offset int, v int8) {
	*(*int8)(unsafe.Add(p.p, offset)) = v
}

func (p Pointer) SetUint8(offset int, v uint8) {
	*(*uint8)(unsafe.Add(p.p, offset)) = v
}

func (p Pointer) SetByte(offset int, v byte) {
	*(*byte)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Int16 Native Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) Int16(offset int) int16 {
	return *(*int16)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetInt16(offset int, v int16) {
	*(*int16)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Uint16 Native Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) Uint16(offset int) uint16 {
	return *(*uint16)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetUint16(offset int, v uint16) {
	*(*uint16)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Int32 Native Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) Int32(offset int) int32 {
	return *(*int32)(unsafe.Add(p.p, offset))
}
func (p Pointer) Int32Alt(offset uintptr) int32 {
	return *(*int32)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetInt32(offset int, v int32) {
	*(*int32)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Uint32 Native Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) Uint32(offset int) uint32 {
	return *(*uint32)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetUint32(offset int, v uint32) {
	*(*uint32)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Int64 Native Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) Int64(offset int) int64 {
	return *(*int64)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetInt64(offset int, v int64) {
	*(*int64)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Uint64 Native Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) Uint64(offset int) uint64 {
	return *(*uint64)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetUint64(offset int, v uint64) {
	*(*uint64)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Float32 Native Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) Float32(offset int) float32 {
	return *(*float32)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetFloat32(offset int, v float32) {
	*(*float32)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Float64 Native Endian
///////////////////////////////////////////////////////////////////////////////////////////////

func (p Pointer) Float64(offset int) float64 {
	return *(*float64)(unsafe.Add(p.p, offset))
}

func (p Pointer) SetFloat64(offset int, v float64) {
	*(*float64)(unsafe.Add(p.p, offset)) = v
}

///////////////////////////////////////////////////////////////////////////////////////////////
// String
///////////////////////////////////////////////////////////////////////////////////////////////

type _string struct {
	Data unsafe.Pointer
	Len  int
}

func (p Pointer) String(offset, size int) string {
	return *(*string)(unsafe.Pointer(&_string{
		Data: unsafe.Add(p.p, offset),
		Len:  size,
	}))
}

func (p Pointer) SetString(offset int, value string) {
	dst := *(*[]byte)(unsafe.Pointer(&_bytes{
		Data: unsafe.Add(p.p, offset),
		Len:  len(value),
		Cap:  len(value),
	}))
	copy(dst, value)
}

func (p Pointer) SetBytes(offset int, value []byte) {
	dst := *(*[]byte)(unsafe.Pointer(&_bytes{
		Data: unsafe.Add(p.p, offset),
		Len:  len(value),
		Cap:  len(value),
	}))
	copy(dst, value)
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Byte Slice
///////////////////////////////////////////////////////////////////////////////////////////////

type _bytes struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func (p Pointer) AsBytes(length int) []byte {
	return *(*[]byte)(unsafe.Pointer(&_bytes{
		Data: p.p,
		Len:  length,
		Cap:  length,
	}))
}

func (p Pointer) Bytes(offset, length, capacity int) []byte {
	return *(*[]byte)(unsafe.Pointer(&_bytes{
		Data: unsafe.Add(p.p, offset),
		Len:  length,
		Cap:  capacity,
	}))
}

func (p Pointer) StringFixed(offset int, maxLength int) string {
	s := p.String(offset, maxLength)
	index := strings.IndexByte(s, 0)
	if index > -1 {
		return s[0:index]
	}
	return s
}

func (p Pointer) SetStringFixed(offset int, maxLength int, value string) {
	if len(value) >= maxLength {
		value = value[0 : maxLength-1]
	}
	//memory.Zero(p.Pointer(offset).Unsafe(), uintptr(maxLength))
	p.SetString(offset, value)
	p.SetUint8(offset+len(value), 0)
}

//type VLS_t struct {
//	Offset uint16
//	Length uint16
//}

// StringVLS returns an unsafe Go string
func (p Pointer) StringVLS(offset int) string {
	var (
		size     = p.AsUint16LE()
		baseSize = p.Uint16LE(4)
	)
	_ = baseSize
	vlsOffset := p.Uint16LE(offset)
	vlsLength := p.Uint16LE(offset + 2)
	if vlsLength > 4096 {
		vlsLength = 4096
	}
	if vlsOffset == 0 || vlsLength == 0 {
		return ""
	}
	if size < vlsOffset+vlsLength {
		return ""
	}

	if p.Byte(int(vlsOffset+vlsLength)-1) == 0 {
		return p.String(int(vlsOffset), int(vlsLength)-1)
	} else {
		return p.String(int(vlsOffset), int(vlsLength))
	}
}

// SetStringVLS replaces existing VLS if new one fits, otherwise appends to end possibly growing
// the existing allocation in order to do so.
//func (vls *VLS) SetStringVLS(offset int, value string) {
//	vlsOffset := int(vls.Uint16LE(offset))
//	vlsLength := int(vls.Uint16LE(offset + 2))
//	newLength := len(value) + 1
//	if vlsLength >= newLength {
//		// Set new length
//		vls.SetUint16LE(offset+2, uint16(len(value)+1))
//		vls.SetString(vlsOffset, value)
//		vls.SetByte(vlsOffset+len(value), 0)
//		return
//	}
//	newSize := int(vls.Size()) + newLength
//	if newSize > MaxSize {
//		return
//	}
//	vlsOffset = int(vls.Size())
//	if vls.capacity < newSize {
//		vls.Extend(newLength)
//	}
//
//	// Set new size
//	vls.SetUint16LE(0, uint16(newSize))
//	// Set VLS offset
//	vls.SetUint16LE(offset, uint16(vlsOffset))
//	// Set VLS length
//	vls.SetUint16LE(offset+2, uint16(newLength))
//	// Set string
//	vls.SetString(vlsOffset, value)
//	// Set null terminator
//	vls.SetByte(vlsOffset+len(value), 0)
//}
