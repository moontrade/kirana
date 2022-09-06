package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasm_exporttype_new_t {
	size_t name;
	size_t extern_type;
	size_t result;
} do_wasm_exporttype_new_t;

void do_wasm_exporttype_new(size_t arg0, size_t arg1) {
	do_wasm_exporttype_new_t* args = (do_wasm_exporttype_new_t*)(void*)arg0;
	args->result = (size_t)(void*)wasm_exporttype_new(
		(wasm_byte_vec_t*)(void*)args->name,
		(wasm_externtype_t*)(void*)args->extern_type
	);
}

void do_wasm_exporttype_delete(size_t arg0, size_t arg1) {
	wasm_exporttype_delete(
		(wasm_exporttype_t*)(void*)arg0
	);
}

void do_wasm_exporttype_name(size_t arg0, size_t arg1) {
	*((const wasm_name_t**)(void*)arg1) = wasm_exporttype_name(
		(const wasm_exporttype_t*)(void*)arg0
	);
}

void do_wasm_exporttype_type(size_t arg0, size_t arg1) {
	*((const wasm_externtype_t**)(void*)arg1) = wasm_exporttype_type(
		(const wasm_exporttype_t*)(void*)arg0
	);
}

void do_wasm_exporttype_vec_new_uninitialized(size_t arg0, size_t arg1) {
	wasm_exporttype_vec_new_uninitialized(
		(wasm_exporttype_vec_t*)(void*)arg0,
		arg1
	);
}

typedef struct do_wasm_exporttype_vec_new_t {
	size_t out;
	size_t size;
	size_t data;
} do_wasm_exporttype_vec_new_t;

void do_wasm_exporttype_vec_new(size_t arg0, size_t arg1) {
	do_wasm_exporttype_vec_new_t* args = (do_wasm_exporttype_vec_new_t*)(void*)arg0;
	wasm_exporttype_vec_new(
		(wasm_exporttype_vec_t*)(void*)args->out,
		args->size,
		(wasm_exporttype_t* const*)(void*)args->data
	);
}

void do_wasm_exporttype_vec_delete(size_t arg0, size_t arg1) {
	wasm_exporttype_vec_delete((wasm_exporttype_vec_t*)(void*)arg0);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"reflect"
	"unsafe"
)

// ExportType is one of the exports component
type ExportType C.wasm_exporttype_t

func NewExportType(name ByteVec, externType *ExternType) *ExportType {
	args := struct {
		name       uintptr
		externType uintptr
		result     uintptr
	}{
		name:       uintptr(unsafe.Pointer(&name)),
		externType: uintptr(unsafe.Pointer(externType)),
	}
	cgo.NonBlocking((*byte)(C.do_wasm_exporttype_new), uintptr(unsafe.Pointer(&args)), 0)
	return (*ExportType)(unsafe.Pointer(args.result))
}

func (et *ExportType) Delete() {
	cgo.NonBlocking((*byte)(C.do_wasm_exporttype_delete), uintptr(unsafe.Pointer(et)), 0)
}

func (et *ExportType) Name() ByteVec {
	var ret ByteVec
	cgo.NonBlocking((*byte)(C.do_wasm_exporttype_name), uintptr(unsafe.Pointer(et)), uintptr(unsafe.Pointer(&ret)))
	return ret
}

func (et *ExportType) Type() *ExternType {
	var ret *ExternType
	cgo.NonBlocking((*byte)(C.do_wasm_exporttype_type), uintptr(unsafe.Pointer(et)), uintptr(unsafe.Pointer(&ret)))
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

type ExportTypeVec struct {
	size uintptr
	data uintptr
}

func NewExportTypeVec(size int) ExportTypeVec {
	if size < 0 {
		size = 0
	}
	var result ExportTypeVec
	cgo.NonBlocking((*byte)(C.do_wasm_exporttype_vec_new_uninitialized), uintptr(unsafe.Pointer(&result)), uintptr(size))
	return result
}

func NewExportTypeVecOf(data []*ExportType) ExportTypeVec {
	var result ExportTypeVec
	args := struct {
		out  uintptr
		size uintptr
		data uintptr
	}{
		out:  uintptr(unsafe.Pointer(&result)),
		size: uintptr(len(data)),
		data: (*(*reflect.SliceHeader)(unsafe.Pointer(&data))).Data,
	}
	cgo.NonBlocking((*byte)(C.do_wasm_exporttype_vec_new), uintptr(unsafe.Pointer(&args)), 0)
	return result
}

func (vec *ExportTypeVec) Delete() {
	if vec.data == 0 {
		return
	}
	cgo.NonBlocking((*byte)(C.do_wasm_exporttype_vec_delete), uintptr(unsafe.Pointer(vec)), 0)
}

func (vec *ExportTypeVec) Unsafe() []*ExportType {
	return *(*[]*ExportType)(unsafe.Pointer(&reflect.SliceHeader{
		Data: vec.data,
		Len:  int(vec.size),
		Cap:  int(vec.size),
	}))
}

func (vec *ExportTypeVec) Get(index int) *ExportType {
	if vec.data == 0 || index < 0 || index >= int(vec.size) {
		return nil
	}
	return *(**ExportType)(unsafe.Pointer(vec.data + (unsafe.Sizeof(uintptr(0)) * uintptr(index))))
}

func (vec *ExportTypeVec) Set(index int, value *ExportType) {
	if vec.data == 0 || index < 0 || index >= int(vec.size) {
		return
	}
	*(**ExportType)(unsafe.Pointer(vec.data + (unsafe.Sizeof(uintptr(0)) * uintptr(index)))) = value
}
