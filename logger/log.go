package logger

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"
	"unsafe"
)

var (
	DurationAsString   = true
	RawFieldName       = "raw"
	DataFieldName      = "data"
	DurationFieldName  = "dur"
	DurationsFieldName = "durs"
	ErrorsFieldName    = "errors"

	EmptyMessage = ""
)

func doLog(b []byte, ts int64, level int, fn *FuncInfo, args ...interface{}) []byte {
	// Builder?
	if len(args) == 0 {
		return b
	}

	b = binary.LittleEndian.AppendUint64(b, uint64(ts))
	b = binary.LittleEndian.AppendUint64(b, uint64(level))
	if fn != nil {
		b = append(b, fn.formatted...)
	}

	//// Treat as simple message?
	if len(args) == 1 {
		arg0 := args[0]
		msg, ok := arg0.(string)
		if !ok {
			msg = fmt.Sprintf("%s", arg0)
		}
		b = append(b, msg...)
		return b
	}

	switch t := args[0].(type) {
	// Handle error
	case error:
		_ = t
		args = args[1:]
	}

	for i := 0; i < len(args); i += 2 {
		//key := args[i]
		//if key == nil {
		//	// Shift by one
		//	i -= 1
		//	continue
		//}

		//switch k := key.(type) {
		//case string:
		//	// Treat it like a format template?
		//	//if strings.Contains(k, "%") {
		//	//	// The remaining args will be the format values
		//	//	//event.Msgf(k, args[i+1:]...)
		//	//	return b
		//	//}
		//
		//	valueIndex := i + 1
		//	// Treat key as message?
		//	if valueIndex == len(args) {
		//		//event.Msg(k)
		//		return b
		//	}
		//
		//	b = append(b, k...)
		//
		//	// Add to field map
		//	value := args[valueIndex]
		//	switch v := value.(type) {
		//	case string:
		//		b = append(b, v...)
		//	case time.Time:
		//		b = binary.LittleEndian.AppendUint64(b, uint64(v.UnixNano()))
		//	case *time.Time:
		//		b = binary.LittleEndian.AppendUint64(b, uint64(v.UnixNano()))
		//	case int:
		//		b = binary.LittleEndian.AppendUint64(b, uint64(v))
		//	case int8:
		//		b = append(b, byte(v))
		//	case int16:
		//		b = binary.LittleEndian.AppendUint16(b, uint16(v))
		//	case int32:
		//		b = binary.LittleEndian.AppendUint32(b, uint32(v))
		//	case int64:
		//		b = binary.LittleEndian.AppendUint64(b, uint64(v))
		//	case uint:
		//		b = binary.LittleEndian.AppendUint64(b, uint64(v))
		//	case uint8:
		//		b = append(b, v)
		//	case uint16:
		//		b = binary.LittleEndian.AppendUint16(b, v)
		//	case uint32:
		//		b = binary.LittleEndian.AppendUint32(b, v)
		//	case uint64:
		//		b = binary.LittleEndian.AppendUint64(b, v)
		//	case float32:
		//		b = binary.LittleEndian.AppendUint32(b, math.Float32bits(v))
		//	case float64:
		//		b = binary.LittleEndian.AppendUint64(b, math.Float64bits(v))
		//	case bool:
		//		if v {
		//			b = append(b, 1)
		//		} else {
		//			b = append(b, 0)
		//		}
		//	case error:
		//		b = append(b, v.Error()...)
		//	case time.Duration:
		//		if DurationAsString {
		//			b = binary.LittleEndian.AppendUint64(b, uint64(v))
		//		} else {
		//			b = binary.LittleEndian.AppendUint64(b, uint64(v))
		//		}
		//		//case json.Marshaler:
		//		//	bytes, err := v.MarshalJSON()
		//		//	if err != nil {
		//		//		event.AnErr(k, err)
		//		//	} else {
		//		//		event.RawJSON(k, bytes)
		//		//	}
		//		//case JSON:
		//		//	event.RawJSON(k, v)
		//		//case Builder:
		//		//	v(event)
		//		//default:
		//		//	appendInterface(event, k, v)
		//	}
		//	continue
		//
		//	//case error:
		//	//	event.Err(k)
		//	//case []error:
		//	//	event.Errs(ErrorsFieldName, k)
		//	//case time.Duration:
		//	//	if DurationAsString {
		//	//		event.Str(DurationFieldName, k.String())
		//	//	} else {
		//	//		event.Dur(DurationFieldName, k)
		//	//	}
		//	//case []time.Duration:
		//	//	event.Durs(DurationsFieldName, k)
		//	//case json.Marshaler:
		//	//	bytes, err := k.MarshalJSON()
		//	//	if err != nil {
		//	//		event.AnErr(DataFieldName, err)
		//	//	} else {
		//	//		event.RawJSON(DataFieldName, bytes)
		//	//	}
		//	//case []byte:
		//	//	event.Bytes(RawFieldName, k)
		//	//case JSON:
		//	//	event.RawJSON(DataFieldName, k)
		//	//case Builder:
		//	//	k(event)
		//	//case Request:
		//	//	appendInterface(event, RequestName, k)
		//	//case Response:
		//	//	appendInterface(event, ResponseName, k)
		//}
		i -= 1
	}
	return b
}

func SerializeBinary(b []byte, ts int64, level int, fn *FuncInfo, args ...Attr) []byte {
	// Builder?
	if len(args) == 0 {
		return b
	}

	b = binary.LittleEndian.AppendUint64(b, uint64(ts))
	b = binary.LittleEndian.AppendUint64(b, uint64(level))
	if fn != nil {
		b = append(b, fn.formatted...)
	}

	for _, attr := range args {
		switch attr.Value.Kind() {
		case KindString:
			b = append(b, attr.Value.String()...)
		case KindInt64:
			b = binary.LittleEndian.AppendUint64(b, uint64(attr.Value.Int64()))
		case KindUint64:
			b = binary.LittleEndian.AppendUint64(b, attr.Value.Uint64())
		case KindFloat64:
			b = binary.LittleEndian.AppendUint64(b, math.Float64bits(attr.Value.Float64()))
		}
	}

	return b
}

func SerializeJSON(b []byte, ts int64, level int, fn *FuncInfo, args ...Attr) []byte {
	// Builder?
	if len(args) == 0 {
		return b
	}

	b = append(b, '{')
	b = append(b, "\"time\":"...)
	b = strconv.AppendInt(b, ts, 10)
	b = append(b, ',')
	b = append(b, "\"level\":"...)
	b = strconv.AppendInt(b, int64(level), 10)
	if fn != nil {
		b = append(b, ',')
		b = append(b, "\"caller\":"...)
		b = append(b, '"')
		b = appendEscapedJSONString(b, fn.formatted)
		b = append(b, '"')
		b = append(b, ',')
		b = append(b, "\"func\":"...)
		b = append(b, '"')
		b = appendEscapedJSONString(b, fn.name)
		b = append(b, '"')

	}

	for _, attr := range args {
		b = append(b, ',')
		b = append(b, '"')
		b = appendEscapedJSONString(b, attr.Key)
		b = append(b, '"')
		b = append(b, ':')
		b = appendJSONValue(b, attr.Value)
		//switch attr.Value.Kind() {
		//case KindString:
		//	b = append(b, attr.Value.String()...)
		//case KindInt64:
		//	b = binary.LittleEndian.AppendUint64(b, uint64(attr.Value.Int64()))
		//case KindUint64:
		//	b = binary.LittleEndian.AppendUint64(b, attr.Value.Uint64())
		//case KindFloat64:
		//	b = binary.LittleEndian.AppendUint64(b, math.Float64bits(attr.Value.Float64()))
		//default:
		//	b = append(b, '0')
		//}
	}

	b = append(b, '}')

	return b
}

// Adapted from time.Time.MarshalJSON to avoid allocation.
func appendJSONTime(b []byte, t time.Time) []byte {
	if y := t.Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		//s.appendError(errors.New("time.Time year outside of range [0,9999]"))
	}
	b = append(b, '"')
	b = t.AppendFormat(b, time.RFC3339Nano)
	b = append(b, '"')
	return b
}

func appendJSONValue(b []byte, v Value) []byte {
	switch v.Kind() {
	case KindString:
		b = append(b, v.str()...)
	case KindInt64:
		b = strconv.AppendInt(b, v.Int64(), 10)
	case KindUint64:
		b = strconv.AppendUint(b, v.Uint64(), 10)
	case KindFloat64:
		//strconv.AppendFloat()
		// json.Marshal is funny about floats; it doesn't
		// always match strconv.AppendFloat. So just call it.
		// That's expensive, but floats are rare.
		//if err := appendJSONMarshal(s.buf, v.Float64()); err != nil {
		//	return err
		//}
	case KindBool:
		b = strconv.AppendBool(b, v.Bool())
	case KindDuration:
		// Do what json.Marshal does.
		b = strconv.AppendInt(b, int64(v.Duration()), 10)
	case KindTime:
		b = appendJSONTime(b, v.Time())
	case KindAny:
		a := v.Any()
		//_, jm := a.(json.Marshaler)
		if err, ok := a.(error); ok {
			b = append(b, '"')
			b = appendEscapedJSONString(b, err.Error())
			b = append(b, '"')
		} else {
			b = append(b, '0')
		}
		//	return appendJSONMarshal(s.buf, a)
		//}
	default:
		panic(fmt.Sprintf("bad kind: %s", v.Kind()))
	}
	return b
}

// appendEscapedJSONString escapes s for JSON and appends it to buf.
// It does not surround the string in quotation marks.
//
// Modified from encoding/json/encode.go:encodeState.string,
// with escapeHTML set to false.
func appendEscapedJSONString(buf []byte, s string) []byte {
	char := func(b byte) { buf = append(buf, b) }
	str := func(s string) { buf = append(buf, s...) }

	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if safeSet[b] {
				i++
				continue
			}
			if start < i {
				str(s[start:i])
			}
			char('\\')
			switch b {
			case '\\', '"':
				char(b)
			case '\n':
				char('n')
			case '\r':
				char('r')
			case '\t':
				char('t')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				str(`u00`)
				char(hex[b>>4])
				char(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				str(s[start:i])
			}
			str(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				str(s[start:i])
			}
			str(`\u202`)
			char(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		str(s[start:])
	}
	return buf
}

var hex = "0123456789abcdef"

// Copied from encoding/json/tables.go.
//
// safeSet holds the value true if the ASCII character with the given array
// position can be represented inside a JSON string without any further
// escaping.
//
// All values are true except for the ASCII control characters (0-31), the
// double quote ("), and the backslash character ("\").
var safeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}

func Log(args ...any) {
	gp := procPin()
	prevWriteBuf := gp.writebuf
	prevG0WriteBuf := gp.m.g0.writebuf
	gp.writebuf = *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(gp)),
		Len:  0,
		Cap:  0,
	}))
	gp.m.g0.writebuf = gp.writebuf
	systemstack(func() {
		this := getg()
		gpp := (*g)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.writebuf)).Data))
		this.writebuf = nil
		var (
			sp0 uintptr
			sp  uintptr
		)
		if gpp.syscallsp != 0 {
			sp0 = gpp.syscallsp
		} else {
			sp0 = gpp.sched.sp
		}
		if usesLR {
			sp = sp0 + callerSPOffset
		} else {
			sp = sp0 + callerSPOffset
		}
		(*reflect.SliceHeader)(unsafe.Pointer(&gpp.writebuf)).Data = sp
	})
	sp := *(*uintptr)(unsafe.Pointer(&gp.writebuf))
	gp.writebuf = prevWriteBuf
	gp.m.g0.writebuf = prevG0WriteBuf
	pc := *(*uintptr)(unsafe.Pointer(sp))
	procUnpinGp(gp)

	f := funcInfoMap.GetForPC(pc)
	_ = f
	b := gp.writebuf
	if b == nil {
		b = make([]byte, 0, 512)
		println("allocated")
	}

	//b = doLog(b, timex.NanoTime(), 0, f, args...)
	b = doLog(b, 0, 0, f, args...)
	//b = doLog(b, 0, 0, nil, args...)
	gp.writebuf = b[:0]
}

func Slog(msg string, args ...Attr) {
	gp := procPin()
	prevWriteBuf := gp.writebuf
	prevG0WriteBuf := gp.m.g0.writebuf
	gp.writebuf = *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(gp)),
		Len:  0,
		Cap:  0,
	}))
	gp.m.g0.writebuf = gp.writebuf
	systemstack(func() {
		this := getg()
		gpp := (*g)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.writebuf)).Data))
		this.writebuf = nil
		var (
			sp0 uintptr
			sp  uintptr
		)
		if gpp.syscallsp != 0 {
			sp0 = gpp.syscallsp
		} else {
			sp0 = gpp.sched.sp
		}
		if usesLR {
			sp = sp0 + callerSPOffset
		} else {
			sp = sp0 + callerSPOffset
		}
		(*reflect.SliceHeader)(unsafe.Pointer(&gpp.writebuf)).Data = sp
	})
	sp := *(*uintptr)(unsafe.Pointer(&gp.writebuf))
	gp.writebuf = prevWriteBuf
	gp.m.g0.writebuf = prevG0WriteBuf
	pc := *(*uintptr)(unsafe.Pointer(sp))
	procUnpinGp(gp)

	f := funcInfoMap.GetForPC(pc)
	_ = f
	b := gp.writebuf
	if b == nil {
		b = make([]byte, 0, 512)
	}

	//p := memory.Alloc(512)
	//b := p.Bytes(0, 512, 512)
	//b = b[:0-]
	//defer memory.Free(p)
	//b := pool.Alloc(512)[:0]
	//defer pool.Free(b[:])

	//b = SerializeBinary(b, timex.Now(), 1, f, args...)
	b = SerializeBinary(b, 0, 1, f, args...)
	//b = SerializeJSON(b, timex.Now(), 1, f, args...)
	//logger.LogAttrs(context.Background(), slog.LevelInfo, msg, args...)
	//b = doLog(b, timex.NanoTime(), 0, f, args...)
	//b = doLog(b, 0, 0, nil, args...)
	gp.writebuf = b[:0]
}

const TestMessage = "Test logging, but use a somewhat realistic message length."

var (
	TestTime     = time.Date(2022, time.May, 1, 0, 0, 0, 0, time.UTC)
	TestString   = "7e3b3b2aaeff56a7108fe11e154200dd/7819479873059528190"
	TestInt      = 32768
	TestDuration = 23 * time.Second
	TestError    = errors.New("fail")
)
