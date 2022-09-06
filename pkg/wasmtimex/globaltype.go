package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasm_globaltype_new_t {
	size_t content;
	size_t mutability;
	size_t result;
} do_wasm_globaltype_new_t;

void do_wasm_globaltype_new(size_t arg0, size_t arg1) {
	do_wasm_globaltype_new_t* args = (do_wasm_globaltype_new_t*)(void*)arg0;
	args->result = (size_t)(void*)wasm_globaltype_new(
		(wasm_valtype_t*)(void*)args->content,
		(wasm_mutability_t)args->mutability
	);
}

void do_wasm_globaltype_delete(size_t arg0, size_t arg1) {
	wasm_globaltype_delete(
		(wasm_globaltype_t*)(void*)arg0
	);
}

void do_wasm_globaltype_content(size_t arg0, size_t arg1) {
	*((const wasm_valtype_t**)arg1) = wasm_globaltype_content(
		(const wasm_globaltype_t*)(void*)arg0
	);
}

void do_wasm_globaltype_mutability(size_t arg0, size_t arg1) {
	*((wasm_mutability_t*)arg1) = wasm_globaltype_mutability(
		(const wasm_globaltype_t*)(void*)arg0
	);
}

void do_wasm_globaltype_as_externtype_const(size_t arg0, size_t arg1) {
	*((const wasm_externtype_t**)arg1) = wasm_globaltype_as_externtype_const(
		(const wasm_globaltype_t*)(void*)arg0
	);
}

*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

type Mutability uint8

const (
	MutabilityConst = Mutability(C.WASM_CONST)
	MutabilityVar   = Mutability(C.WASM_VAR)
)

// GlobalType is a ValType, which classify global variables and hold a value and can either be mutable or immutable.
type GlobalType C.wasm_globaltype_t

// NewGlobalType creates a new `GlobalType` with the `kind` provided and whether it's
// `mutable` or not
func NewGlobalType(content *ValType, mutability Mutability) *GlobalType {
	args := struct {
		content    uintptr
		mutability uintptr
		result     uintptr
	}{
		content:    uintptr(unsafe.Pointer(content.Clone())),
		mutability: uintptr(mutability),
	}
	cgo.NonBlocking((*byte)(C.do_wasm_globaltype_new), uintptr(unsafe.Pointer(&args)), 0)
	return (*GlobalType)(unsafe.Pointer(args.result))
}

func (gt *GlobalType) Delete() {
	cgo.NonBlocking((*byte)(C.do_wasm_globaltype_delete), uintptr(unsafe.Pointer(gt)), 0)
}

// Content returns the type of value stored in this global
func (gt *GlobalType) Content() *ValType {
	var ty *ValType
	cgo.NonBlocking((*byte)(C.do_wasm_globaltype_content), uintptr(unsafe.Pointer(gt)), uintptr(unsafe.Pointer(&ty)))
	return ty
}

// Mutability returns whether this global type is mutable or not
func (gt *GlobalType) Mutability() Mutability {
	var result Mutability
	cgo.NonBlocking((*byte)(C.do_wasm_globaltype_mutability), uintptr(unsafe.Pointer(gt)), uintptr(unsafe.Pointer(&result)))
	return result
}

// AsExternType converts this type to an instance of `ExternType`
func (gt *GlobalType) AsExternType() *ExternType {
	var ty *ExternType
	cgo.NonBlocking((*byte)(C.do_wasm_globaltype_as_externtype_const), uintptr(unsafe.Pointer(gt)), uintptr(unsafe.Pointer(&ty)))
	return ty
}
