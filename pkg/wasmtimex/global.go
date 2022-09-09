package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasmtime_global_new_t {
	size_t context;
	size_t global_type;
	size_t val;
	size_t result;
	size_t error;
} do_wasmtime_global_new_t;

void do_wasmtime_global_new(size_t arg0, size_t arg1) {
	do_wasmtime_global_new_t* args = (do_wasmtime_global_new_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_global_new(
		(wasmtime_context_t*)(void*)args->context,
		(const wasm_globaltype_t*)(void*)args->global_type,
		(const wasmtime_val_t*)(void*)args->val,
		(wasmtime_global_t*)(void*)&args->result
	);
}

typedef struct do_wasmtime_global_type_t {
	size_t context;
	size_t global;
	size_t result;
} do_wasmtime_global_type_t;

void do_wasmtime_global_type(size_t arg0, size_t arg1) {
	do_wasmtime_global_type_t* args = (do_wasmtime_global_type_t*)(void*)arg0;
	args->result = (size_t)(void*)wasmtime_global_type(
		(const wasmtime_context_t*)(void*)args->context,
		(const wasmtime_global_t*)(void*)args->global
	);
}

typedef struct do_wasmtime_global_get_t {
	size_t context;
	size_t global;
	size_t val;
} do_wasmtime_global_get_t;

void do_wasmtime_global_get(size_t arg0, size_t arg1) {
	do_wasmtime_global_get_t* args = (do_wasmtime_global_get_t*)(void*)arg0;
	wasmtime_global_get(
		(wasmtime_context_t*)(void*)args->context,
		(const wasmtime_global_t*)(void*)args->global,
		(wasmtime_val_t*)(void*)args->val
	);
}

typedef struct do_wasmtime_global_set_t {
	size_t context;
	size_t global;
	size_t val;
	size_t error;
} do_wasmtime_global_set_t;

void do_wasmtime_global_set(size_t arg0, size_t arg1) {
	do_wasmtime_global_set_t* args = (do_wasmtime_global_set_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_global_set(
		(wasmtime_context_t*)(void*)args->context,
		(const wasmtime_global_t*)(void*)args->global,
		(const wasmtime_val_t*)(void*)args->val
	);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

// Global is a global instance, which is the runtime representation of a global variable.
// It holds an individual value and a flag indicating whether it is mutable.
// Read more in [spec](https://webassembly.github.io/spec/core/exec/runtime.html#global-instances)
type Global C.wasmtime_global_t

// NewGlobal
// \brief Creates a new global value.
//
// Creates a new host-defined global value within the provided `store`
//
// \param store the store in which to create the global
// \param type the wasm type of the global being created
// \param val the initial value of the global
// \param ret a return pointer for the created global.
//
// This function may return an error if the `val` argument does not match the
// specified type of the global, or if `val` comes from a different store than
// the one provided.
//
// This function does not take ownership of its arguments but error is
// owned by the caller.
func NewGlobal(ctx *Context, ty *GlobalType, val Val) (*Global, *Error) {
	args := struct {
		context    uintptr
		globalType uintptr
		val        uintptr
		result     uintptr
		error      uintptr
	}{
		context:    uintptr(unsafe.Pointer(ctx)),
		globalType: uintptr(unsafe.Pointer(ty)),
		val:        uintptr(unsafe.Pointer(&val)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_global_new), uintptr(unsafe.Pointer(&args)), 0)
	if args.error != 0 {
		return nil, (*Error)(unsafe.Pointer(args.error))
	}
	return (*Global)(unsafe.Pointer(args.result)), nil
}

// Type returns the type of this global.
func (g *Global) Type(ctx *Context) *GlobalType {
	args := struct {
		context uintptr
		global  uintptr
		result  uintptr
	}{
		context: uintptr(unsafe.Pointer(ctx)),
		global:  uintptr(unsafe.Pointer(g)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_global_type), uintptr(unsafe.Pointer(&args)), 0)
	return (*GlobalType)(unsafe.Pointer(args.result))
}

// Get the value of the specified global.
//
// \param store the store that owns `global`
// \param global the global to get
// \param out where to store the value in this global.
//
// This function returns ownership of the contents of `out`, so
// #wasmtime_val_delete may need to be called on the value.
func (g *Global) Get(ctx *Context) (val Val) {
	args := struct {
		context uintptr
		global  uintptr
		val     uintptr
	}{
		context: uintptr(unsafe.Pointer(ctx)),
		global:  uintptr(unsafe.Pointer(g)),
		val:     uintptr(unsafe.Pointer(&val)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_global_get), uintptr(unsafe.Pointer(&args)), 0)
	return
}

// Set a global to a new value.
//
// \param store the store that owns `global`
// \param global the global to set
// \param val the value to store in the global
//
// This function may return an error if `global` is not mutable or if `val` has
// the wrong type for `global`.
//
// This does not take ownership of any argument but returns ownership of the error.
func (g *Global) Set(ctx *Context, val Val) *Error {
	args := struct {
		context uintptr
		global  uintptr
		val     uintptr
		error   uintptr
	}{
		context: uintptr(unsafe.Pointer(ctx)),
		global:  uintptr(unsafe.Pointer(g)),
		val:     uintptr(unsafe.Pointer(&val)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_global_set), uintptr(unsafe.Pointer(&args)), 0)
	return (*Error)(unsafe.Pointer(args.error))
}

func (g *Global) AsExtern() Extern {
	var ret Extern
	ret.SetGlobal(*g)
	return ret
}
