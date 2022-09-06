package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

void do_wasm_valtype_new(size_t arg0, size_t arg1) {
	*((wasm_valtype_t**)arg1) = wasm_valtype_new(
		(wasm_valkind_t)arg0
	);
}

void do_wasm_valtype_delete(size_t arg0, size_t arg1) {
	wasm_valtype_delete((wasm_valtype_t*)(void*)arg0);
}

void do_wasm_valtype_kind(size_t arg0, size_t arg1) {
	*((wasm_valkind_t*)arg1) = wasm_valtype_kind(
		(const wasm_valtype_t*)(void*)arg0
	);
}

void do_wasm_valtype_vec_new_uninitialized(size_t arg0, size_t arg1) {
	wasm_valtype_vec_new_uninitialized(
		(wasm_valtype_vec_t*)(void*)arg0,
		arg1
	);
}

typedef struct do_wasm_valtype_vec_new_t {
	size_t out;
	size_t size;
	size_t data;
} do_wasm_valtype_vec_new_t;

void do_wasm_valtype_vec_new(size_t arg0, size_t arg1) {
	do_wasm_valtype_vec_new_t* args = (do_wasm_valtype_vec_new_t*)(void*)arg0;
	wasm_valtype_vec_new(
		(wasm_valtype_vec_t*)(void*)args->out,
		args->size,
		(wasm_valtype_t* const*)(void*)args->data
	);
}

void do_wasm_valtype_vec_delete(size_t arg0, size_t arg1) {
	wasm_valtype_vec_delete((wasm_valtype_vec_t*)(void*)arg0);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"reflect"
	"unsafe"
)

// ValKind enumeration of different kinds of value types
type ValKind C.wasm_valkind_t

const (
	// KindI32 is the types i32 classify 32-bit integers. Integers are not inherently signed or unsigned, their interpretation is determined by individual operations.
	KindI32 ValKind = C.WASM_I32
	// KindI64 is the types i64 classify 64-bit integers. Integers are not inherently signed or unsigned, their interpretation is determined by individual operations.
	KindI64 ValKind = C.WASM_I64
	// KindF32 is the types f32 classify 32-bit floating-point data. They correspond to the respective binary floating-point representations, also known as single and double precision, as defined by the IEEE 754-2019 standard.
	KindF32 ValKind = C.WASM_F32
	// KindF64 is the types f64 classify 64-bit floating-point data. They correspond to the respective binary floating-point representations, also known as single and double precision, as defined by the IEEE 754-2019 standard.
	KindF64  ValKind = C.WASM_F64
	KindV128 ValKind = 4
	// TODO: Unknown
	KindExternref ValKind = C.WASM_ANYREF
	// KindFuncref is the infinite union of all function types.
	KindFuncref ValKind = C.WASM_FUNCREF
)

// String renders this kind as a string, similar to the `*.wat` format
func (ty ValKind) String() string {
	switch ty {
	case KindI32:
		return "i32"
	case KindI64:
		return "i64"
	case KindF32:
		return "f32"
	case KindF64:
		return "f64"
	case KindExternref:
		return "externref"
	case KindFuncref:
		return "funcref"
	}
	panic("unknown kind")
}

// ValType means one of the value types, which classify the individual values that WebAssembly code can compute with and the values that a variable accepts.
type ValType C.wasm_valtype_t

// NewValType creates a new `ValType` with the `kind` provided
func NewValType(kind ValKind) *ValType {
	var result *ValType
	cgo.NonBlocking((*byte)(C.do_wasm_valtype_new), uintptr(kind), uintptr(unsafe.Pointer(&result)))
	return result
}

func (t *ValType) Delete() {
	cgo.NonBlocking((*byte)(C.do_wasm_valtype_delete), uintptr(unsafe.Pointer(t)), 0)
}

// Kind returns the corresponding `ValKind` for this `ValType`
func (t *ValType) Kind() ValKind {
	var kind ValKind
	cgo.NonBlocking((*byte)(C.do_wasm_valtype_kind), uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(&kind)))
	return kind
}

// Converts this `ValType` into a string according to the string representation
// of `ValKind`.
func (t *ValType) String() string {
	return t.Kind().String()
}

func (t *ValType) ptr() *C.wasm_valtype_t {
	return (*C.wasm_valtype_t)(unsafe.Pointer(t))
}

func (t *ValType) Clone() *ValType {
	return NewValType(t.Kind())
}

/*
#define WASM_DECLARE_VEC(name, ptr_or_none) \
  typedef struct wasm_##name##_vec_t { \
    size_t size; \
    wasm_##name##_t ptr_or_none* data; \
  } wasm_##name##_vec_t; \
  \
  WASM_API_EXTERN void wasm_##name##_vec_new_empty(own wasm_##name##_vec_t* out); \
  WASM_API_EXTERN void wasm_##name##_vec_new_uninitialized( \
    own wasm_##name##_vec_t* out, size_t); \
  WASM_API_EXTERN void wasm_##name##_vec_new( \
    own wasm_##name##_vec_t* out, \
    size_t, own wasm_##name##_t ptr_or_none const[]); \
  WASM_API_EXTERN void wasm_##name##_vec_copy( \
    own wasm_##name##_vec_t* out, const wasm_##name##_vec_t*); \
  WASM_API_EXTERN void wasm_##name##_vec_delete(own wasm_##name##_vec_t*);
*/

type ValTypeVec struct {
	size uintptr
	data uintptr
}

func SizeofvalTypeVec() int {
	return int(unsafe.Sizeof(C.wasm_valtype_vec_t{}))
}

func NewValTypeVec(size int) ValTypeVec {
	if size < 0 {
		size = 0
	}
	var result ValTypeVec
	cgo.NonBlocking((*byte)(C.do_wasm_valtype_vec_new_uninitialized), uintptr(unsafe.Pointer(&result)), uintptr(size))
	return result
}

func NewValTypeVecOf(data []*ValType) ValTypeVec {
	var result ValTypeVec
	args := struct {
		out  uintptr
		size uintptr
		data uintptr
	}{
		out:  uintptr(unsafe.Pointer(&result)),
		size: uintptr(len(data)),
		data: (*(*reflect.SliceHeader)(unsafe.Pointer(&data))).Data,
	}
	cgo.NonBlocking((*byte)(C.do_wasm_valtype_vec_new), uintptr(unsafe.Pointer(&args)), 0)
	return result
}

func (vec *ValTypeVec) Delete() {
	if vec.data == 0 {
		return
	}
	cgo.NonBlocking((*byte)(C.do_wasm_valtype_vec_delete), uintptr(unsafe.Pointer(vec)), 0)
}

func (vec *ValTypeVec) Unsafe() []*ValType {
	return *(*[]*ValType)(unsafe.Pointer(&reflect.SliceHeader{
		Data: vec.data,
		Len:  int(vec.size),
		Cap:  int(vec.size),
	}))
}

func (vec *ValTypeVec) Get(index int) *ValType {
	if vec.data == 0 || index < 0 || index >= int(vec.size) {
		return nil
	}
	return *(**ValType)(unsafe.Pointer(vec.data + (unsafe.Sizeof(uintptr(0)) * uintptr(index))))
}

func (vec *ValTypeVec) Set(index int, value *ValType) {
	if vec.data == 0 || index < 0 || index >= int(vec.size) {
		return
	}
	*(**ValType)(unsafe.Pointer(vec.data + (unsafe.Sizeof(uintptr(0)) * uintptr(index)))) = value
}
