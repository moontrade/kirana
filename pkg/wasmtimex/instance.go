package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasmtime_instance_new_t {
	size_t context;
	size_t module;
	size_t imports;
	size_t imports_count;
	size_t instance;
	size_t trap;
	size_t error;
} do_wasmtime_instance_new_t;

void do_wasmtime_instance_new(size_t arg0, size_t arg1) {
	do_wasmtime_instance_new_t* args = (do_wasmtime_instance_new_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_instance_new(
		(wasmtime_context_t*)args->context,
		(const wasmtime_module_t*)args->module,
		(const wasmtime_extern_t*)args->imports,
		args->imports_count,
		(wasmtime_instance_t*)args->instance,
		(wasm_trap_t**)(void*)&args->trap
	);
}

typedef struct do_wasmtime_instance_export_get_t {
	size_t context;
	size_t instance;
	size_t name;
	size_t name_len;
	size_t item;
	size_t ok;
} do_wasmtime_instance_export_get_t;

void do_wasmtime_instance_export_get(size_t arg0, size_t arg1) {
	do_wasmtime_instance_export_get_t* args = (do_wasmtime_instance_export_get_t*)(void*)arg0;
	args->ok = wasmtime_instance_export_get(
		(wasmtime_context_t*)args->context,
		(const wasmtime_instance_t*)args->instance,
		(const char*)args->name,
		args->name_len,
		(wasmtime_extern_t*)args->item
	) ? 1 : 0;
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

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
func NewInstance(ctx *Context, module *Module, imports ...Extern) (instance Instance, trap *Trap, err *Error) {
	/*
		size_t context;
			size_t module;
			size_t imports;
			size_t imports_count;
			size_t instance;
			size_t trap;
			size_t error;
	*/
	args := struct {
		context      uintptr
		module       uintptr
		imports      uintptr
		importsCount uintptr
		instance     uintptr
		trap         uintptr
		error        uintptr
	}{
		context:      uintptr(unsafe.Pointer(ctx)),
		module:       uintptr(unsafe.Pointer(module)),
		imports:      dataPtr(imports),
		importsCount: uintptr(len(imports)),
		instance:     uintptr(unsafe.Pointer(&instance)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_instance_new), uintptr(unsafe.Pointer(&args)), 0)
	return instance, (*Trap)(unsafe.Pointer(args.trap)), (*Error)(unsafe.Pointer(args.error))
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
func (inst *Instance) ExportNamed(ctx *Context, name string) (item Extern, ok bool) {
	args := struct {
		context  uintptr
		instance uintptr
		name     uintptr
		nameLen  uintptr
		item     uintptr
		ok       uintptr
	}{
		context:  uintptr(unsafe.Pointer(ctx)),
		instance: uintptr(unsafe.Pointer(inst)),
		name:     strDataPtr(name),
		nameLen:  uintptr(len(name)),
		item:     uintptr(unsafe.Pointer(&item)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_instance_export_get), uintptr(unsafe.Pointer(&args)), 0)
	return item, args.ok != 0
}

//func (inst *Instance) GetFunc(ctx *Context, name string) (Func, bool) {
//	item, ok := inst.ExportNamed(ctx, name)
//	if !ok {
//		return Func{}, false
//	}
//}

//// ExportAt gets an export by index from an instance.
////
//// \param store the store that owns `instance`
//// \param instance the instance to lookup within
//// \param index the index to lookup
//// \param name where to store the name of the export
//// \param name_len where to store the byte length of the name
//// \param item where to store the export itself
////
//// Returns nonzero if the export was found, and `name`, `name_len`, and `item`
//// are filled in. Otherwise, returns 0.
////
//// Doesn't take ownership of any arguments but does return ownership of the
//// #wasmtime_extern_t. The `name` pointer return value is owned by the `store`
//// and must be immediately used before calling any other APIs on
//// #wasmtime_context_t.
//func (inst *Instance) ExportAt(index int) (bool, ByteVec, *Extern) {
//	return false, ByteVec{}, nil
//}
