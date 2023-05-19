package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

void do_wasmtime_linker_new(size_t arg0, size_t arg1) {
	*((wasmtime_linker_t**)(void*)arg1) = wasmtime_linker_new(
		(wasm_engine_t*)(void*)arg0
	);
}

void do_wasmtime_linker_delete(size_t arg0, size_t arg1) {
	wasmtime_linker_delete(
		(wasmtime_linker_t*)(void*)arg0
	);
}

void do_wasmtime_linker_allow_shadowing(size_t arg0, size_t arg1) {
	wasmtime_linker_allow_shadowing(
		(wasmtime_linker_t*)(void*)arg0,
		arg1 != 0
	);
}

typedef struct do_wasmtime_linker_define_t {
	size_t linker;
	size_t store;
	size_t module;
	size_t module_len;
	size_t name;
	size_t name_len;
	size_t item;
	size_t error;
} do_wasmtime_linker_define_t;

void do_wasmtime_linker_define(size_t arg0, size_t arg1) {
	do_wasmtime_linker_define_t* args = (do_wasmtime_linker_define_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_linker_define(
		(wasmtime_linker_t*)args->linker,
		(wasmtime_context_t*)args->store,
		(const char*)args->module,
		args->module_len,
		(const char*)args->name,
		args->name_len,
		(const wasmtime_extern_t*)args->item
	);
}

typedef struct do_wasmtime_linker_define_func_t {
	size_t linker;
	size_t module;
	size_t module_len;
	size_t name;
	size_t name_len;
	size_t func_type;
	size_t callback;
	size_t env;
	size_t finalizer;
	size_t error;
} do_wasmtime_linker_define_func_t;

void do_wasmtime_linker_define_func(size_t arg0, size_t arg1) {
	do_wasmtime_linker_define_func_t* args = (do_wasmtime_linker_define_func_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_linker_define_func(
		(wasmtime_linker_t*)(void*)args->linker,
		(const char*)args->module,
		args->module_len,
		(const char*)args->name,
		args->name_len,
		(const wasm_functype_t*)(void*)args->func_type,
		(wasmtime_func_callback_t)(void*)args->callback,
		(void*)args->env,
		(void (*)(void*))(void*)args->finalizer
	);
}

typedef struct do_wasmtime_linker_define_func_unchecked_t {
	size_t linker;
	size_t module;
	size_t module_len;
	size_t name;
	size_t name_len;
	size_t func_type;
	size_t callback;
	size_t env;
	size_t finalizer;
	size_t error;
} do_wasmtime_linker_define_func_unchecked_t;

void do_wasmtime_linker_define_func_unchecked(size_t arg0, size_t arg1) {
	do_wasmtime_linker_define_func_unchecked_t* args = (do_wasmtime_linker_define_func_unchecked_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_linker_define_func_unchecked(
		(wasmtime_linker_t*)(void*)args->linker,
		(const char*)args->module,
		args->module_len,
		(const char*)args->name,
		args->name_len,
		(const wasm_functype_t*)(void*)args->func_type,
		(wasmtime_func_unchecked_callback_t)(void*)args->callback,
		(void*)args->env,
		(void (*)(void*))(void*)args->finalizer
	);
}

void do_wasmtime_linker_define_wasi(size_t arg0, size_t arg1) {
	*((wasmtime_error_t**)arg1) = wasmtime_linker_define_wasi(
		(wasmtime_linker_t*)(void*)arg0
	);
}

typedef struct do_wasmtime_linker_define_instance_t {
	size_t linker;
	size_t context;
	size_t name;
	size_t name_len;
	size_t instance;
	size_t error;
} do_wasmtime_linker_define_instance_t;

void do_wasmtime_linker_define_instance(size_t arg0, size_t arg1) {
	do_wasmtime_linker_define_instance_t* args = (do_wasmtime_linker_define_instance_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_linker_define_instance(
		(wasmtime_linker_t*)(void*)args->linker,
		(wasmtime_context_t*)(void*)args->context,
		(const char*)args->name,
		args->name_len,
		(const wasmtime_instance_t*)(void*)args->instance
	);
}

typedef struct do_wasmtime_linker_instantiate_t {
	size_t linker;
	size_t context;
	size_t module;
	size_t instance;
	size_t trap;
	size_t error;
} do_wasmtime_linker_instantiate_t;

void do_wasmtime_linker_instantiate(size_t arg0, size_t arg1) {
	do_wasmtime_linker_instantiate_t* args = (do_wasmtime_linker_instantiate_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_linker_instantiate(
		(const wasmtime_linker_t*)(void*)args->linker,
		(wasmtime_context_t*)(void*)args->context,
		(const wasmtime_module_t*)(void*)args->module,
		(wasmtime_instance_t*)(void*)args->instance,
		(wasm_trap_t**)(void*)&args->trap
	);
}

typedef struct do_wasmtime_linker_module_t {
	size_t linker;
	size_t context;
	size_t name;
	size_t name_len;
	size_t module;
	size_t error;
} do_wasmtime_linker_module_t;

void do_wasmtime_linker_module(size_t arg0, size_t arg1) {
	do_wasmtime_linker_module_t* args = (do_wasmtime_linker_module_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_linker_module(
		(wasmtime_linker_t*)(void*)args->linker,
		(wasmtime_context_t*)(void*)args->context,
		(const char*)args->name,
		args->name_len,
		(const wasmtime_module_t*)(void*)args->module
	);
}

typedef struct do_wasmtime_linker_get_default_t {
	size_t linker;
	size_t context;
	size_t name;
	size_t name_len;
	size_t func;
	size_t error;
} do_wasmtime_linker_get_default_t;

void do_wasmtime_linker_get_default(size_t arg0, size_t arg1) {
	do_wasmtime_linker_get_default_t* args = (do_wasmtime_linker_get_default_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_linker_get_default(
		(wasmtime_linker_t*)(void*)args->linker,
		(wasmtime_context_t*)(void*)args->context,
		(const char*)args->name,
		args->name_len,
		(wasmtime_func_t*)(void*)args->func
	);
}

typedef struct do_wasmtime_linker_get_t {
	size_t linker;
	size_t context;
	size_t module;
	size_t module_len;
	size_t name;
	size_t name_len;
	size_t item;
	size_t ok;
} do_wasmtime_linker_get_t;

void do_wasmtime_linker_get(size_t arg0, size_t arg1) {
	do_wasmtime_linker_get_t* args = (do_wasmtime_linker_get_t*)(void*)arg0;
	args->ok = (size_t)(void*)wasmtime_linker_get(
		(wasmtime_linker_t*)(void*)args->linker,
		(wasmtime_context_t*)(void*)args->context,
		(const char*)args->module,
		args->module_len,
		(const char*)args->name,
		args->name_len,
		(wasmtime_extern_t*)(void*)args->item
	) ? 1 : 0;
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

// Linker
// \typedef wasmtime_linker_t
// \brief Alias to #wasmtime_linker
//
// \struct #wasmtime_linker
// \brief Object used to conveniently link together and instantiate wasm
// modules.
//
// This type corresponds to the `wasmtime::Linker` type in Rust. This
// type is intended to make it easier to manage a set of modules that link
// together, or to make it easier to link WebAssembly modules to WASI.
//
// A #wasmtime_linker_t is a higher level way to instantiate a module than
// #wasm_instance_new since it works at the "string" level of imports rather
// than requiring 1:1 mappings.
type Linker C.wasmtime_linker_t

// NewLinker creates a new linker for the specified engine.
//
// This function does not take ownership of the engine argument, and the caller
// is expected to delete the returned linker.
func NewLinker(engine *Engine) *Linker {
	var linker uintptr
	cgo.NonBlocking((*byte)(C.do_wasmtime_linker_new), uintptr(unsafe.Pointer(engine)), uintptr(unsafe.Pointer(&linker)))
	return (*Linker)(unsafe.Pointer(linker))
}

func (l *Linker) Delete() {
	cgo.NonBlocking((*byte)(C.do_wasmtime_linker_delete), uintptr(unsafe.Pointer(l)), 0)
}

// AllowShadowing configures whether this linker allows later definitions to shadow
// previous definitions.
//
// By default this setting is `false`.
func (l *Linker) AllowShadowing(allow bool) {
	arg := uintptr(0)
	if allow {
		arg = 1
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_linker_allow_shadowing), uintptr(unsafe.Pointer(l)), arg)
}

// Define a new item in this linker.
//
// \param linker the linker the name is being defined in.
// \param module the module name the item is defined under.
// \param module_len the byte length of `module`
// \param name the field name the item is defined under
// \param name_len the byte length of `name`
// \param item the item that is being defined in this linker.
//
// \return On success `NULL` is returned, otherwise an error is returned which
// describes why the definition failed.
//
// For more information about name resolution consult the [Rust
// documentation](https://bytecodealliance.github.io/wasmtime/api/wasmtime/struct.Linker.html#name-resolution).
func (l *Linker) Define(module, name string, store *Context, item *Extern) *Error {
	args := struct {
		linker    uintptr
		store     uintptr
		module    uintptr
		moduleLen uintptr
		name      uintptr
		nameLen   uintptr
		item      uintptr
		error     uintptr
	}{
		linker:    uintptr(unsafe.Pointer(l)),
		store:     uintptr(unsafe.Pointer(store)),
		module:    strDataPtr(module),
		moduleLen: uintptr(len(module)),
		name:      strDataPtr(name),
		nameLen:   uintptr(len(name)),
		item:      uintptr(unsafe.Pointer(item)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_linker_define), uintptr(unsafe.Pointer(&args)), 0)
	return (*Error)(unsafe.Pointer(args.error))
}

// DefineFunc defines a new function in this linker.
//
// \param linker the linker the name is being defined in.
// \param module the module name the item is defined under.
// \param module_len the byte length of `module`
// \param name the field name the item is defined under
// \param name_len the byte length of `name`
// \param ty the type of the function that's being defined
// \param cb the host callback to invoke when the function is called
// \param data the host-provided data to provide as the first argument to the callback
// \param finalizer an optional finalizer for the `data` argument.
//
// \return On success `NULL` is returned, otherwise an error is returned which
// describes why the definition failed.
//
// For more information about name resolution consult the [Rust
// documentation](https://bytecodealliance.github.io/wasmtime/api/wasmtime/struct.Linker.html#name-resolution).
//
// Note that this function does not create a #wasmtime_func_t. This creates a
// store-independent function within the linker, allowing this function
// definition to be used with multiple stores.
//
// For more information about host callbacks see #wasmtime_func_new.
func (l *Linker) DefineFunc(
	module, name string,
	funcType *FuncType,
	callback Callback,
	env uintptr,
	finalizer FuncFinalizer) *Error {
	/*
		size_t linker;
		size_t module;
		size_t module_len;
		size_t name;
		size_t name_len;
		size_t func_type;
		size_t callback;
		size_t env;
		size_t finalizer;
		size_t error;
	*/
	args := struct {
		linker    uintptr
		module    uintptr
		moduleLen uintptr
		name      uintptr
		nameLen   uintptr
		funcType  uintptr
		callback  uintptr
		env       uintptr
		finalizer uintptr
		error     uintptr
	}{
		linker:    uintptr(unsafe.Pointer(l)),
		module:    strDataPtr(module),
		moduleLen: uintptr(len(module)),
		name:      strDataPtr(name),
		nameLen:   uintptr(len(name)),
		funcType:  uintptr(unsafe.Pointer(funcType)),
		callback:  uintptr(unsafe.Pointer(callback)),
		env:       env,
		finalizer: uintptr(unsafe.Pointer(finalizer)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_linker_define_func), uintptr(unsafe.Pointer(&args)), 0)
	return (*Error)(unsafe.Pointer(args.error))
}

// DefineFuncUnchecked Defines a new function in this linker.
//
// This is the same as #wasmtime_linker_define_func except that it's the analog
// of #wasmtime_func_new_unchecked instead of #wasmtime_func_new. Be sure to
// consult the documentation of #wasmtime_linker_define_func for argument
// information as well as #wasmtime_func_new_unchecked for why this is an
// unsafe API.
func (l *Linker) DefineFuncUnchecked(
	module, name string,
	funcType *FuncType,
	callback UncheckedCallback,
	env uintptr,
	finalizer FuncFinalizer) *Error {
	/*
		size_t linker;
		size_t module;
		size_t module_len;
		size_t name;
		size_t name_len;
		size_t func_type;
		size_t callback;
		size_t env;
		size_t finalizer;
		size_t error;
	*/
	args := struct {
		linker    uintptr
		module    uintptr
		moduleLen uintptr
		name      uintptr
		nameLen   uintptr
		funcType  uintptr
		callback  uintptr
		env       uintptr
		finalizer uintptr
		error     uintptr
	}{
		linker:    uintptr(unsafe.Pointer(l)),
		module:    strDataPtr(module),
		moduleLen: uintptr(len(module)),
		name:      strDataPtr(name),
		nameLen:   uintptr(len(name)),
		funcType:  uintptr(unsafe.Pointer(funcType)),
		callback:  uintptr(unsafe.Pointer(callback)),
		env:       env,
		finalizer: uintptr(unsafe.Pointer(finalizer)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_linker_define_func_unchecked), uintptr(unsafe.Pointer(&args)), 0)
	return (*Error)(unsafe.Pointer(args.error))
}

// DefineWasi defines WASI functions in this linker.
//
// \param linker the linker the name is being defined in.
//
// \return On success `NULL` is returned, otherwise an error is returned which
// describes why the definition failed.
//
// This function will provide WASI function names in the specified linker. Note
// that when an instance is created within a store then the store also needs to
// have its WASI settings configured with #wasmtime_context_set_wasi for WASI
// functions to work, otherwise an assert will be tripped that will abort the
// process.
//
// For more information about name resolution consult the [Rust
// documentation](https://bytecodealliance.github.io/wasmtime/api/wasmtime/struct.Linker.html#name-resolution).
func (l *Linker) DefineWasi() *Error {
	var err uintptr
	cgo.NonBlocking((*byte)(C.do_wasmtime_linker_define_wasi), uintptr(unsafe.Pointer(l)), uintptr(unsafe.Pointer(&err)))
	return (*Error)(unsafe.Pointer(err))
}

// DefineInstance defines an instance under the specified name in this linker.
//
// \param linker the linker the name is being defined in.
// \param store the store that owns `instance`
// \param name the module name to define `instance` under.
// \param name_len the byte length of `name`
// \param instance a previously-created instance.
//
// \return On success `NULL` is returned, otherwise an error is returned which
// describes why the definition failed.
//
// This function will take all of the exports of the `instance` provided and
// defined them under a module called `name` with a field name as the export's
// own name.
//
// For more information about name resolution consult the [Rust
// documentation](https://bytecodealliance.github.io/wasmtime/api/wasmtime/struct.Linker.html#name-resolution).
func (l *Linker) DefineInstance(ctx *Context, name string, instance *Instance) *Error {
	args := struct {
		linker   uintptr
		context  uintptr
		name     uintptr
		nameLen  uintptr
		instance uintptr
		error    uintptr
	}{
		linker:   uintptr(unsafe.Pointer(l)),
		context:  uintptr(unsafe.Pointer(ctx)),
		name:     strDataPtr(name),
		nameLen:  uintptr(len(name)),
		instance: uintptr(unsafe.Pointer(instance)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_linker_define_instance), uintptr(unsafe.Pointer(&args)), 0)
	return (*Error)(unsafe.Pointer(args.error))
}

// Instantiate instantiates a #wasm_module_t with the items defined in this linker.
//
// \param linker the linker used to instantiate the provided module.
// \param store the store that is used to instantiate within
// \param module the module that is being instantiated.
// \param instance the returned instance, if successful.
// \param trap a trap returned, if the start function traps.
//
// \return One of three things can happen as a result of this function. First
// the module could be successfully instantiated and returned through
// `instance`, meaning the return value and `trap` are both set to `NULL`.
// Second the start function may trap, meaning the return value and `instance`
// are set to `NULL` and `trap` describes the trap that happens. Finally
// instantiation may fail for another reason, in which case an error is returned
// and `trap` and `instance` are set to `NULL`.
//
// This function will attempt to satisfy all of the imports of the `module`
// provided with items previously defined in this linker. If any name isn't
// defined in the linker than an error is returned. (or if the previously
// defined item is of the wrong type).
func (l *Linker) Instantiate(ctx *Context, module *Module) (Instance, *Trap, *Error) {
	var instance Instance
	args := struct {
		linker   uintptr
		context  uintptr
		module   uintptr
		instance uintptr
		trap     uintptr
		error    uintptr
	}{
		linker:   uintptr(unsafe.Pointer(l)),
		context:  uintptr(unsafe.Pointer(ctx)),
		module:   uintptr(unsafe.Pointer(module)),
		instance: uintptr(unsafe.Pointer(&instance)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_linker_instantiate), uintptr(unsafe.Pointer(&args)), 0)
	return instance, (*Trap)(unsafe.Pointer(args.trap)), (*Error)(unsafe.Pointer(args.error))
}

// Module defines automatic instantiations of a #wasm_module_t in this linker.
//
// \param linker the linker the module is being added to
// \param store the store that is used to instantiate `module`
// \param name the name of the module within the linker
// \param name_len the byte length of `name`
// \param module the module that's being instantiated
//
// \return An error if the module could not be instantiated or added or `NULL`
// on success.
//
// This function automatically handles [Commands and
// Reactors](https://github.com/WebAssembly/WASI/blob/master/design/application-abi.md#current-unstable-abi)
// instantiation and initialization.
//
// For more information see the [Rust
// documentation](https://bytecodealliance.github.io/wasmtime/api/wasmtime/struct.Linker.html#method.module).
func (l *Linker) Module(ctx *Context, name string, module *Module) *Error {
	args := struct {
		linker  uintptr
		context uintptr
		name    uintptr
		nameLen uintptr
		module  uintptr
		error   uintptr
	}{
		linker:  uintptr(unsafe.Pointer(l)),
		context: uintptr(unsafe.Pointer(ctx)),
		name:    strDataPtr(name),
		nameLen: uintptr(len(name)),
		module:  uintptr(unsafe.Pointer(module)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_linker_module), uintptr(unsafe.Pointer(&args)), 0)
	return (*Error)(unsafe.Pointer(args.error))
}

// GetDefault acquires the "default export" of the named module in this linker.
//
// \param linker the linker to load from
// \param store the store to load a function into
// \param name the name of the module to get the default export for
// \param name_len the byte length of `name`
// \param func where to store the extracted default function.
//
// \return An error is returned if the default export could not be found, or
// `NULL` is returned and `func` is filled in otherwise.
//
// For more information see the [Rust
// documentation](https://bytecodealliance.github.io/wasmtime/api/wasmtime/struct.Linker.html#method.get_default).
func (l *Linker) GetDefault(ctx *Context, name string) (Func, *Error) {
	var fn Func
	args := struct {
		linker  uintptr
		context uintptr
		name    uintptr
		nameLen uintptr
		fn      uintptr
		error   uintptr
	}{
		linker:  uintptr(unsafe.Pointer(l)),
		context: uintptr(unsafe.Pointer(ctx)),
		name:    strDataPtr(name),
		nameLen: uintptr(len(name)),
		fn:      uintptr(unsafe.Pointer(&fn)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_linker_get_default), uintptr(unsafe.Pointer(&args)), 0)
	return fn, (*Error)(unsafe.Pointer(args.error))
}

// Get loads an item by name from this linker.
//
// \param linker the linker to load from
// \param store the store to load the item into
// \param module the name of the module to get
// \param module_len the byte length of `module`
// \param name the name of the field to get
// \param name_len the byte length of `name`
// \param item where to store the extracted item
//
// \return A nonzero value if the item is defined, in which case `item` is also
// filled in. Otherwise, zero is returned.
func (l *Linker) Get(ctx *Context, module, name string) (Extern, bool) {
	var item Extern
	args := struct {
		linker    uintptr
		context   uintptr
		module    uintptr
		moduleLen uintptr
		name      uintptr
		nameLen   uintptr
		item      uintptr
		ok        uintptr
	}{
		linker:    uintptr(unsafe.Pointer(l)),
		context:   uintptr(unsafe.Pointer(ctx)),
		module:    strDataPtr(module),
		moduleLen: uintptr(len(module)),
		name:      strDataPtr(name),
		nameLen:   uintptr(len(name)),
		item:      uintptr(unsafe.Pointer(&item)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_linker_get), uintptr(unsafe.Pointer(&args)), 0)
	return item, args.ok != 0
}
