package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

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
		(wasmtime_val_raw_t*)(void*)args->args
	);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"reflect"
	"unsafe"
)

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

// (*void *env,
// wasmtime_caller_t* caller,
// const wasmtime_val_t *args,
// size_t nargs,
// wasmtime_val_t *results,
// size_t nresults) wasm_trap_t*
type FuncCallback C.wasmtime_func_callback_t

//func NewFunc()

// Call a WebAssembly function in an "unchecked" fashion.
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
func (f *Func) Call(ctx *Context, argsAndResults []Val) *Trap {
	args := struct {
		context uintptr
		fn      uintptr
		args    uintptr
		trap    uintptr
	}{
		context: uintptr(unsafe.Pointer(ctx)),
		fn:      uintptr(unsafe.Pointer(f)),
		args:    (*reflect.SliceHeader)(unsafe.Pointer(&argsAndResults)).Data,
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_func_call_unchecked), uintptr(unsafe.Pointer(&args)), 0)
	return (*Trap)(unsafe.Pointer(args.trap))
}

func (f *Func) AsExtern() Extern {
	var ret Extern
	ret.SetFunc(f)
	return ret
}
