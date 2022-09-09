package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasmtime_module_new_t {
	size_t engine;
	size_t wasm;
	size_t wasm_len;
	size_t result;
	size_t error;
} do_wasmtime_module_new_t;

void do_wasmtime_module_new(size_t arg0, size_t arg1) {
	do_wasmtime_module_new_t* args = (do_wasmtime_module_new_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_module_new(
		(wasm_engine_t*)args->engine,
		(const uint8_t*)args->wasm,
		args->wasm_len,
		(wasmtime_module_t**)(void*)&args->result
	);
}

typedef struct do_wasmtime_module_validate_t {
	size_t engine;
	size_t wasm;
	size_t wasm_len;
	size_t result;
	size_t error;
} do_wasmtime_module_validate_t;

void do_wasmtime_module_validate(size_t arg0, size_t arg1) {
	do_wasmtime_module_validate_t* args = (do_wasmtime_module_validate_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_module_validate(
		(wasm_engine_t*)args->engine,
		(const uint8_t*)args->wasm,
		args->wasm_len
	);
}

void do_wasmtime_module_delete(size_t arg0, size_t arg1) {
	wasmtime_module_delete(
		(wasmtime_module_t*)(void*)arg0
	);
}

void do_wasmtime_module_clone(size_t arg0, size_t arg1) {
	*((wasmtime_module_t**)(void*)arg1) = wasmtime_module_clone(
		(wasmtime_module_t*)(void*)arg0
	);
}

void do_wasmtime_module_imports(size_t arg0, size_t arg1) {
	wasmtime_module_imports(
		(const wasmtime_module_t*)(void*)arg0,
		(wasm_importtype_vec_t*)(void*)arg1
	);
}

void do_wasmtime_module_exports(size_t arg0, size_t arg1) {
	wasmtime_module_exports(
		(const wasmtime_module_t*)(void*)arg0,
		(wasm_exporttype_vec_t*)(void*)arg1
	);
}

typedef struct do_wasmtime_module_serialize_t {
	size_t module;
	size_t result;
	size_t error;
} do_wasmtime_module_serialize_t;

void do_wasmtime_module_serialize(size_t arg0, size_t arg1) {
	do_wasmtime_module_serialize_t* args = (do_wasmtime_module_serialize_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_module_serialize(
		(wasmtime_module_t*)args->module,
		(wasm_byte_vec_t*)args->result
	);
}

typedef struct do_wasmtime_module_deserialize_t {
	size_t engine;
	size_t bytes;
	size_t bytes_len;
	size_t result;
	size_t error;
} do_wasmtime_module_deserialize_t;

void do_wasmtime_module_deserialize(size_t arg0, size_t arg1) {
	do_wasmtime_module_deserialize_t* args = (do_wasmtime_module_deserialize_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_module_deserialize(
		(wasm_engine_t*)args->engine,
		(const uint8_t*)args->bytes,
		args->bytes_len,
		(wasmtime_module_t**)(void*)&args->result
	);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"reflect"
	"unsafe"
)

// Module is a module which collects definitions for types, functions, tables, memories, and globals.
// In addition, it can declare imports and exports and provide initialization logic in the form of
// data and element segments or a start function. Modules organized WebAssembly programs as the unit
// of deployment, loading, and compilation.
type Module C.wasmtime_module_t

// NewModule compiles a WebAssembly binary into a #wasmtime_module_t
//
// This function will compile a WebAssembly binary into an owned #wasm_module_t.
// This performs the same as #wasm_module_new except that it returns a
// #wasmtime_error_t type to get richer error information.
//
// On success the returned #wasmtime_error_t is `NULL` and the `ret` pointer is
// filled in with a #wasm_module_t. On failure the #wasmtime_error_t is
// non-`NULL` and the `ret` pointer is unmodified.
//
// This function does not take ownership of its arguments, but the
// returned error and module are owned by the caller.
func NewModule(engine *Engine, wasm ByteVec) (*Module, *Error) {
	args := struct {
		engine  uintptr
		wasm    uintptr
		wasmLen uintptr
		result  uintptr
		error   uintptr
	}{
		engine:  uintptr(unsafe.Pointer(engine)),
		wasm:    uintptr(wasm.Data()),
		wasmLen: uintptr(wasm.Size()),
	}
	cgo.Blocking((*byte)(C.do_wasmtime_module_new), uintptr(unsafe.Pointer(&args)), 0)
	return (*Module)(unsafe.Pointer(args.result)), (*Error)(unsafe.Pointer(args.error))
}

// NewModuleFromBytes compiles a WebAssembly binary into a #wasmtime_module_t
//
// This function will compile a WebAssembly binary into an owned #wasm_module_t.
// This performs the same as #wasm_module_new except that it returns a
// #wasmtime_error_t type to get richer error information.
//
// On success the returned #wasmtime_error_t is `NULL` and the `ret` pointer is
// filled in with a #wasm_module_t. On failure the #wasmtime_error_t is
// non-`NULL` and the `ret` pointer is unmodified.
//
// This function does not take ownership of its arguments, but the
// returned error and module are owned by the caller.
func NewModuleFromBytes(engine *Engine, wasm []byte) (*Module, *Error) {
	args := struct {
		engine  uintptr
		wasm    uintptr
		wasmLen uintptr
		result  uintptr
		error   uintptr
	}{
		engine:  uintptr(unsafe.Pointer(engine)),
		wasm:    dataPtr(wasm),
		wasmLen: uintptr(len(wasm)),
	}
	cgo.Blocking((*byte)(C.do_wasmtime_module_new), uintptr(unsafe.Pointer(&args)), 0)
	return (*Module)(unsafe.Pointer(args.result)), (*Error)(unsafe.Pointer(args.error))
}

// Validate a WebAssembly binary.
//
// This function will validate the provided byte sequence to determine if it is
// a valid WebAssembly binary within the context of the engine provided.
//
// This function does not take ownership of its arguments but the caller is
// expected to deallocate the returned error if it is non-`NULL`.
//
// If the binary validates then `NULL` is returned, otherwise the error returned
// describes why the binary did not validate.
func Validate(engine *Engine, wasm []byte) *Error {
	args := struct {
		engine  uintptr
		wasm    uintptr
		wasmLen uintptr
		error   uintptr
	}{
		engine:  uintptr(unsafe.Pointer(engine)),
		wasm:    (*(*reflect.SliceHeader)(unsafe.Pointer(&wasm))).Data,
		wasmLen: uintptr(len(wasm)),
	}
	cgo.Blocking((*byte)(C.do_wasmtime_module_validate), uintptr(unsafe.Pointer(&args)), 0)
	return (*Error)(unsafe.Pointer(args.error))
}

func (m *Module) Delete() {
	cgo.NonBlocking((*byte)(C.wasmtime_module_delete), uintptr(unsafe.Pointer(m)), 0)
}

// Clone creates a shallow clone of the specified module, increasing the
// internal reference count.
func (m *Module) Clone() *Module {
	var clone *Module
	cgo.NonBlocking((*byte)(C.wasmtime_module_clone), uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(&clone)))
	return clone
}

// Imports same as #wasm_module_imports, but for #wasmtime_module_t.
func (m *Module) Imports() ImportTypeVec {
	var vec ImportTypeVec
	cgo.NonBlocking((*byte)(C.do_wasmtime_module_imports), uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(&vec)))
	return vec
}

// Exports same as #wasm_module_exports, but for #wasmtime_module_t.
func (m *Module) Exports() ExportTypeVec {
	var vec ExportTypeVec
	cgo.NonBlocking((*byte)(C.do_wasmtime_module_exports), uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(&vec)))
	return vec
}

// Serialize serializes compiled module artifacts as blob data.
//
// \param module the module
// \param ret if the conversion is successful, this byte vector is filled in with
//
//	the serialized compiled module.
//
// \return a non-null error if parsing fails, or returns `NULL`. If parsing
// fails then `ret` isn't touched.
//
// This function does not take ownership of `module`, and the caller is
// expected to deallocate the returned #wasmtime_error_t and #wasm_byte_vec_t.
func (m *Module) Serialize() (ByteVec, *Error) {
	var vec ByteVec
	args := struct {
		module uintptr
		vec    uintptr
		error  uintptr
	}{
		module: uintptr(unsafe.Pointer(m)),
		vec:    uintptr(unsafe.Pointer(&vec)),
	}
	cgo.Blocking((*byte)(C.do_wasmtime_module_serialize), uintptr(unsafe.Pointer(&args)), 0)
	return vec, (*Error)(unsafe.Pointer(args.error))
}

// Deserialize builds a module from serialized data.
//
// This function does not take ownership of its arguments, but the
// returned error and module are owned by the caller.
//
// This function is not safe to receive arbitrary user input. See the Rust
// documentation for more information on what inputs are safe to pass in here
// (e.g. only that of #wasmtime_module_serialize)
func Deserialize(engine *Engine, serialized []byte) (*Module, *Error) {
	args := struct {
		engine   uintptr
		bytes    uintptr
		bytesLen uintptr
		result   uintptr
		error    uintptr
	}{
		engine:   uintptr(unsafe.Pointer(engine)),
		bytes:    (*(*reflect.SliceHeader)(unsafe.Pointer(&serialized))).Data,
		bytesLen: uintptr(len(serialized)),
	}
	cgo.Blocking((*byte)(C.do_wasmtime_module_deserialize), uintptr(unsafe.Pointer(&args)), 0)
	return (*Module)(unsafe.Pointer(args.result)), (*Error)(unsafe.Pointer(args.error))
}
