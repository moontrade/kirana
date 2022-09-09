package wasmtimex

/*
#include <wasmtime.h>
#define WASMTIME_EXTERN_OF_OFFSET offsetof(struct wasmtime_extern, of)

void do_wasmtime_extern_delete(size_t arg0, size_t arg1) {
	wasmtime_extern_delete(
		(wasmtime_extern_t*)(void*)arg0
	);
}

typedef struct do_wasmtime_extern_type_t {
	size_t context;
	size_t val;
	size_t result;
} do_wasmtime_extern_type_t;

void do_wasmtime_extern_type(size_t arg0, size_t arg1) {
	do_wasmtime_extern_type_t* args = (do_wasmtime_extern_type_t*)(void*)arg0;
	args->result = (size_t)(void*)wasmtime_extern_type(
		(wasmtime_context_t*)(void*)args->context,
		(wasmtime_extern_t*)(void*)args->val
	);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

type ExternKind uint8

const (
	ExternKindFunc   = ExternKind(C.WASMTIME_EXTERN_FUNC)
	ExternKindGlobal = ExternKind(C.WASMTIME_EXTERN_GLOBAL)
	ExternKindTable  = ExternKind(C.WASMTIME_EXTERN_TABLE)
	ExternKindMemory = ExternKind(C.WASMTIME_EXTERN_MEMORY)
)

type ExternUnion C.wasmtime_extern_union_t

func (extern *ExternUnion) Func() *Func {
	return (*Func)(unsafe.Pointer(extern))
}

func (extern *ExternUnion) Global() *Global {
	return (*Global)(unsafe.Pointer(extern))
}

func (extern *ExternUnion) Table() *Table {
	return (*Table)(unsafe.Pointer(extern))
}

func (extern *ExternUnion) Memory() *Memory {
	return (*Memory)(unsafe.Pointer(extern))
}

func (eu *ExternUnion) SetFunc(fn Func) {
	*(*Func)(unsafe.Pointer(eu)) = fn
}

func (eu *ExternUnion) SetGlobal(global Global) {
	*(*Global)(unsafe.Pointer(eu)) = global
}

func (eu *ExternUnion) SetMemory(memory Memory) {
	*(*Memory)(unsafe.Pointer(eu)) = memory
}

func (eu *ExternUnion) SetTable(table Table) {
	*(*Table)(unsafe.Pointer(eu)) = table
}

type Extern struct {
	Kind ExternKind
	_    [C.WASMTIME_EXTERN_OF_OFFSET - 1]byte
	Of   ExternUnion
}

func (extern *Extern) Delete() {
	cgo.NonBlocking((*byte)(C.do_wasmtime_extern_delete), uintptr(unsafe.Pointer(extern)), 0)
}

// Type returns the type of the #wasmtime_extern_t defined within the given store.
//
// Does not take ownership of `context` or `val`, but the returned
// #wasm_externtype_t is an owned value that needs to be deleted.
func (extern *Extern) Type(ctx *Context) *ExternType {
	args := struct {
		context uintptr
		val     uintptr
		result  uintptr
	}{
		context: uintptr(unsafe.Pointer(ctx)),
		val:     uintptr(unsafe.Pointer(extern)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_extern_type), uintptr(unsafe.Pointer(&args)), 0)
	return (*ExternType)(unsafe.Pointer(args.result))
}

func (extern *Extern) Func() *Func {
	return (*Func)(unsafe.Pointer(&extern.Of))
}

func (extern *Extern) Global() *Global {
	return (*Global)(unsafe.Pointer(&extern.Of))
}

func (extern *Extern) Table() *Table {
	return (*Table)(unsafe.Pointer(&extern.Of))
}

func (extern *Extern) Memory() *Memory {
	return (*Memory)(unsafe.Pointer(&extern.Of))
}

func (extern *Extern) SetFunc(fn Func) {
	extern.Kind = ExternKindFunc
	extern.Of.SetFunc(fn)
}

func (extern *Extern) SetGlobal(global Global) {
	extern.Kind = ExternKindGlobal
	extern.Of.SetGlobal(global)
}

func (extern *Extern) SetMemory(memory Memory) {
	extern.Kind = ExternKindMemory
	extern.Of.SetMemory(memory)
}

func (extern *Extern) SetTable(table Table) {
	extern.Kind = ExternKindTable
	extern.Of.SetTable(table)
}
