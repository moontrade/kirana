package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

void do_wasmtime_error_delete(size_t arg0, size_t arg1) {
	wasmtime_error_delete(
		(wasmtime_error_t*)arg0
	);
}

void do_wasmtime_error_message(size_t arg0, size_t arg1) {
	wasmtime_error_message(
		(wasmtime_error_t*)arg0,
		(wasm_byte_vec_t*)arg1
	);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

type Error C.wasmtime_error_t

func (e *Error) Delete() {
	if e == nil {
		return
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_error_delete), uintptr(unsafe.Pointer(e)), 0)
}

func (e *Error) ptr() *C.wasmtime_error_t {
	return (*C.wasmtime_error_t)(unsafe.Pointer(e))
}

func (e *Error) Message() (result ByteVec) {
	cgo.NonBlocking((*byte)(C.do_wasmtime_error_message), uintptr(unsafe.Pointer(e)), uintptr(unsafe.Pointer(&result)))
	return
}

func (e *Error) Error() string {
	m := e.Message()
	defer m.Delete()
	s := m.String()
	return s
}
