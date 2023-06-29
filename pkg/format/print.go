package format

import (
	"github.com/moontrade/kirana/pkg/format/internal/fmtsort"
	"io"
	"os"
	"reflect"
	"strconv"
	"unicode/utf8"
)

const (
	ldigits = "0123456789abcdefx"
	udigits = "0123456789ABCDEFX"
)

const (
	signed   = true
	unsigned = false
)

// flags placed in a separate struct for easy clearing.
type fmtFlags struct {
	widPresent  bool
	precPresent bool
	minus       bool
	plus        bool
	sharp       bool
	space       bool
	zero        bool

	// For the formats %+v %#v, we set the plusV/sharpV flags
	// and clear the plus/sharp flags since %+v and %#v are in effect
	// different, flagless formats set at the top level.
	plusV  bool
	sharpV bool
}

// Strings for use with buffer.WriteString.
// This is less overhead than using buffer.Write with byte arrays.
const (
	commaSpaceString  = ", "
	nilAngleString    = "<nil>"
	nilParenString    = "(nil)"
	nilString         = "nil"
	mapString         = "map["
	percentBangString = "%!"
	missingString     = "(MISSING)"
	badIndexString    = "(BADINDEX)"
	panicString       = "(PANIC="
	extraString       = "%!(EXTRA "
	badWidthString    = "%!(BADWIDTH)"
	badPrecString     = "%!(BADPREC)"
	noVerbString      = "%!(NOVERB)"
	invReflectString  = "<invalid reflect.Value>"
)

// Use simple []byte instead of bytes.Buffer to avoid large dependency.
type buffer []byte

func (b *buffer) write(p []byte) {
	*b = append(*b, p...)
}

func (b *buffer) writeString(s string) {
	*b = append(*b, s...)
}

func (b *buffer) writeByte(c byte) {
	*b = append(*b, c)
}

func (bp *buffer) writeRune(r rune) {
	*bp = utf8.AppendRune(*bp, r)
}

// pp is used to store a printer's state and is reused with sync.Pool to avoid allocations.
type pp struct {
	buf buffer

	// arg holds the current item, as an interface{}.
	//arg any

	// value is used instead of arg for reflect values.
	value reflect.Value

	// fmt is used to format basic items such as integers or strings.
	//fmt
	fmtFlags

	wid  int // width
	prec int // precision

	// intbuf is large enough to store %b of an int64 with a sign and
	// avoids padding at the end of the struct on 32 bit architectures.
	intbuf [68]byte

	// reordered records whether the format string used argument reordering.
	reordered bool
	// goodArgNum records whether the most recent reordering directive was valid.
	goodArgNum bool
	// panicking is set by catchPanic to avoid infinite panic, recover, panic, ... recursion.
	panicking bool
	// erroring is set when printing an error string to guard against calling handleMethods.
	erroring bool
	// wrapErrs is set when the format string may contain a %w verb.
	wrapErrs bool
	// wrappedErrs records the targets of the %w verb.
	wrappedErrs []int
}

// newPrinter allocates a new pp struct or grabs a cached one.
//func newPrinter() pp {
//	p := pp{}
//	return p
//}

func (p *pp) Width() (wid int, ok bool) { return p.wid, p.widPresent }

func (p *pp) Precision() (prec int, ok bool) { return p.prec, p.precPresent }

func (p *pp) Flag(b int) bool {
	switch b {
	case '-':
		return p.minus
	case '+':
		return p.plus || p.plusV
	case '#':
		return p.sharp || p.sharpV
	case ' ':
		return p.space
	case '0':
		return p.zero
	}
	return false
}

// Implement Write so we can call Fprintf on a pp (through State), for
// recursive use in custom verbs.
func (p *pp) Write(b []byte) (ret int, err error) {
	p.buf.write(b)
	return len(b), nil
}

// Implement WriteString so that we can call io.WriteString
// on a pp (through state), for efficiency.
func (p *pp) WriteString(s string) (ret int, err error) {
	p.buf.writeString(s)
	return len(s), nil
}

// These routines end in 'f' and take a format string.

// Fprintf formats according to a format specifier and writes to w.
// It returns the number of bytes written and any write error encountered.
func Fprintf(w io.Writer, format string, a ...any) (n int, err error) {
	p := pp{}
	p.doPrintf(format, a)
	n, err = w.Write(p.buf)
	return
}

// Printf formats according to a format specifier and writes to standard output.
// It returns the number of bytes written and any write error encountered.
func Printf(format string, a ...any) (n int, err error) {
	return Fprintf(os.Stdout, format, a...)
}

func Printlnf(format string, a ...any) (n int, err error) {
	n, err = Fprintf(os.Stdout, format, a...)
	if err != nil {
		return n, err
	}
	var n2 int
	n2, err = os.Stdout.WriteString("\n")
	return n + n2, err
}

// Sprintf formats according to a format specifier and returns the resulting string.
func Sprintf(format string, a ...any) string {
	p := pp{}
	p.doPrintf(format, a)
	s := string(p.buf)
	return s
}

// Appendf formats according to a format specifier, appends the result to the byte
// slice, and returns the updated slice.
func Appendf(b []byte, format string, a ...any) []byte {
	p := pp{}
	p.buf = b
	p.doPrintf(format, a)
	return p.buf
}

// These routines do not take a format string

// Fprint formats using the default formats for its operands and writes to w.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
func Fprint(w io.Writer, a ...any) (n int, err error) {
	p := pp{}
	p.doPrint(a)
	n, err = w.Write(p.buf)
	return
}

// Print formats using the default formats for its operands and writes to standard output.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
func Print(a ...any) (n int, err error) {
	return Fprint(os.Stdout, a...)
}

// Sprint formats using the default formats for its operands and returns the resulting string.
// Spaces are added between operands when neither is a string.
func Sprint(a ...any) string {
	p := pp{}
	p.doPrint(a)
	s := string(p.buf)
	return s
}

// Append formats using the default formats for its operands, appends the result to
// the byte slice, and returns the updated slice.
func Append(b []byte, a ...any) []byte {
	p := pp{}
	p.buf = b
	p.doPrint(a)
	b = append(b, p.buf...)
	return b
}

// These routines end in 'ln', do not take a format string,
// always add spaces between operands, and add a newline
// after the last operand.

// Fprintln formats using the default formats for its operands and writes to w.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func Fprintln(w io.Writer, a ...any) (n int, err error) {
	p := pp{}
	p.doPrintln(a)
	n, err = w.Write(p.buf)
	return
}

// Println formats using the default formats for its operands and writes to standard output.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func Println(a ...any) (n int, err error) {
	return Fprintln(os.Stdout, a...)
}

// Sprintln formats using the default formats for its operands and returns the resulting string.
// Spaces are always added between operands and a newline is appended.
func Sprintln(a ...any) string {
	p := pp{}
	p.doPrintln(a)
	s := string(p.buf)
	return s
}

// Appendln formats using the default formats for its operands, appends the result
// to the byte slice, and returns the updated slice. Spaces are always added
// between operands and a newline is appended.
func Appendln(b []byte, a ...any) []byte {
	p := pp{}
	p.doPrintln(a)
	b = append(b, p.buf...)
	return b
}

// getField gets the i'th field of the struct value.
// If the field is itself is an interface, return a value for
// the thing inside the interface, not the interface itself.
func getField(v reflect.Value, i int) reflect.Value {
	val := v.Field(i)
	if val.Kind() == reflect.Interface && !val.IsNil() {
		val = val.Elem()
	}
	return val
}

// tooLarge reports whether the magnitude of the integer is
// too large to be used as a formatting width or precision.
func tooLarge(x int) bool {
	const max int = 1e6
	return x > max || x < -max
}

// parsenum converts ASCII to integer.  num is 0 (and isnum is false) if no number present.
func parsenum(s string, start, end int) (num int, isnum bool, newi int) {
	if start >= end {
		return 0, false, end
	}
	for newi = start; newi < end && '0' <= s[newi] && s[newi] <= '9'; newi++ {
		if tooLarge(num) {
			return 0, false, end // Overflow; crazy long number most likely.
		}
		num = num*10 + int(s[newi]-'0')
		isnum = true
	}
	return
}

func (p *pp) unknownType(v reflect.Value) {
	if !v.IsValid() {
		p.buf.writeString(nilAngleString)
		return
	}
	p.buf.writeByte('?')
	p.buf.writeString(v.Type().String())
	p.buf.writeByte('?')
}

func (p *pp) badVerb(verb rune) {
	p.erroring = true
	p.buf.writeString(percentBangString)
	p.buf.writeRune(verb)
	p.buf.writeByte('(')
	switch {
	//case p.arg != nil:
	//	p.buf.writeString(reflect.TypeOf(p.arg).String())
	//	p.buf.writeByte('=')
	//	p.printArg(p.arg, 'v')
	case p.value.IsValid():
		p.buf.writeString(p.value.Type().String())
		p.buf.writeByte('=')
		p.printValue(p.value, 'v', 0)
	default:
		p.buf.writeString(nilAngleString)
	}
	p.buf.writeByte(')')
	p.erroring = false
}

func (p *pp) fmtBool(v bool, verb rune) {
	switch verb {
	case 't', 'v':
		p._fmtBoolean(v)
	default:
		p.badVerb(verb)
	}
}

// fmt0x64 formats a uint64 in hexadecimal and prefixes it with 0x or
// not, as requested, by temporarily setting the sharp flag.
func (p *pp) fmt0x64(v uint64, leading0x bool) {
	sharp := p.sharp
	p.sharp = leading0x
	p._fmtInteger(v, 16, unsigned, 'v', ldigits)
	p.sharp = sharp
}

// fmtInteger formats a signed or unsigned integer.
func (p *pp) fmtInteger(v uint64, isSigned bool, verb rune) {
	switch verb {
	case 'v':
		if p.sharpV && !isSigned {
			p.fmt0x64(v, true)
		} else {
			p._fmtInteger(v, 10, isSigned, verb, ldigits)
		}
	case 'd':
		p._fmtInteger(v, 10, isSigned, verb, ldigits)
	case 'b':
		p._fmtInteger(v, 2, isSigned, verb, ldigits)
	case 'o', 'O':
		p._fmtInteger(v, 8, isSigned, verb, ldigits)
	case 'x':
		p._fmtInteger(v, 16, isSigned, verb, ldigits)
	case 'X':
		p._fmtInteger(v, 16, isSigned, verb, udigits)
	case 'c':
		p._fmtC(v)
	case 'q':
		p._fmtQc(v)
	case 'U':
		p._fmtUnicode(v)
	default:
		p.badVerb(verb)
	}
}

// fmtFloat formats a float. The default precision for each verb
// is specified as last argument in the call to fmt_float.
func (p *pp) fmtFloat(v float64, size int, verb rune) {
	switch verb {
	case 'v':
		p._fmtFloat(v, size, 'g', -1)
	case 'b', 'g', 'G', 'x', 'X':
		p._fmtFloat(v, size, verb, -1)
	case 'f', 'e', 'E':
		p._fmtFloat(v, size, verb, 6)
	case 'F':
		p._fmtFloat(v, size, 'f', 6)
	default:
		p.badVerb(verb)
	}
}

// fmtComplex formats a complex number v with
// r = real(v) and j = imag(v) as (r+ji) using
// fmtFloat for r and j formatting.
func (p *pp) fmtComplex(v complex128, size int, verb rune) {
	// Make sure any unsupported verbs are found before the
	// calls to _fmtFloat to not generate an incorrect error string.
	switch verb {
	case 'v', 'b', 'g', 'G', 'x', 'X', 'f', 'F', 'e', 'E':
		oldPlus := p.plus
		p.buf.writeByte('(')
		p.fmtFloat(real(v), size/2, verb)
		// Imaginary part always has a sign.
		p.plus = true
		p.fmtFloat(imag(v), size/2, verb)
		p.buf.writeString("i)")
		p.plus = oldPlus
	default:
		p.badVerb(verb)
	}
}

func (p *pp) fmtString(v string, verb rune) {
	switch verb {
	case 'v':
		if p.sharpV {
			p._fmtQ(v)
		} else {
			p._fmtS(v)
		}
	case 's':
		p._fmtS(v)
	case 'x':
		p._fmtSx(v, ldigits)
	case 'X':
		p._fmtSx(v, udigits)
	case 'q':
		p._fmtQ(v)
	default:
		p.badVerb(verb)
	}
}

func (p *pp) fmtBytes(v []byte, verb rune, typeString string) {
	switch verb {
	case 'v', 'd':
		if p.sharpV {
			p.buf.writeString(typeString)
			if v == nil {
				p.buf.writeString(nilParenString)
				return
			}
			p.buf.writeByte('{')
			for i, c := range v {
				if i > 0 {
					p.buf.writeString(commaSpaceString)
				}
				p.fmt0x64(uint64(c), true)
			}
			p.buf.writeByte('}')
		} else {
			p.buf.writeByte('[')
			for i, c := range v {
				if i > 0 {
					p.buf.writeByte(' ')
				}
				p._fmtInteger(uint64(c), 10, unsigned, verb, ldigits)
			}
			p.buf.writeByte(']')
		}
	case 's':
		p._fmtBs(v)
	case 'x':
		p._fmtBx(v, ldigits)
	case 'X':
		p._fmtBx(v, udigits)
	case 'q':
		p._fmtQ(string(v))
	default:
		p.printValue(reflect.ValueOf(v), verb, 0)
	}
}

func (p *pp) fmtPointer(value reflect.Value, verb rune) {
	var u uintptr
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		u = value.Pointer()
	default:
		p.badVerb(verb)
		return
	}

	switch verb {
	case 'v':
		if p.sharpV {
			p.buf.writeByte('(')
			p.buf.writeString(value.Type().String())
			p.buf.writeString(")(")
			if u == 0 {
				p.buf.writeString(nilString)
			} else {
				p.fmt0x64(uint64(u), true)
			}
			p.buf.writeByte(')')
		} else {
			if u == 0 {
				p._padString(nilAngleString)
			} else {
				p.fmt0x64(uint64(u), !p.sharp)
			}
		}
	case 'p':
		p.fmt0x64(uint64(u), !p.sharp)
	case 'b', 'o', 'd', 'x', 'X':
		p.fmtInteger(uint64(u), unsigned, verb)
	default:
		p.badVerb(verb)
	}
}

func (p *pp) catchPanic(arg any, verb rune, method string) {
	if err := recover(); err != nil {
		// If it's a nil pointer, just say "<nil>". The likeliest causes are a
		// Stringer that fails to guard against nil or a nil pointer for a
		// value receiver, and in either case, "<nil>" is a nice result.
		if v := reflect.ValueOf(arg); v.Kind() == reflect.Pointer && v.IsNil() {
			p.buf.writeString(nilAngleString)
			return
		}
		// Otherwise print a concise panic message. Most of the time the panic
		// value will print itself nicely.
		if p.panicking {
			// Nested panics; the recursion in printArg cannot succeed.
			panic(err)
		}

		oldFlags := p.fmtFlags
		// For this output we want default behavior.
		p._clearflags()

		p.buf.writeString(percentBangString)
		p.buf.writeRune(verb)
		p.buf.writeString(panicString)
		p.buf.writeString(method)
		p.buf.writeString(" method: ")
		p.panicking = true
		p.printArg(err, 'v')
		p.panicking = false
		p.buf.writeByte(')')

		p.fmtFlags = oldFlags
	}
}

func (p *pp) handleMethods(verb rune) (handled bool) {
	//if p.erroring {
	//	return
	//}
	//if verb == 'w' {
	//	// It is invalid to use %w other than with Errorf or with a non-error arg.
	//	_, ok := arg.(error)
	//	if !ok || !p.wrapErrs {
	//		p.badVerb(verb)
	//		return true
	//	}
	//	// If the arg is a Formatter, pass 'v' as the verb to it.
	//	verb = 'v'
	//}
	//
	//// Is it a Formatter?
	//if formatter, ok := arg.(fmt.Formatter); ok {
	//	handled = true
	//	defer p.catchPanic(arg, verb, "Format")
	//	formatter.Format(p, verb)
	//	return
	//}
	//
	//// If we're doing Go syntax and the argument knows how to supply it, take care of it now.
	//if p.sharpV {
	//	if stringer, ok := arg.(fmt.GoStringer); ok {
	//		handled = true
	//		defer p.catchPanic(arg, verb, "GoString")
	//		// Print the result of GoString unadorned.
	//		p._fmtS(stringer.GoString())
	//		return
	//	}
	//} else {
	//	// If a string is acceptable according to the format, see if
	//	// the value satisfies one of the string-valued interfaces.
	//	// Println etc. set verb to %v, which is "stringable".
	//	switch verb {
	//	case 'v', 's', 'x', 'X', 'q':
	//		// Is it an error or Stringer?
	//		// The duplication in the bodies is necessary:
	//		// setting handled and deferring catchPanic
	//		// must happen before calling the method.
	//		switch v := arg.(type) {
	//		case error:
	//			handled = true
	//			defer p.catchPanic(arg, verb, "Error")
	//			p.fmtString(v.Error(), verb)
	//			return
	//
	//		case fmt.Stringer:
	//			handled = true
	//			defer p.catchPanic(arg, verb, "String")
	//			p.fmtString(v.String(), verb)
	//			return
	//		}
	//	}
	//}
	return false
}

func (p *pp) printArg(arg any, verb rune) {
	//p.arg = arg
	p.value = reflect.Value{}

	if arg == nil {
		switch verb {
		case 'T', 'v':
			p._padString(nilAngleString)
		default:
			p.badVerb(verb)
		}
		return
	}

	// Special processing considerations.
	// %T (the value's type) and %p (its address) are special; we always do them first.
	switch verb {
	case 'T':
		p._fmtS(reflect.TypeOf(arg).String())
		return
	case 'p':
		p.fmtPointer(reflect.ValueOf(arg), 'p')
		return
	}

	// Some types can be done without reflection.
	switch f := arg.(type) {
	case bool:
		p.fmtBool(f, verb)
	case float32:
		p.fmtFloat(float64(f), 32, verb)
	case float64:
		p.fmtFloat(f, 64, verb)
	case complex64:
		p.fmtComplex(complex128(f), 64, verb)
	case complex128:
		p.fmtComplex(f, 128, verb)
	case int:
		p.fmtInteger(uint64(f), signed, verb)
	case int8:
		p.fmtInteger(uint64(f), signed, verb)
	case int16:
		p.fmtInteger(uint64(f), signed, verb)
	case int32:
		p.fmtInteger(uint64(f), signed, verb)
	case int64:
		p.fmtInteger(uint64(f), signed, verb)
	case uint:
		p.fmtInteger(uint64(f), unsigned, verb)
	case uint8:
		p.fmtInteger(uint64(f), unsigned, verb)
	case uint16:
		p.fmtInteger(uint64(f), unsigned, verb)
	case uint32:
		p.fmtInteger(uint64(f), unsigned, verb)
	case uint64:
		p.fmtInteger(f, unsigned, verb)
	case uintptr:
		p.fmtInteger(uint64(f), unsigned, verb)
	case string:
		p.fmtString(f, verb)
	case []byte:
		p.fmtBytes(f, verb, "[]byte")
	case reflect.Value:
		// Handle extractable values with special methods
		// since printValue does not handle them at depth 0.
		if f.IsValid() && f.CanInterface() {
			//p.arg = f.Interface()
			if p.handleMethods(verb) {
				return
			}
		}
		p.printValue(f, verb, 0)
	default:
		// If the type is not simple, it might have methods.
		if !p.handleMethods(verb) {
			// Need to use reflection, since the type had no
			// interface methods that could be used for formatting.
			p.printValue(reflect.ValueOf(f), verb, 0)
		}
	}
}

// printValue is similar to printArg but starts with a reflect value, not an interface{} value.
// It does not handle 'p' and 'T' verbs because these should have been already handled by printArg.
func (p *pp) printValue(value reflect.Value, verb rune, depth int) {
	// Handle values with special methods if not already handled by printArg (depth == 0).
	if depth > 0 && value.IsValid() && value.CanInterface() {
		//arg := value.Interface()
		if p.handleMethods(verb) {
			return
		}
	}
	//p.arg = nil
	p.value = value

	switch f := value; value.Kind() {
	case reflect.Invalid:
		if depth == 0 {
			p.buf.writeString(invReflectString)
		} else {
			switch verb {
			case 'v':
				p.buf.writeString(nilAngleString)
			default:
				p.badVerb(verb)
			}
		}
	case reflect.Bool:
		p.fmtBool(f.Bool(), verb)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p.fmtInteger(uint64(f.Int()), signed, verb)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p.fmtInteger(f.Uint(), unsigned, verb)
	case reflect.Float32:
		p.fmtFloat(f.Float(), 32, verb)
	case reflect.Float64:
		p.fmtFloat(f.Float(), 64, verb)
	case reflect.Complex64:
		p.fmtComplex(f.Complex(), 64, verb)
	case reflect.Complex128:
		p.fmtComplex(f.Complex(), 128, verb)
	case reflect.String:
		p.fmtString(f.String(), verb)
	case reflect.Map:
		if p.sharpV {
			p.buf.writeString(f.Type().String())
			if f.IsNil() {
				p.buf.writeString(nilParenString)
				return
			}
			p.buf.writeByte('{')
		} else {
			p.buf.writeString(mapString)
		}
		sorted := fmtsort.Sort(f)
		for i, key := range sorted.Key {
			if i > 0 {
				if p.sharpV {
					p.buf.writeString(commaSpaceString)
				} else {
					p.buf.writeByte(' ')
				}
			}
			p.printValue(key, verb, depth+1)
			p.buf.writeByte(':')
			p.printValue(sorted.Value[i], verb, depth+1)
		}
		if p.sharpV {
			p.buf.writeByte('}')
		} else {
			p.buf.writeByte(']')
		}
	case reflect.Struct:
		if p.sharpV {
			p.buf.writeString(f.Type().String())
		}
		p.buf.writeByte('{')
		for i := 0; i < f.NumField(); i++ {
			if i > 0 {
				if p.sharpV {
					p.buf.writeString(commaSpaceString)
				} else {
					p.buf.writeByte(' ')
				}
			}
			if p.plusV || p.sharpV {
				if name := f.Type().Field(i).Name; name != "" {
					p.buf.writeString(name)
					p.buf.writeByte(':')
				}
			}
			p.printValue(getField(f, i), verb, depth+1)
		}
		p.buf.writeByte('}')
	case reflect.Interface:
		value := f.Elem()
		if !value.IsValid() {
			if p.sharpV {
				p.buf.writeString(f.Type().String())
				p.buf.writeString(nilParenString)
			} else {
				p.buf.writeString(nilAngleString)
			}
		} else {
			p.printValue(value, verb, depth+1)
		}
	case reflect.Array, reflect.Slice:
		switch verb {
		case 's', 'q', 'x', 'X':
			// Handle byte and uint8 slices and arrays special for the above verbs.
			t := f.Type()
			if t.Elem().Kind() == reflect.Uint8 {
				var bytes []byte
				if f.Kind() == reflect.Slice {
					bytes = f.Bytes()
				} else if f.CanAddr() {
					bytes = f.Slice(0, f.Len()).Bytes()
				} else {
					// We have an array, but we cannot Slice() a non-addressable array,
					// so we build a slice by hand. This is a rare case but it would be nice
					// if reflection could help a little more.
					bytes = make([]byte, f.Len())
					for i := range bytes {
						bytes[i] = byte(f.Index(i).Uint())
					}
				}
				p.fmtBytes(bytes, verb, t.String())
				return
			}
		}
		if p.sharpV {
			p.buf.writeString(f.Type().String())
			if f.Kind() == reflect.Slice && f.IsNil() {
				p.buf.writeString(nilParenString)
				return
			}
			p.buf.writeByte('{')
			for i := 0; i < f.Len(); i++ {
				if i > 0 {
					p.buf.writeString(commaSpaceString)
				}
				p.printValue(f.Index(i), verb, depth+1)
			}
			p.buf.writeByte('}')
		} else {
			p.buf.writeByte('[')
			for i := 0; i < f.Len(); i++ {
				if i > 0 {
					p.buf.writeByte(' ')
				}
				p.printValue(f.Index(i), verb, depth+1)
			}
			p.buf.writeByte(']')
		}
	case reflect.Pointer:
		// pointer to array or slice or struct? ok at top level
		// but not embedded (avoid loops)
		if depth == 0 && f.Pointer() != 0 {
			switch a := f.Elem(); a.Kind() {
			case reflect.Array, reflect.Slice, reflect.Struct, reflect.Map:
				p.buf.writeByte('&')
				p.printValue(a, verb, depth+1)
				return
			}
		}
		fallthrough
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		p.fmtPointer(f, verb)
	default:
		p.unknownType(f)
	}
}

// intFromArg gets the argNumth element of a. On return, isInt reports whether the argument has integer type.
func intFromArg(a []any, argNum int) (num int, isInt bool, newArgNum int) {
	newArgNum = argNum
	if argNum < len(a) {
		num, isInt = a[argNum].(int) // Almost always OK.
		if !isInt {
			// Work harder.
			switch v := reflect.ValueOf(a[argNum]); v.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				n := v.Int()
				if int64(int(n)) == n {
					num = int(n)
					isInt = true
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				n := v.Uint()
				if int64(n) >= 0 && uint64(int(n)) == n {
					num = int(n)
					isInt = true
				}
			default:
				// Already 0, false.
			}
		}
		newArgNum = argNum + 1
		if tooLarge(num) {
			num = 0
			isInt = false
		}
	}
	return
}

// parseArgNumber returns the value of the bracketed number, minus 1
// (explicit argument numbers are one-indexed but we want zero-indexed).
// The opening bracket is known to be present at format[0].
// The returned values are the index, the number of bytes to consume
// up to the closing paren, if present, and whether the number parsed
// ok. The bytes to consume will be 1 if no closing paren is present.
func parseArgNumber(format string) (index int, wid int, ok bool) {
	// There must be at least 3 bytes: [n].
	if len(format) < 3 {
		return 0, 1, false
	}

	// Find closing bracket.
	for i := 1; i < len(format); i++ {
		if format[i] == ']' {
			width, ok, newi := parsenum(format, 1, i)
			if !ok || newi != i {
				return 0, i + 1, false
			}
			return width - 1, i + 1, true // arg numbers are one-indexed and skip paren.
		}
	}
	return 0, 1, false
}

// argNumber returns the next argument to evaluate, which is either the value of the passed-in
// argNum or the value of the bracketed integer that begins format[i:]. It also returns
// the new value of i, that is, the index of the next byte of the format to process.
func (p *pp) argNumber(argNum int, format string, i int, numArgs int) (newArgNum, newi int, found bool) {
	if len(format) <= i || format[i] != '[' {
		return argNum, i, false
	}
	p.reordered = true
	index, wid, ok := parseArgNumber(format[i:])
	if ok && 0 <= index && index < numArgs {
		return index, i + wid, true
	}
	p.goodArgNum = false
	return argNum, i + wid, ok
}

func (p *pp) badArgNum(verb rune) {
	p.buf.writeString(percentBangString)
	p.buf.writeRune(verb)
	p.buf.writeString(badIndexString)
}

func (p *pp) missingArg(verb rune) {
	p.buf.writeString(percentBangString)
	p.buf.writeRune(verb)
	p.buf.writeString(missingString)
}

func (p *pp) doPrintf(format string, a []any) {
	end := len(format)
	argNum := 0         // we process one argument per non-trivial format
	afterIndex := false // previous item in format was an index like [3].
	p.reordered = false
formatLoop:
	for i := 0; i < end; {
		p.goodArgNum = true
		lasti := i
		for i < end && format[i] != '%' {
			i++
		}
		if i > lasti {
			p.buf.writeString(format[lasti:i])
		}
		if i >= end {
			// done processing format string
			break
		}

		// Process one verb
		i++

		// Do we have flags?
		p._clearflags()
	simpleFormat:
		for ; i < end; i++ {
			c := format[i]
			switch c {
			case '#':
				p.sharp = true
			case '0':
				p.zero = !p.minus // Only allow zero padding to the left.
			case '+':
				p.plus = true
			case '-':
				p.minus = true
				p.zero = false // Do not pad with zeros to the right.
			case ' ':
				p.space = true
			default:
				// Fast path for common case of ascii lower case simple verbs
				// without precision or width or argument indices.
				if 'a' <= c && c <= 'z' && argNum < len(a) {
					switch c {
					case 'w':
						p.wrappedErrs = append(p.wrappedErrs, argNum)
						fallthrough
					case 'v':
						// Go syntax
						p.sharpV = p.sharp
						p.sharp = false
						// Struct-field syntax
						p.plusV = p.plus
						p.plus = false
					}
					p.printArg(a[argNum], rune(c))
					argNum++
					i++
					continue formatLoop
				}
				// Format is more complex than simple flags and a verb or is malformed.
				break simpleFormat
			}
		}

		// Do we have an explicit argument index?
		argNum, i, afterIndex = p.argNumber(argNum, format, i, len(a))

		// Do we have width?
		if i < end && format[i] == '*' {
			i++
			p.wid, p.widPresent, argNum = intFromArg(a, argNum)

			if !p.widPresent {
				p.buf.writeString(badWidthString)
			}

			// We have a negative width, so take its value and ensure
			// that the minus flag is set
			if p.wid < 0 {
				p.wid = -p.wid
				p.minus = true
				p.zero = false // Do not pad with zeros to the right.
			}
			afterIndex = false
		} else {
			p.wid, p.widPresent, i = parsenum(format, i, end)
			if afterIndex && p.widPresent { // "%[3]2d"
				p.goodArgNum = false
			}
		}

		// Do we have precision?
		if i+1 < end && format[i] == '.' {
			i++
			if afterIndex { // "%[3].2d"
				p.goodArgNum = false
			}
			argNum, i, afterIndex = p.argNumber(argNum, format, i, len(a))
			if i < end && format[i] == '*' {
				i++
				p.prec, p.precPresent, argNum = intFromArg(a, argNum)
				// Negative precision arguments don't make sense
				if p.prec < 0 {
					p.prec = 0
					p.precPresent = false
				}
				if !p.precPresent {
					p.buf.writeString(badPrecString)
				}
				afterIndex = false
			} else {
				p.prec, p.precPresent, i = parsenum(format, i, end)
				if !p.precPresent {
					p.prec = 0
					p.precPresent = true
				}
			}
		}

		if !afterIndex {
			argNum, i, afterIndex = p.argNumber(argNum, format, i, len(a))
		}

		if i >= end {
			p.buf.writeString(noVerbString)
			break
		}

		verb, size := rune(format[i]), 1
		if verb >= utf8.RuneSelf {
			verb, size = utf8.DecodeRuneInString(format[i:])
		}
		i += size

		switch {
		case verb == '%': // Percent does not absorb operands and ignores f.wid and f.prec.
			p.buf.writeByte('%')
		case !p.goodArgNum:
			p.badArgNum(verb)
		case argNum >= len(a): // No argument left over to print for the current verb.
			p.missingArg(verb)
		case verb == 'w':
			p.wrappedErrs = append(p.wrappedErrs, argNum)
			fallthrough
		case verb == 'v':
			// Go syntax
			p.sharpV = p.sharp
			p.sharp = false
			// Struct-field syntax
			p.plusV = p.plus
			p.plus = false
			fallthrough
		default:
			p.printArg(a[argNum], verb)
			argNum++
		}
	}

	// Check for extra arguments unless the call accessed the arguments
	// out of order, in which case it's too expensive to detect if they've all
	// been used and arguably OK if they're not.
	if !p.reordered && argNum < len(a) {
		p._clearflags()
		p.buf.writeString(extraString)
		for i, arg := range a[argNum:] {
			if i > 0 {
				p.buf.writeString(commaSpaceString)
			}
			if arg == nil {
				p.buf.writeString(nilAngleString)
			} else {
				p.buf.writeString(reflect.TypeOf(arg).String())
				p.buf.writeByte('=')
				p.printArg(arg, 'v')
			}
		}
		p.buf.writeByte(')')
	}
}

func (p *pp) doPrint(a []any) {
	prevString := false
	for argNum, arg := range a {
		isString := arg != nil && reflect.TypeOf(arg).Kind() == reflect.String
		// Add a space between two non-string arguments.
		if argNum > 0 && !isString && !prevString {
			p.buf.writeByte(' ')
		}
		p.printArg(arg, 'v')
		prevString = isString
	}
}

// doPrintln is like doPrint but always adds a space between arguments
// and a newline after the last argument.
func (p *pp) doPrintln(a []any) {
	for argNum, arg := range a {
		if argNum > 0 {
			p.buf.writeByte(' ')
		}
		p.printArg(arg, 'v')
	}
	p.buf.writeByte('\n')
}

func (f *pp) _clearflags() {
	f.fmtFlags = fmtFlags{}
}

func (f *pp) _init() {
	f._clearflags()
}

// writePadding generates n bytes of padding.
func (f *pp) _writePadding(n int) {
	if n <= 0 { // No padding bytes needed.
		return
	}
	buf := f.buf
	oldLen := len(buf)
	newLen := oldLen + n
	// Make enough room for padding.
	if newLen > cap(buf) {
		buf = make(buffer, cap(buf)*2+n)
		copy(buf, f.buf)
	}
	// Decide which byte the padding should be filled with.
	padByte := byte(' ')
	if f.zero {
		padByte = byte('0')
	}
	// Fill padding with padByte.
	padding := buf[oldLen:newLen]
	for i := range padding {
		padding[i] = padByte
	}
	f.buf = buf[:newLen]
}

// pad appends b to f.buf, padded on left (!f.minus) or right (f.minus).
func (f *pp) _pad(b []byte) {
	if !f.widPresent || f.wid == 0 {
		f.buf.write(b)
		return
	}
	width := f.wid - utf8.RuneCount(b)
	if !f.minus {
		// left padding
		f._writePadding(width)
		f.buf.write(b)
	} else {
		// right padding
		f.buf.write(b)
		f._writePadding(width)
	}
}

// padString appends s to f.buf, padded on left (!f.minus) or right (f.minus).
func (f *pp) _padString(s string) {
	if !f.widPresent || f.wid == 0 {
		f.buf.writeString(s)
		return
	}
	width := f.wid - utf8.RuneCountInString(s)
	if !f.minus {
		// left padding
		f._writePadding(width)
		f.buf.writeString(s)
	} else {
		// right padding
		f.buf.writeString(s)
		f._writePadding(width)
	}
}

// fmtBoolean formats a boolean.
func (f *pp) _fmtBoolean(v bool) {
	if v {
		f._padString("true")
	} else {
		f._padString("false")
	}
}

// fmtUnicode formats a uint64 as "U+0078" or with f.sharp set as "U+0078 'x'".
func (f *pp) _fmtUnicode(u uint64) {
	buf := f.intbuf[0:]

	// With default precision set the maximum needed buf length is 18
	// for formatting -1 with %#U ("U+FFFFFFFFFFFFFFFF") which fits
	// into the already allocated intbuf with a capacity of 68 bytes.
	prec := 4
	if f.precPresent && f.prec > 4 {
		prec = f.prec
		// Compute space needed for "U+" , number, " '", character, "'".
		width := 2 + prec + 2 + utf8.UTFMax + 1
		if width > len(buf) {
			buf = make([]byte, width)
		}
	}

	// Format into buf, ending at buf[i]. Formatting numbers is easier right-to-left.
	i := len(buf)

	// For %#U we want to add a space and a quoted character at the end of the buffer.
	if f.sharp && u <= utf8.MaxRune && strconv.IsPrint(rune(u)) {
		i--
		buf[i] = '\''
		i -= utf8.RuneLen(rune(u))
		utf8.EncodeRune(buf[i:], rune(u))
		i--
		buf[i] = '\''
		i--
		buf[i] = ' '
	}
	// Format the Unicode code point u as a hexadecimal number.
	for u >= 16 {
		i--
		buf[i] = udigits[u&0xF]
		prec--
		u >>= 4
	}
	i--
	buf[i] = udigits[u]
	prec--
	// Add zeros in front of the number until requested precision is reached.
	for prec > 0 {
		i--
		buf[i] = '0'
		prec--
	}
	// Add a leading "U+".
	i--
	buf[i] = '+'
	i--
	buf[i] = 'U'

	oldZero := f.zero
	f.zero = false
	f._pad(buf[i:])
	f.zero = oldZero
}

// fmtInteger formats signed and unsigned integers.
func (f *pp) _fmtInteger(u uint64, base int, isSigned bool, verb rune, digits string) {
	negative := isSigned && int64(u) < 0
	if negative {
		u = -u
	}

	buf := f.intbuf[0:]
	// The already allocated f.intbuf with a capacity of 68 bytes
	// is large enough for integer formatting when no precision or width is set.
	if f.widPresent || f.precPresent {
		// Account 3 extra bytes for possible addition of a sign and "0x".
		width := 3 + f.wid + f.prec // wid and prec are always positive.
		if width > len(buf) {
			// We're going to need a bigger boat.
			buf = make([]byte, width)
		}
	}

	// Two ways to ask for extra leading zero digits: %.3d or %03d.
	// If both are specified the f.zero flag is ignored and
	// padding with spaces is used instead.
	prec := 0
	if f.precPresent {
		prec = f.prec
		// Precision of 0 and value of 0 means "print nothing" but padding.
		if prec == 0 && u == 0 {
			oldZero := f.zero
			f.zero = false
			f._writePadding(f.wid)
			f.zero = oldZero
			return
		}
	} else if f.zero && f.widPresent {
		prec = f.wid
		if negative || f.plus || f.space {
			prec-- // leave room for sign
		}
	}

	// Because printing is easier right-to-left: format u into buf, ending at buf[i].
	// We could make things marginally faster by splitting the 32-bit case out
	// into a separate block but it's not worth the duplication, so u has 64 bits.
	i := len(buf)
	// Use constants for the division and modulo for more efficient code.
	// Switch cases ordered by popularity.
	switch base {
	case 10:
		for u >= 10 {
			i--
			next := u / 10
			buf[i] = byte('0' + u - next*10)
			u = next
		}
	case 16:
		for u >= 16 {
			i--
			buf[i] = digits[u&0xF]
			u >>= 4
		}
	case 8:
		for u >= 8 {
			i--
			buf[i] = byte('0' + u&7)
			u >>= 3
		}
	case 2:
		for u >= 2 {
			i--
			buf[i] = byte('0' + u&1)
			u >>= 1
		}
	default:
		panic("fmt: unknown base; can't happen")
	}
	i--
	buf[i] = digits[u]
	for i > 0 && prec > len(buf)-i {
		i--
		buf[i] = '0'
	}

	// Various prefixes: 0x, -, etc.
	if f.sharp {
		switch base {
		case 2:
			// Add a leading 0b.
			i--
			buf[i] = 'b'
			i--
			buf[i] = '0'
		case 8:
			if buf[i] != '0' {
				i--
				buf[i] = '0'
			}
		case 16:
			// Add a leading 0x or 0X.
			i--
			buf[i] = digits[16]
			i--
			buf[i] = '0'
		}
	}
	if verb == 'O' {
		i--
		buf[i] = 'o'
		i--
		buf[i] = '0'
	}

	if negative {
		i--
		buf[i] = '-'
	} else if f.plus {
		i--
		buf[i] = '+'
	} else if f.space {
		i--
		buf[i] = ' '
	}

	// Left padding with zeros has already been handled like precision earlier
	// or the f.zero flag is ignored due to an explicitly set precision.
	oldZero := f.zero
	f.zero = false
	f._pad(buf[i:])
	f.zero = oldZero
}

// truncateString truncates the string s to the specified precision, if present.
func (f *pp) _truncateString(s string) string {
	if f.precPresent {
		n := f.prec
		for i := range s {
			n--
			if n < 0 {
				return s[:i]
			}
		}
	}
	return s
}

// truncate truncates the byte slice b as a string of the specified precision, if present.
func (f *pp) _truncate(b []byte) []byte {
	if f.precPresent {
		n := f.prec
		for i := 0; i < len(b); {
			n--
			if n < 0 {
				return b[:i]
			}
			wid := 1
			if b[i] >= utf8.RuneSelf {
				_, wid = utf8.DecodeRune(b[i:])
			}
			i += wid
		}
	}
	return b
}

// fmtS formats a string.
func (f *pp) _fmtS(s string) {
	s = f._truncateString(s)
	f._padString(s)
}

// fmtBs formats the byte slice b as if it was formatted as string with fmtS.
func (f *pp) _fmtBs(b []byte) {
	b = f._truncate(b)
	f._pad(b)
}

// fmtSbx formats a string or byte slice as a hexadecimal encoding of its bytes.
func (f *pp) _fmtSbx(s string, b []byte, digits string) {
	length := len(b)
	if b == nil {
		// No byte slice present. Assume string s should be encoded.
		length = len(s)
	}
	// Set length to not process more bytes than the precision demands.
	if f.precPresent && f.prec < length {
		length = f.prec
	}
	// Compute width of the encoding taking into account the f.sharp and f.space flag.
	width := 2 * length
	if width > 0 {
		if f.space {
			// Each element encoded by two hexadecimals will get a leading 0x or 0X.
			if f.sharp {
				width *= 2
			}
			// Elements will be separated by a space.
			width += length - 1
		} else if f.sharp {
			// Only a leading 0x or 0X will be added for the whole string.
			width += 2
		}
	} else { // The byte slice or string that should be encoded is empty.
		if f.widPresent {
			f._writePadding(f.wid)
		}
		return
	}
	// Handle padding to the left.
	if f.widPresent && f.wid > width && !f.minus {
		f._writePadding(f.wid - width)
	}
	// Write the encoding directly into the output buffer.
	buf := f.buf
	if f.sharp {
		// Add leading 0x or 0X.
		buf = append(buf, '0', digits[16])
	}
	var c byte
	for i := 0; i < length; i++ {
		if f.space && i > 0 {
			// Separate elements with a space.
			buf = append(buf, ' ')
			if f.sharp {
				// Add leading 0x or 0X for each element.
				buf = append(buf, '0', digits[16])
			}
		}
		if b != nil {
			c = b[i] // Take a byte from the input byte slice.
		} else {
			c = s[i] // Take a byte from the input string.
		}
		// Encode each byte as two hexadecimal digits.
		buf = append(buf, digits[c>>4], digits[c&0xF])
	}
	f.buf = buf
	// Handle padding to the right.
	if f.widPresent && f.wid > width && f.minus {
		f._writePadding(f.wid - width)
	}
}

// fmtSx formats a string as a hexadecimal encoding of its bytes.
func (f *pp) _fmtSx(s, digits string) {
	f._fmtSbx(s, nil, digits)
}

// fmtBx formats a byte slice as a hexadecimal encoding of its bytes.
func (f *pp) _fmtBx(b []byte, digits string) {
	f._fmtSbx("", b, digits)
}

// fmtQ formats a string as a double-quoted, escaped Go string constant.
// If f.sharp is set a raw (backquoted) string may be returned instead
// if the string does not contain any control characters other than tab.
func (f *pp) _fmtQ(s string) {
	s = f._truncateString(s)
	if f.sharp && strconv.CanBackquote(s) {
		f._padString("`" + s + "`")
		return
	}
	buf := f.intbuf[:0]
	if f.plus {
		f._pad(strconv.AppendQuoteToASCII(buf, s))
	} else {
		f._pad(strconv.AppendQuote(buf, s))
	}
}

// fmtC formats an integer as a Unicode character.
// If the character is not valid Unicode, it will print '\ufffd'.
func (f *pp) _fmtC(c uint64) {
	// Explicitly check whether c exceeds utf8.MaxRune since the conversion
	// of a uint64 to a rune may lose precision that indicates an overflow.
	r := rune(c)
	if c > utf8.MaxRune {
		r = utf8.RuneError
	}
	buf := f.intbuf[:0]
	f._pad(utf8.AppendRune(buf, r))
}

// fmtQc formats an integer as a single-quoted, escaped Go character constant.
// If the character is not valid Unicode, it will print '\ufffd'.
func (f *pp) _fmtQc(c uint64) {
	r := rune(c)
	if c > utf8.MaxRune {
		r = utf8.RuneError
	}
	buf := f.intbuf[:0]
	if f.plus {
		f._pad(strconv.AppendQuoteRuneToASCII(buf, r))
	} else {
		f._pad(strconv.AppendQuoteRune(buf, r))
	}
}

// fmtFloat formats a float64. It assumes that verb is a valid format specifier
// for strconv.AppendFloat and therefore fits into a byte.
func (f *pp) _fmtFloat(v float64, size int, verb rune, prec int) {
	// Explicit precision in format specifier overrules default precision.
	if f.precPresent {
		prec = f.prec
	}
	// Format number, reserving space for leading + sign if needed.
	num := strconv.AppendFloat(f.intbuf[:1], v, byte(verb), prec, size)
	if num[1] == '-' || num[1] == '+' {
		num = num[1:]
	} else {
		num[0] = '+'
	}
	// f.space means to add a leading space instead of a "+" sign unless
	// the sign is explicitly asked for by f.plus.
	if f.space && num[0] == '+' && !f.plus {
		num[0] = ' '
	}
	// Special handling for infinities and NaN,
	// which don't look like a number so shouldn't be padded with zeros.
	if num[1] == 'I' || num[1] == 'N' {
		oldZero := f.zero
		f.zero = false
		// Remove sign before NaN if not asked for.
		if num[1] == 'N' && !f.space && !f.plus {
			num = num[1:]
		}
		f._pad(num)
		f.zero = oldZero
		return
	}
	// The sharp flag forces printing a decimal point for non-binary formats
	// and retains trailing zeros, which we may need to restore.
	if f.sharp && verb != 'b' {
		digits := 0
		switch verb {
		case 'v', 'g', 'G', 'x':
			digits = prec
			// If no precision is set explicitly use a precision of 6.
			if digits == -1 {
				digits = 6
			}
		}

		// Buffer pre-allocated with enough room for
		// exponent notations of the form "e+123" or "p-1023".
		var tailBuf [6]byte
		tail := tailBuf[:0]

		hasDecimalPoint := false
		sawNonzeroDigit := false
		// Starting from i = 1 to skip sign at num[0].
		for i := 1; i < len(num); i++ {
			switch num[i] {
			case '.':
				hasDecimalPoint = true
			case 'p', 'P':
				tail = append(tail, num[i:]...)
				num = num[:i]
			case 'e', 'E':
				if verb != 'x' && verb != 'X' {
					tail = append(tail, num[i:]...)
					num = num[:i]
					break
				}
				fallthrough
			default:
				if num[i] != '0' {
					sawNonzeroDigit = true
				}
				// Count significant digits after the first non-zero digit.
				if sawNonzeroDigit {
					digits--
				}
			}
		}
		if !hasDecimalPoint {
			// Leading digit 0 should contribute once to digits.
			if len(num) == 2 && num[1] == '0' {
				digits--
			}
			num = append(num, '.')
		}
		for digits > 0 {
			num = append(num, '0')
			digits--
		}
		num = append(num, tail...)
	}
	// We want a sign if asked for and if the sign is not positive.
	if f.plus || num[0] != '+' {
		// If we're zero padding to the left we want the sign before the leading zeros.
		// Achieve this by writing the sign out and then padding the unsigned number.
		if f.zero && f.widPresent && f.wid > len(num) {
			f.buf.writeByte(num[0])
			f._writePadding(f.wid - len(num))
			f.buf.write(num[1:])
			return
		}
		f._pad(num)
		return
	}
	// No sign to show and the number is positive; just print the unsigned number.
	f._pad(num[1:])
}
