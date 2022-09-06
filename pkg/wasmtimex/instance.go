package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>
*/
import "C"
import "unsafe"

// Instance is a representation of an instance in Wasmtime.
//
// Instances are represented with a 64-bit identifying integer in Wasmtime.
// They do not have any destructor associated with them. Instances cannot
// interoperate between #wasmtime_store_t instances and if the wrong instance
// is passed to the wrong store then it may trigger an assertion to abort the
// process.
type Instance struct {
	storeId uint64
	index   uintptr
}

func (inst *Instance) ptr() *C.wasmtime_instance_t {
	return (*C.wasmtime_instance_t)(unsafe.Pointer(inst))
}

// NewInstance instantiates a wasm module.
//
// This function will instantiate a WebAssembly module with the provided
// imports, creating a WebAssembly instance. The returned instance can then
// afterwards be inspected for exports.
//
// \param store the store in which to create the instance
// \param module the module that's being instantiated
// \param imports the imports provided to the module
// \param nimports the size of `imports`
// \param instance where to store the returned instance
// \param trap where to store the returned trap
//
// This function requires that `imports` is the same size as the imports that
// `module` has. Additionally, the `imports` array must be 1:1 lined up with the
// imports of the `module` specified. This is intended to be relatively low
// level, and #wasmtime_linker_instantiate is provided for a more ergonomic
// name-based resolution API.
//
// The states of return values from this function are similar to
// #wasmtime_func_call where an error can be returned meaning something like a
// link error in this context. A trap can be returned (meaning no error or
// instance is returned), or an instance can be returned (meaning no error or
// trap is returned).
//
// Note that this function requires that all `imports` specified must be owned
// by the `store` provided as well.
//
// This function does not take ownership of its arguments, but all return
// values are owned by the caller.
func NewInstance(ctx *Context, module *Module, imports []*Extern) (*Instance, *Trap, *Error) {
	return nil, nil, nil
}

// ExportNamed gets an export by name from an instance.
//
// \param store the store that owns `instance`
// \param instance the instance to lookup within
// \param name the export name to lookup
// \param name_len the byte length of `name`
// \param item where to store the returned value
//
// Returns nonzero if the export was found, and `item` is filled in. Otherwise
// returns 0.
//
// Doesn't take ownership of any arguments but does return ownership of the
// #wasmtime_extern_t.
// /
func (inst *Instance) ExportNamed(name string) *Extern {
	return nil
}

// ExportAt gets an export by index from an instance.
//
// \param store the store that owns `instance`
// \param instance the instance to lookup within
// \param index the index to lookup
// \param name where to store the name of the export
// \param name_len where to store the byte length of the name
// \param item where to store the export itself
//
// Returns nonzero if the export was found, and `name`, `name_len`, and `item`
// are filled in. Otherwise returns 0.
//
// Doesn't take ownership of any arguments but does return ownership of the
// #wasmtime_extern_t. The `name` pointer return value is owned by the `store`
// and must be immediately used before calling any other APIs on
// #wasmtime_context_t.
func (inst *Instance) ExportAt(index int) (bool, ByteVec, *Extern) {
	return false, ByteVec{}, nil
}
