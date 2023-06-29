package logger

import (
	"encoding/binary"
	message "github.com/moontrade/kirana/logger/internal"
	"github.com/moontrade/kirana/pkg/timex"
	"reflect"
	"unsafe"
)

func AppendRecord(b []byte, ts int64, level Level, fn *FuncInfo, msg string) []byte {
	// Builder?
	//if len(args) == 0 {
	//	return b
	//}

	b = binary.LittleEndian.AppendUint64(b, uint64(ts))
	b = binary.LittleEndian.AppendUint64(b, uint64(level))
	if fn != nil {
		b = append(b, fn.formatted...)
	}

	//for _, attr := range args {
	//	switch attr.Value.Kind() {
	//	case KindString:
	//		b = append(b, attr.Value.String()...)
	//	case KindInt64:
	//		b = binary.LittleEndian.AppendUint64(b, uint64(attr.Value.Int64()))
	//	case KindUint64:
	//		b = binary.LittleEndian.AppendUint64(b, attr.Value.Uint64())
	//	case KindFloat64:
	//		b = binary.LittleEndian.AppendUint64(b, math.Float64bits(attr.Value.Float64()))
	//	case KindDuration:
	//		b = binary.LittleEndian.AppendUint64(b, uint64(attr.Value.Duration()))
	//	case KindTime:
	//		b = binary.LittleEndian.AppendUint64(b, uint64(attr.Value.Time().UnixNano()))
	//	case KindBool:
	//		if attr.Value.Bool() {
	//			b = append(b, 1)
	//		} else {
	//			b = append(b, 0)
	//		}
	//	case KindAny:
	//	case KindGroup:
	//	}
	//}

	return b
}

type Flags uint8

const (
	HighFrequency Flags = 1 << iota
)

func (f Flags) Set(b, flag Flags) Flags    { return b | flag }
func (f Flags) Clear(b, flag Flags) Flags  { return b &^ flag }
func (f Flags) Toggle(b, flag Flags) Flags { return b ^ flag }
func (f Flags) Has(b, flag Flags) bool     { return b&flag != 0 }
func (f Flags) IsHighFrequency() bool      { return f&HighFrequency != 0 }

type Record struct {
	p message.Pointer
}

func (r Record) Size() int {
	return int(r.p.Uint16LE(0))
}

func (r Record) BaseSize() int {
	return int(r.p.Uint16LE(2))
}

func (r Record) Level() Level {
	return Level(r.p.Byte(4))
}

func (r Record) Flags() Flags {
	return Flags(r.p.Byte(5))
}

type RecordBuilder struct {
	p message.Pointer
	c int
}

func AllocRecordBuilder() RecordBuilder {
	rb := RecordBuilder{p: message.PointerOf(make([]byte, 512)), c: 512}
	rb.p.SetUint16(0, uint16(unsafe.Sizeof(RecordHeader{})))
	rb.p.SetUint16(2, uint16(unsafe.Sizeof(RecordHeader{})))
	return rb
}

func (r RecordBuilder) Reset() RecordBuilder {
	r.p.Zero(unsafe.Sizeof(RecordHeader{}))
	return r
}

func (r RecordBuilder) size() int {
	return int(r.p.Uint16LE(0))
}

func (r RecordBuilder) Size(size uint16) RecordBuilder {
	r.p.SetUint16LE(0, size)
	return r
}

func (r RecordBuilder) TimeHF() RecordBuilder {
	r.p.SetInt64(24, timex.Now())
	return r
}

func (r RecordBuilder) Str(name string, value string) RecordBuilder {
	return r
}

func (r RecordBuilder) Strf(name string, value string, a ...any) RecordBuilder {
	return r
}

func (r RecordBuilder) Msg(msg string, a ...any) {
}

func (r RecordBuilder) Msgf(format string, a ...any) {
}

type RecordHeader struct {
	size      uint16 // size of entire message
	baseSize  uint16 // baseSize of header
	level     int8   // Level logger Level
	flags     Flags  //
	attrCount uint16 // attrCount number of attributes
	machineID uint64 // machineID unique ID of machine that generated the record
	seq       uint64 // seq is the machine specific Sequence number
	time      int64  // time nanos since epoch
	revision  slice
	source    slice
	caller    slice
	msg       slice
	attrs     slice
	_         [4]byte
}

func (r *RecordHeader) Source() string {
	return r.slice(r.source)
}

func (r *RecordHeader) Caller() string {
	return r.slice(r.caller)
}

func (r *RecordHeader) Msg() string {
	return r.slice(r.msg)
}

func (r *RecordHeader) slice(s slice) string {
	size := s.size
	if s.offset+s.size > r.size {
		if s.offset >= r.size {
			return ""
		}
		size = r.size - s.offset
	}
	return unsafeString(unsafe.Add(unsafe.Pointer(&r.size), s.offset), int(size))
}

//func (r *RecordHeader) Attr(attr RecordAttr) Attr {
//	key := r.slice(attr.name)
//	switch Kind(attr.kind) {
//	case KindInt64:
//		return Int64(key, *(*int64)(unsafe.Pointer(&attr.value)))
//	case KindUint64:
//		return Uint64(key, *(*uint64)(unsafe.Pointer(&attr.value)))
//	case KindFloat64:
//		return Float64(key, *(*float64)(unsafe.Pointer(&attr.value)))
//	case KindString:
//		return String(key, r.slice(attr.valueSlice()))
//	case KindBool:
//		return Bool(key, attr.value[0] != 0)
//	case KindDuration:
//		return Duration(key, *(*time.Duration)(unsafe.Pointer(&attr.value)))
//	case KindTime:
//		return Time(key, time.UnixMicro(*(*int64)(unsafe.Pointer(&attr.value))/1000))
//	}
//	return Any(key, nil)
//}

func (r *RecordHeader) AttrReader() {

}

type RecordAttributes struct {
	table [8]uint16
}

type slice struct {
	offset uint16
	size   uint16
}

type RecordAttr struct {
	name  slice
	kind  byte
	_     byte
	next  uint16
	value [8]byte
}

func (ra *RecordAttr) Kind() Kind {
	return Kind(ra.kind)
}

func (ra *RecordAttr) valueSlice() slice {
	return *(*slice)(unsafe.Pointer(&ra.value))
}

func unsafeString(p unsafe.Pointer, length int) string {
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(p),
		Len:  length,
	}))
}
