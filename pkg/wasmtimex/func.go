package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasmtime_func_new_t {
	size_t context;
	size_t func_type;
	size_t callback;
	size_t env;
	size_t finalizer;
	size_t ret;
} do_wasmtime_func_new_t;

void do_wasmtime_func_new(size_t arg0, size_t arg1) {
	do_wasmtime_func_new_t* args = (do_wasmtime_func_new_t*)(void*)arg0;
	wasmtime_func_new(
		(wasmtime_context_t*)(void*)args->context,
		(const wasm_functype_t*)(void*)args->func_type,
		(wasmtime_func_callback_t)(void*)args->callback,
		(void*)args->env,
		(void (*)(void*))(void*)args->finalizer,
		(wasmtime_func_t*)(void*)args->ret
	);
}

typedef struct do_wasmtime_func_new_unchecked_t {
	size_t context;
	size_t func_type;
	size_t callback;
	size_t env;
	size_t finalizer;
	size_t ret;
} do_wasmtime_func_new_unchecked_t;

void do_wasmtime_func_new_unchecked(size_t arg0, size_t arg1) {
	do_wasmtime_func_new_unchecked_t* args = (do_wasmtime_func_new_unchecked_t*)(void*)arg0;
	wasmtime_func_new_unchecked(
		(wasmtime_context_t*)(void*)args->context,
		(const wasm_functype_t*)(void*)args->func_type,
		(wasmtime_func_unchecked_callback_t)(void*)args->callback,
		(void*)args->env,
		(void (*)(void*))(void*)args->finalizer,
		(wasmtime_func_t*)(void*)args->ret
	);
}

typedef struct do_wasmtime_func_type_t {
	size_t context;
	size_t func;
	size_t ret;
} do_wasmtime_func_type_t;

void do_wasmtime_func_type(size_t arg0, size_t arg1) {
	do_wasmtime_func_type_t* args = (do_wasmtime_func_type_t*)(void*)arg0;
	args->ret = (size_t)(void*)wasmtime_func_type(
		(const wasmtime_context_t*)(void*)args->context,
		(const wasmtime_func_t*)(void*)args->func
	);
}

void do_wasmtime_caller_context(size_t arg0, size_t arg1) {
	*((wasmtime_context_t**)arg1) = wasmtime_caller_context(
		(wasmtime_caller_t*)(void*)arg0
	);
}

typedef struct do_wasmtime_func_call_unchecked_t {
	size_t context;
	size_t func;
	size_t args;
	size_t trap;
} do_wasmtime_func_call_unchecked_t;

void do_wasmtime_func_call_unchecked(size_t arg0, size_t arg1) {
	do_wasmtime_func_call_unchecked_t* args = (do_wasmtime_func_call_unchecked_t*)(void*)arg0;
	args->trap = (size_t)(void*)wasmtime_func_call_unchecked(
		(wasmtime_context_t*)(void*)args->context,
		(const wasmtime_func_t*)(void*)args->func,
		(wasmtime_val_raw_t*)(void*)args->args,
		(wasm_trap_t**)(void*)args->trap
	);
}

typedef struct do_wasmtime_func_call_t {
	size_t context;
	size_t func;
	size_t args;
	size_t args_len;
	size_t results;
	size_t results_len;
	size_t trap;
	size_t error;
} do_wasmtime_func_call_t;

void do_wasmtime_func_call(size_t arg0, size_t arg1) {
	do_wasmtime_func_call_t* args = (do_wasmtime_func_call_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_func_call(
		(wasmtime_context_t*)(void*)args->context,
		(const wasmtime_func_t*)(void*)args->func,
		(const wasmtime_val_t*)(void*)args->args,
		args->args_len,
		(wasmtime_val_t*)(void*)args->results,
		args->results_len,
		(wasm_trap_t**)(void*)&args->trap
	);
}

//typedef wasm_trap_t* (*wasmtime_func_callback_t)(
//    void *env,
//    wasmtime_caller_t* caller,
//    const wasmtime_val_t *args,
//    size_t nargs,
//    wasmtime_val_t *results,
//    size_t nresults);

wasm_trap_t* *wasmtime_func_callback_stub(
    void *env,
    wasmtime_caller_t* caller,
    const wasmtime_val_t *args,
    size_t nargs,
    wasmtime_val_t *results,
    size_t nresults) { return NULL; }

//typedef wasm_trap_t* (*wasmtime_func_unchecked_callback_t)(
//    void *env,
//    wasmtime_caller_t* caller,
//    wasmtime_val_raw_t *args_and_results,
//    size_t num_args_and_results);

wasm_trap_t* wasmtime_func_unchecked_callback_stub(void *env,
    wasmtime_caller_t* caller,
    wasmtime_val_raw_t *args_and_results,
    size_t num_args_and_results
) { return NULL; }


typedef void (*wasmtime_func_finalizer_t)(void *env);

*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

type FuncFinalizer C.wasmtime_func_finalizer_t

// Callback signature for #wasmtime_func_new.
//
// This is the function signature for host functions that can be made accessible
// to WebAssembly. The arguments to this function are:
//
// \param env user-provided argument passed to #wasmtime_func_new
// \param caller a temporary object that can only be used during this function
// call. Used to acquire #wasmtime_context_t or caller's state
// \param args the arguments provided to this function invocation
// \param nargs how many arguments are provided
// \param results where to write the results of this function
// \param nresults how many results must be produced
//
// Callbacks are guaranteed to get called with the right types of arguments, but
// they must produce the correct number and types of results. Failure to do so
// will cause traps to get raised on the wasm side.
//
// This callback can optionally return a #wasm_trap_t indicating that a trap
// should be raised in WebAssembly. It's expected that in this case the caller
// relinquishes ownership of the trap, and it is passed back to the engine.
type Callback C.wasmtime_func_callback_t

var CallbackStub = Callback(C.wasmtime_func_callback_stub)

// UncheckedCallback callback signature for #wasmtime_func_new_unchecked.
//
// This is the function signature for host functions that can be made accessible
// to WebAssembly. The arguments to this function are:
//
// \param env user-provided argument passed to #wasmtime_func_new_unchecked
//
// \param caller a temporary object that can only be used during this function
//
//	call. Used to acquire #wasmtime_context_t or caller's state
//
// \param args_and_results storage space for both the parameters to the
//
//	function and the results of the function. The size of this
//	array depends on the function type that the host function is created
//	with, but it will be the maximum of the number of parameters and
//	number of results.
//
// \param num_args_and_results the size of the `args_and_results` parameter in
//
//	units of #wasmtime_val_raw_t.
//
// This callback can optionally return a #wasm_trap_t indicating that a trap
// should be raised in WebAssembly. It's expected that in this case the caller
// relinquishes ownership of the trap, and it is passed back to the engine.
//
// This differs from #wasmtime_func_callback_t in that the payload of
// `args_and_results` does not have type information, nor does it have sizing
// information. This is especially unsafe because it's only valid within the
// particular #wasm_functype_t that the function was created with. The onus is
// on the embedder to ensure that `args_and_results` are all read correctly
// for parameters and all written for results within the execution of a
// function.
//
// Parameters will be listed starting at index 0 in the `args_and_results`
// array. Results are also written starting at index 0, which will overwrite
// the arguments.
type UncheckedCallback C.wasmtime_func_unchecked_callback_t

var UncheckedCallbackStub = UncheckedCallback(C.wasmtime_func_unchecked_callback_stub)

// Func is a function instance, which is the runtime representation of a function.
// It effectively is a closure of the original function over the runtime module instance of its originating module.
// The module instance is used to resolve references to other definitions during execution of the function.
// Read more in [spec](https://webassembly.github.io/spec/core/exec/runtime.html#function-instances)
type Func C.wasmtime_func_t

// Caller
//
//	\typedef wasmtime_caller_t
//	\brief Alias to #wasmtime_caller
//
//	\brief Structure used to learn about the caller of a host-defined function.
//	\struct wasmtime_caller
//
//	This structure is an argument to #wasmtime_func_callback_t. The purpose
//	of this structure is acquire a #wasmtime_context_t pointer to interact with
//	objects, but it can also be used for inspect the state of the caller (such as
//	getting memories and functions) with #wasmtime_caller_export_get.
//
//	This object is never owned and does not need to be deleted.
type Caller C.wasmtime_caller_t

// Context returns the store context of the caller object.
func (c *Caller) Context() *Context {
	var ctx *Context
	cgo.NonBlocking((*byte)(C.do_wasmtime_caller_context), uintptr(unsafe.Pointer(c)), uintptr(unsafe.Pointer(&ctx)))
	return ctx
}

// NewFunc Creates a new host-defined function.
//
// Inserts a host-defined function into the `store` provided which can be used
// to then instantiate a module with or define within a #wasmtime_linker_t.
//
// \param store the store in which to create the function
// \param type the wasm type of the function that's being created
// \param callback the host-defined callback to invoke
// \param env host-specific data passed to the callback invocation, can be
// `NULL`
// \param finalizer optional finalizer for `env`, can be `NULL`
// \param ret the #wasmtime_func_t return value to be filled in.
//
// The returned function can only be used with the specified `store`.
func NewFunc(context *Context, ty *FuncType, callback Callback, env uintptr, finalizer FuncFinalizer) (fn Func) {
	args := struct {
		context   uintptr
		funcType  uintptr
		callback  uintptr
		env       uintptr
		finalizer uintptr
		ret       uintptr
	}{
		context:   uintptr(unsafe.Pointer(context)),
		funcType:  uintptr(unsafe.Pointer(ty)),
		callback:  uintptr(unsafe.Pointer(callback)),
		env:       env,
		finalizer: uintptr(unsafe.Pointer(finalizer)),
		ret:       uintptr(unsafe.Pointer(&fn)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_func_new), uintptr(unsafe.Pointer(&args)), 0)
	return
}

// NewFuncUnchecked creates a new host function in the same manner of #wasmtime_func_new,
//
//	but the function-to-call has no type information available at runtime.
//
// This function is very similar to #wasmtime_func_new. The difference is that
// this version is "more unsafe" in that when the host callback is invoked there
// is no type information and no checks that the right types of values are
// produced. The onus is on the consumer of this API to ensure that all
// invariants are upheld such as:
//
// // The host callback reads parameters correctly and interprets their types
//
//	correctly.
//
// // If a trap doesn't happen then all results are written to the results
//
//	pointer. All results must have the correct type.
//
// // Types such as `funcref` cannot cross stores.
// // Types such as `externref` have valid reference counts.
//
// It's generally only recommended to use this if your application can wrap
// this in a safe embedding. This should not be frequently used due to the
// number of invariants that must be upheld on the wasm<->host boundary. On the
// upside, though, this flavor of host function will be faster to call than
// those created by #wasmtime_func_new (hence the reason for this function's
// existence)
func NewFuncUnchecked(context *Context, ty *FuncType, callback UncheckedCallback, env uintptr, finalizer FuncFinalizer) (fn Func) {
	args := struct {
		context   uintptr
		funcType  uintptr
		callback  uintptr
		env       uintptr
		finalizer uintptr
		ret       uintptr
	}{
		context:   uintptr(unsafe.Pointer(context)),
		funcType:  uintptr(unsafe.Pointer(ty)),
		callback:  uintptr(unsafe.Pointer(callback)),
		env:       env,
		finalizer: uintptr(unsafe.Pointer(finalizer)),
		ret:       uintptr(unsafe.Pointer(&fn)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_func_new_unchecked), uintptr(unsafe.Pointer(&args)), 0)
	return
}

// Type returns the type of the function specified
//
// The returned #wasm_functype_t is owned by the caller.
func (f *Func) Type(context *Context) *FuncType {
	args := struct {
		context uintptr
		fn      uintptr
		ret     uintptr
	}{
		context: uintptr(unsafe.Pointer(context)),
		fn:      uintptr(unsafe.Pointer(f)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_func_type), uintptr(unsafe.Pointer(&args)), 0)
	return (*FuncType)(unsafe.Pointer(args.ret))
}

// Call calls a WebAssembly function.
//
// This function is used to invoke a function defined within a store. For
// example this might be used after extracting a function from a
// #wasmtime_instance_t.
//
// \param store the store which owns `func`
// \param func the function to call
// \param args the arguments to the function call
// \param nargs the number of arguments provided
// \param results where to write the results of the function call
// \param nresults the number of results expected
// \param trap where to store a trap, if one happens.
//
// There are three possible return states from this function:
//
//  1. The returned error is non-null. This means `results`
//     wasn't written to and `trap` will have `NULL` written to it. This state
//     means that programmer error happened when calling the function, for
//     example when the size of the arguments/results was wrong, the types of the
//     arguments were wrong, or arguments may come from the wrong store.
//  2. The trap pointer is filled in. This means the returned error is `NULL` and
//     `results` was not written to. This state means that the function was
//     executing but hit a wasm trap while executing.
//  3. The error and trap returned are both `NULL` and `results` are written to.
//     This means that the function call succeeded and the specified results were
//     produced.
//
// The `trap` pointer cannot be `NULL`. The `args` and `results` pointers may be
// `NULL` if the corresponding length is zero.
//
// Does not take ownership of #wasmtime_val_t arguments. Gives ownership of
// #wasmtime_val_t results.
func (f *Func) Call(ctx *Context, params, results []Val) (*Trap, *Error) {
	args := struct {
		context    uintptr
		fn         uintptr
		args       uintptr
		argsLen    uintptr
		results    uintptr
		resultsLen uintptr
		trap       uintptr
		error      uintptr
	}{
		context:    uintptr(unsafe.Pointer(ctx)),
		fn:         uintptr(unsafe.Pointer(f)),
		args:       dataPtr(params),
		argsLen:    uintptr(len(params)),
		results:    dataPtr(results),
		resultsLen: uintptr(len(results)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_func_call), uintptr(unsafe.Pointer(&args)), 0)
	return (*Trap)(unsafe.Pointer(args.trap)), (*Error)(unsafe.Pointer(args.error))
}

// CallBlocking calls a WebAssembly function.
//
// This function is used to invoke a function defined within a store. For
// example this might be used after extracting a function from a
// #wasmtime_instance_t.
//
// \param store the store which owns `func`
// \param func the function to call
// \param args the arguments to the function call
// \param nargs the number of arguments provided
// \param results where to write the results of the function call
// \param nresults the number of results expected
// \param trap where to store a trap, if one happens.
//
// There are three possible return states from this function:
//
//  1. The returned error is non-null. This means `results`
//     wasn't written to and `trap` will have `NULL` written to it. This state
//     means that programmer error happened when calling the function, for
//     example when the size of the arguments/results was wrong, the types of the
//     arguments were wrong, or arguments may come from the wrong store.
//  2. The trap pointer is filled in. This means the returned error is `NULL` and
//     `results` was not written to. This state means that the function was
//     executing but hit a wasm trap while executing.
//  3. The error and trap returned are both `NULL` and `results` are written to.
//     This means that the function call succeeded and the specified results were
//     produced.
//
// The `trap` pointer cannot be `NULL`. The `args` and `results` pointers may be
// `NULL` if the corresponding length is zero.
//
// Does not take ownership of #wasmtime_val_t arguments. Gives ownership of
// #wasmtime_val_t results.
func (f *Func) CallBlocking(ctx *Context, params, results []Val) (*Trap, *Error) {
	args := struct {
		context    uintptr
		fn         uintptr
		args       uintptr
		argsLen    uintptr
		results    uintptr
		resultsLen uintptr
		trap       uintptr
		error      uintptr
	}{
		context:    uintptr(unsafe.Pointer(ctx)),
		fn:         uintptr(unsafe.Pointer(f)),
		args:       dataPtr(params),
		argsLen:    uintptr(len(params)),
		results:    dataPtr(results),
		resultsLen: uintptr(len(results)),
	}
	cgo.Blocking((*byte)(C.do_wasmtime_func_call), uintptr(unsafe.Pointer(&args)), 0)
	return (*Trap)(unsafe.Pointer(args.trap)), (*Error)(unsafe.Pointer(args.error))
}

// CallUnchecked calls a WebAssembly function in an "unchecked" fashion.
//
//	This function is similar to #wasmtime_func_call except that there is no type
//	information provided with the arguments (or sizing information). Consequently,
//	this is less safe to call since it's up to the caller to ensure that `args`
//	has an appropriate size and all the parameters are configured with their
//	appropriate values/types. Additionally, all the results must be interpreted
//	correctly if this function returns successfully.
//
//	Parameters must be specified starting at index 0 in the `args_and_results`
//	array. Results are written starting at index 0, which will overwrite
//	the arguments.
//
//	Callers must ensure that various correctness variants are upheld when this
//	API is called such as:
//
//	//  The `args_and_results` pointer has enough space to hold all the parameters
//	  and all the results (but not at the same time).
//	//  Parameters must all be configured as if they were the correct type.
//	//  Values such as `externref` and `funcref` are valid within the store being
//	  called.
//
//	When in doubt it's much safer to call #wasmtime_func_call. This function is
//	faster than that function, but the tradeoff is that embeddings must uphold
//	more invariants rather than relying on Wasmtime to check them for you.
func (f *Func) CallUnchecked(ctx *Context, argsAndResults []ValRaw) *Trap {
	var trap *Trap = nil
	args := struct {
		context uintptr
		fn      uintptr
		args    uintptr
		trap    uintptr
	}{
		context: uintptr(unsafe.Pointer(ctx)),
		fn:      uintptr(unsafe.Pointer(f)),
		args:    dataPtr(argsAndResults),
		trap:    uintptr(unsafe.Pointer(&trap)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_func_call_unchecked), uintptr(unsafe.Pointer(&args)), 0)
	return trap
}

// CallUncheckedBlocking calls a WebAssembly function in an "unchecked" fashion.
//
//	This function is similar to #wasmtime_func_call except that there is no type
//	information provided with the arguments (or sizing information). Consequently,
//	this is less safe to call since it's up to the caller to ensure that `args`
//	has an appropriate size and all the parameters are configured with their
//	appropriate values/types. Additionally, all the results must be interpreted
//	correctly if this function returns successfully.
//
//	Parameters must be specified starting at index 0 in the `args_and_results`
//	array. Results are written starting at index 0, which will overwrite
//	the arguments.
//
//	Callers must ensure that various correctness variants are upheld when this
//	API is called such as:
//
//	//  The `args_and_results` pointer has enough space to hold all the parameters
//	  and all the results (but not at the same time).
//	//  Parameters must all be configured as if they were the correct type.
//	//  Values such as `externref` and `funcref` are valid within the store being
//	  called.
//
//	When in doubt it's much safer to call #wasmtime_func_call. This function is
//	faster than that function, but the tradeoff is that embeddings must uphold
//	more invariants rather than relying on Wasmtime to check them for you.
func (f *Func) CallUncheckedBlocking(ctx *Context, argsAndResults []Val) *Trap {
	var trap *Trap = nil
	args := struct {
		context uintptr
		fn      uintptr
		args    uintptr
		trap    uintptr
	}{
		context: uintptr(unsafe.Pointer(ctx)),
		fn:      uintptr(unsafe.Pointer(f)),
		args:    dataPtr(argsAndResults),
		trap:    uintptr(unsafe.Pointer(&trap)),
	}
	cgo.Blocking((*byte)(C.do_wasmtime_func_call_unchecked), uintptr(unsafe.Pointer(&args)), 0)
	return trap
}

func (f *Func) AsExtern() Extern {
	var ret Extern
	ret.SetFunc(*f)
	return ret
}
