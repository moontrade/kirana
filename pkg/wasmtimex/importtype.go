package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasm_importtype_new_t {
	size_t module;
	size_t name;
	size_t extern_type;
	size_t result;
} do_wasm_importtype_new_t;

void do_wasm_importtype_new(size_t arg0, size_t arg1) {
	do_wasm_importtype_new_t* args = (do_wasm_importtype_new_t*)(void*)arg0;
	args->result = (size_t)(void*)wasm_importtype_new(
		(wasm_byte_vec_t*)(void*)args->module,
		(wasm_byte_vec_t*)(void*)args->name,
		(wasm_externtype_t*)(void*)args->extern_type
	);
}

void do_wasm_importtype_delete(size_t arg0, size_t arg1) {
	wasm_importtype_delete(
		(wasm_importtype_t*)(void*)arg0
	);
}

void do_wasm_importtype_module(size_t arg0, size_t arg1) {
	*((const wasm_name_t**)(void*)arg1) = wasm_importtype_module(
		(const wasm_importtype_t*)(void*)arg0
	);
}

void do_wasm_importtype_name(size_t arg0, size_t arg1) {
	*((const wasm_name_t**)(void*)arg1) = wasm_importtype_name(
		(const wasm_importtype_t*)(void*)arg0
	);
}

void do_wasm_importtype_type(size_t arg0, size_t arg1) {
	*((const wasm_externtype_t**)(void*)arg1) = wasm_importtype_type(
		(const wasm_importtype_t*)(void*)arg0
	);
}

void do_wasm_importtype_vec_new_uninitialized(size_t arg0, size_t arg1) {
	wasm_importtype_vec_new_uninitialized(
		(wasm_importtype_vec_t*)(void*)arg0,
		arg1
	);
}

typedef struct do_wasm_importtype_vec_new_t {
	size_t out;
	size_t size;
	size_t data;
} do_wasm_importtype_vec_new_t;

void do_wasm_importtype_vec_new(size_t arg0, size_t arg1) {
	do_wasm_importtype_vec_new_t* args = (do_wasm_importtype_vec_new_t*)(void*)arg0;
	wasm_importtype_vec_new(
		(wasm_importtype_vec_t*)(void*)args->out,
		args->size,
		(wasm_importtype_t* const*)(void*)args->data
	);
}

void do_wasm_importtype_vec_delete(size_t arg0, size_t arg1) {
	wasm_importtype_vec_delete((wasm_importtype_vec_t*)(void*)arg0);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"reflect"
	"unsafe"
)

// ImportType is one of the imports component
// A module defines a set of imports that are required for instantiation.
type ImportType C.wasm_importtype_t

func NewImportType(module, name ByteVec, externType *ExternType) *ImportType {
	args := struct {
		module     uintptr
		name       uintptr
		externType uintptr
		result     uintptr
	}{
		module:     uintptr(unsafe.Pointer(&module)),
		name:       uintptr(unsafe.Pointer(&name)),
		externType: uintptr(unsafe.Pointer(externType)),
	}
	cgo.NonBlocking((*byte)(C.do_wasm_importtype_new), uintptr(unsafe.Pointer(&args)), 0)
	return (*ImportType)(unsafe.Pointer(args.result))
}

func (it *ImportType) Delete() {
	cgo.NonBlocking((*byte)(C.do_wasm_importtype_delete), uintptr(unsafe.Pointer(it)), 0)
}

func (it *ImportType) Module() ByteVec {
	var ret ByteVec
	cgo.NonBlocking((*byte)(C.do_wasm_importtype_module), uintptr(unsafe.Pointer(it)), uintptr(unsafe.Pointer(&ret)))
	return ret
}

func (it *ImportType) Name() ByteVec {
	var ret ByteVec
	cgo.NonBlocking((*byte)(C.do_wasm_importtype_name), uintptr(unsafe.Pointer(it)), uintptr(unsafe.Pointer(&ret)))
	return ret
}

func (it *ImportType) Type() *ExternType {
	var ret *ExternType
	cgo.NonBlocking((*byte)(C.do_wasm_importtype_type), uintptr(unsafe.Pointer(it)), uintptr(unsafe.Pointer(&ret)))
	return ret
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

type ImportTypeVec struct {
	size uintptr
	data uintptr
}

func NewImportTypeVec(size int) ImportTypeVec {
	if size < 0 {
		size = 0
	}
	var result ImportTypeVec
	cgo.NonBlocking((*byte)(C.do_wasm_importtype_vec_new_uninitialized), uintptr(unsafe.Pointer(&result)), uintptr(size))
	return result
}

func NewImportTypeVecOf(data []*ImportType) ImportTypeVec {
	var result ImportTypeVec
	args := struct {
		out  uintptr
		size uintptr
		data uintptr
	}{
		out:  uintptr(unsafe.Pointer(&result)),
		size: uintptr(len(data)),
		data: (*(*reflect.SliceHeader)(unsafe.Pointer(&data))).Data,
	}
	cgo.NonBlocking((*byte)(C.do_wasm_importtype_vec_new), uintptr(unsafe.Pointer(&args)), 0)
	return result
}

func (vec *ImportTypeVec) Delete() {
	if vec.data == 0 {
		return
	}
	cgo.NonBlocking((*byte)(C.do_wasm_importtype_vec_delete), uintptr(unsafe.Pointer(vec)), 0)
}

func (vec *ImportTypeVec) Unsafe() []*ImportType {
	return *(*[]*ImportType)(unsafe.Pointer(&reflect.SliceHeader{
		Data: vec.data,
		Len:  int(vec.size),
		Cap:  int(vec.size),
	}))
}

func (vec *ImportTypeVec) Get(index int) *ImportType {
	if vec.data == 0 || index < 0 || index >= int(vec.size) {
		return nil
	}
	return *(**ImportType)(unsafe.Pointer(vec.data + (unsafe.Sizeof(uintptr(0)) * uintptr(index))))
}

func (vec *ImportTypeVec) Set(index int, value *ImportType) {
	if vec.data == 0 || index < 0 || index >= int(vec.size) {
		return
	}
	*(**ImportType)(unsafe.Pointer(vec.data + (unsafe.Sizeof(uintptr(0)) * uintptr(index)))) = value
}
