package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

//typedef struct do_wasmtime_memorytype_new_t {
//	uint64_t min;
//	uint64_t max;
//	uint32_t max_present;
//	uint32_t is_64;
//	size_t result;
//} do_wasmtime_memorytype_new_t;
//
//void do_wasmtime_memorytype_new(size_t arg0, size_t arg1) {
//	do_wasmtime_memorytype_new_t* args = (do_wasmtime_memorytype_new_t*)(void*)arg0;
//	args->result = (size_t)(void*)wasmtime_memorytype_new(
//		args->min,
//		args->max_present != 0,
//		args->max,
//		args->is_64 != 0
//	);
//}

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
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
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
func NewModule(engine *Engine, wasm []byte) (*Module, *Error) {
	return nil, nil
}

func (m *Module) Delete() {
	cgo.NonBlocking((*byte)(C.wasmtime_module_delete), uintptr(unsafe.Pointer(m)), 0)
}

func (m *Module) Clone() *Module {
	var clone *Module
	cgo.NonBlocking((*byte)(C.wasmtime_module_clone), uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(&clone)))
	return clone
}
