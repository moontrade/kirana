package logger

import (
	"math"
	"strconv"
	"time"
	"unicode/utf8"
)

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

//func appendJSONValue(b []byte, v Value) []byte {
//	switch v.Kind() {
//	case KindString:
//		b = append(b, v.str()...)
//	case KindInt64:
//		b = strconv.AppendInt(b, v.Int64(), 10)
//	case KindUint64:
//		b = strconv.AppendUint(b, v.Uint64(), 10)
//	case KindFloat64:
//		b = appendJSONFloat64(b, v.Float64(), 64)
//	case KindBool:
//		b = strconv.AppendBool(b, v.Bool())
//	case KindDuration:
//		// Do what json.Marshal does.
//		b = strconv.AppendInt(b, int64(v.Duration()), 10)
//	case KindTime:
//		b = appendJSONTime(b, v.Time())
//	case KindAny:
//		a := v.Any()
//		//_, jm := a.(json.Marshaler)
//		if err, ok := a.(error); ok {
//			b = append(b, '"')
//			b = appendEscapedJSONString(b, err.Error())
//			b = append(b, '"')
//		} else {
//			b = append(b, '0')
//		}
//		//	return appendJSONMarshal(s.buf, a)
//		//}
//	default:
//		panic(fmt.Sprintf("bad kind: %s", v.Kind()))
//	}
//	return b
//}

func appendJSONFloat64(b []byte, f float64, bits int) []byte {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return append(b, '0', '.', '0')
	}

	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	// See golang.org/issue/6384 and golang.org/issue/14135.
	// Like format %g, but the exponent cutoffs are different
	// and exponents themselves are not padded to two digits.
	abs := math.Abs(f)
	format := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 {
		if bits == 64 && (abs < 1e-6 || abs >= 1e21) || bits == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
			format = 'e'
		}
	}
	b = strconv.AppendFloat(b, f, format, -1, bits)
	if format == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
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
