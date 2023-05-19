//go:build tinygo.wasm || 386 || amd64 || amd64p32 || arm || arm64 || loong64 || mips64le || mips64p32 || mips64p32le || mipsle || ppc64le || riscv || riscv64 || wasm

package wasmtimex

/*
#include <wasmtime.h>
#define WASMTIME_VAL_OF_OFFSET offsetof(struct wasmtime_val, of)
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func init() {
	if unsafe.Sizeof(C.wasmtime_val_t{}) != unsafe.Sizeof(Val{}) {
		panic(fmt.Sprintf("sizeof(wasmtime_val_t) != sizeof(Val): %d != %d",
			uint(unsafe.Sizeof(C.wasmtime_val_t{})), uint(unsafe.Sizeof(Val{}))))
	}
}

//type ValKindT uint8
//
//const (
//	I32       ValKindT = 0
//	I64       ValKindT = 1
//	F32       ValKindT = 2
//	F64       ValKindT = 3
//	V128      ValKindT = 4
//	FuncRef   ValKindT = 5
//	ExternRef ValKindT = 6
//)

type ValRaw [16]byte

func (v *ValRaw) I32() int32 {
	return *(*int32)(unsafe.Pointer(v))
}
func (v *ValRaw) SetI32(value int32) {
	*(*int32)(unsafe.Pointer(v)) = value
}

func (v *ValRaw) I64() int64 {
	return *(*int64)(unsafe.Pointer(v))
}
func (v *ValRaw) SetI64(value int64) {
	*(*int64)(unsafe.Pointer(v)) = value
}

func (v *ValRaw) F32() float32 {
	return *(*float32)(unsafe.Pointer(v))
}
func (v *ValRaw) SetF32(value float32) {
	*(*float32)(unsafe.Pointer(v)) = value
}

func (v *ValRaw) F64() float64 {
	return *(*float64)(unsafe.Pointer(v))
}
func (v *ValRaw) SetF64(value float64) {
	*(*float64)(unsafe.Pointer(v)) = value
}

// Val is a primitive numeric value.
// Moreover, in the definition of programs, immutable sequences of values
// occur to represent more complex data, such as text strings or other vectors.
type Val struct {
	Kind ValKind
	_    [C.WASMTIME_VAL_OF_OFFSET - 1]byte
	Val  [16]byte
}

func (v *Val) ptr() *C.wasmtime_val_t {
	return (*C.wasmtime_val_t)(unsafe.Pointer(v))
}

// I32 returns the underlying 32-bit integer if this is an `i32`, or panics.
func (v *Val) I32() int32 {
	return *(*int32)(unsafe.Pointer(&v.Val))
}

// I64 returns the underlying 64-bit integer if this is an `i64`, or panics.
func (v *Val) I64() int64 {
	return *(*int64)(unsafe.Pointer(&v.Val))
}

// F32 returns the underlying 32-bit float if this is an `f32`, or panics.
func (v *Val) F32() float32 {
	return *(*float32)(unsafe.Pointer(&v.Val))
}

// F64 returns the underlying 64-bit float if this is an `f64`, or panics.
func (v *Val) F64() float64 {
	return *(*float64)(unsafe.Pointer(&v.Val))
}

func (v *Val) SetI32(value int32) {
	v.Kind = KindI32
	*(*int32)(unsafe.Pointer(&v.Val)) = value
}

func (v *Val) SetI64(value int64) {
	v.Kind = KindI64
	*(*int64)(unsafe.Pointer(&v.Val)) = value
}

func (v *Val) SetF32(value float32) {
	v.Kind = KindF32
	*(*float32)(unsafe.Pointer(&v.Val)) = value
}

func (v *Val) SetF64(value float64) {
	v.Kind = KindF64
	*(*float64)(unsafe.Pointer(&v.Val)) = value
}

// ValI32 converts a go int32 to a i32 Val
func ValI32(val int32) Val {
	ret := Val{}
	ret.SetI32(val)
	return ret
}

// ValI64 converts a go int64 to a i64 Val
func ValI64(val int64) Val {
	ret := Val{}
	ret.SetI64(val)
	return ret
}

// ValF32 converts a go float32 to a f32 Val
func ValF32(val float32) Val {
	ret := Val{}
	ret.SetF32(val)
	return ret
}

// ValF64 converts a go float64 to a f64 Val
func ValF64(val float64) Val {
	ret := Val{}
	ret.SetF64(val)
	return ret
}

// ValFuncref converts a Func to a funcref Val
//
// Note that `f` can be `nil` to represent a null `funcref`.
//func ValFuncref(f *Func) Val {
//	ret := Val{_raw: &C.wasmtime_val_t{kind: C.WASMTIME_FUNCREF}}
//	if f != nil {
//		C.go_wasmtime_val_funcref_set(ret.ptr(), f.val)
//	}
//	return ret
//}
// Funcref returns the underlying function if this is a `funcref`, or panics.
//
// Note that a null `funcref` is returned as `nil`.
//func (v Val) Funcref() *Func {
//	if v.Kind() != KindFuncref {
//		panic("not a funcref")
//	}
//	val := C.go_wasmtime_val_funcref_get(v.ptr())
//	if val.store_id == 0 {
//		return nil
//	} else {
//		return mkFunc(val)
//	}
//}

// Externref returns the underlying value if this is an `externref`, or panics.
//
// Note that a null `externref` is returned as `nil`.
//func (v Val) Externref() interface{} {
//	if v.Kind() != KindExternref {
//		panic("not an externref")
//	}
//	val := C.go_wasmtime_val_externref_get(v.ptr())
//	if val == nil {
//		return nil
//	}
//	data := C.wasmtime_externref_data(val)
//
//	gExternrefLock.Lock()
//	defer gExternrefLock.Unlock()
//	return gExternrefMap[int(uintptr(data))-1]
//}

// Get returns the underlying 64-bit float if this is an `f64`, or panics.
func (v *Val) Get() interface{} {
	switch v.Kind {
	case KindI32:
		return v.I32()
	case KindI64:
		return v.I64()
	case KindF32:
		return v.F32()
	case KindF64:
		return v.F64()
		//case KindFuncref:
		//	return v.Funcref()
		//case KindExternref:
		//	return v.Externref()
	}
	panic("failed to get value of `Val`")
}
