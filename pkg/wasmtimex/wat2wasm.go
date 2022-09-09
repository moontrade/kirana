package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>
*/
import "C"
import (
	"unsafe"
)

// Wat2Wasm converts the text format of WebAssembly to the binary format.
//
// Takes the text format in-memory as input, and returns either the binary
// encoding of the text format or an error if parsing fails.
func Wat2Wasm(wat string) (ByteVec, *Error) {
	var retVec ByteVec
	err := (*Error)(unsafe.Pointer(C.wasmtime_wat2wasm(
		(*C.char)(unsafe.Pointer(strDataPtr(wat))),
		C.size_t(len(wat)),
		&retVec.vec,
	)))
	return retVec, err
}

// Wat2WasmVec converts the text format of WebAssembly to the binary format.
//
// Takes the text format in-memory as input, and returns either the binary
// encoding of the text format or an error if parsing fails.
func Wat2WasmVec(wat ByteVec) (ByteVec, *Error) {
	var retVec ByteVec
	err := (*Error)(unsafe.Pointer(C.wasmtime_wat2wasm(
		(*C.char)(wat.Data()),
		C.size_t(wat.Size()),
		&retVec.vec,
	)))
	return retVec, err
}

// Wat2WasmBytes converts the text format of WebAssembly to the binary format.
//
// Takes the text format in-memory as input, and returns either the binary
// encoding of the text format or an error if parsing fails.
func Wat2WasmBytes(wat string) ([]byte, *Error) {
	var retVec ByteVec
	err := (*Error)(unsafe.Pointer(C.wasmtime_wat2wasm(
		(*C.char)(unsafe.Pointer(strDataPtr(wat))),
		C.size_t(len(wat)),
		&retVec.vec,
	)))
	if err == nil {
		r := retVec.Bytes()
		retVec.Delete()
		return r, nil
	}
	return nil, err
}
