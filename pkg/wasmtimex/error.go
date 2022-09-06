package wasmtimex

// #include <wasmtime.h>
import "C"
import (
	"unsafe"
)

type Error C.wasmtime_error_t

func mkError(ptr *C.wasmtime_error_t) *Error {
	return (*Error)(unsafe.Pointer(ptr))
}

func (e *Error) Delete() {
	if e == nil {
		return
	}
	C.wasmtime_error_delete((*C.wasmtime_error_t)(unsafe.Pointer(e)))
	*(*uintptr)(unsafe.Pointer(e)) = 0
}

func (e *Error) ptr() *C.wasmtime_error_t {
	return (*C.wasmtime_error_t)(unsafe.Pointer(e))
}

func (e *Error) Error() string {
	message := C.wasm_byte_vec_t{}
	C.wasmtime_error_message(e.ptr(), &message)
	ret := C.GoStringN(message.data, C.int(message.size))
	C.wasm_byte_vec_delete(&message)
	return ret
}
