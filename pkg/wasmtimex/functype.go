package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasm_functype_new_t {
	size_t params;
	size_t results;
	size_t result;
} do_wasm_functype_new_t;

void do_wasm_functype_new(size_t arg0, size_t arg1) {
	do_wasm_functype_new_t* args = (do_wasm_functype_new_t*)(void*)arg0;
	args->result = (size_t)(void*)wasm_functype_new(
		(wasm_valtype_vec_t*)(void*)args->params,
		(wasm_valtype_vec_t*)(void*)args->results
	);
}

void do_wasm_functype_delete(size_t arg0, size_t arg1) {
	wasm_functype_delete(
		(wasm_functype_t*)(void*)arg0
	);
}

void do_wasm_functype_params(size_t arg0, size_t arg1) {
	*((const wasm_valtype_vec_t**)arg1) = wasm_functype_params(
		(const wasm_functype_t*)(void*)arg0
	);
}

void do_wasm_functype_results(size_t arg0, size_t arg1) {
	*((const wasm_valtype_vec_t**)arg1) = wasm_functype_results(
		(const wasm_functype_t*)(void*)arg0
	);
}

void do_wasm_functype_as_externtype_const(size_t arg0, size_t arg1) {
	*((const wasm_externtype_t**)arg1) = wasm_functype_as_externtype_const(
		(const wasm_functype_t*)(void*)arg0
	);
}

*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

// FuncType is one of function types which classify the signature of functions, mapping a vector of
// parameters to a vector of results. They are also used to classify the inputs and outputs of instructions.
type FuncType C.wasm_functype_t

//goland:noinspection ALL
func NewFuncType(params, results ValTypeVec) *FuncType {
	args := struct {
		params  uintptr
		results uintptr
		result  uintptr
	}{
		params:  uintptr(unsafe.Pointer(&params)),
		results: uintptr(unsafe.Pointer(&results)),
	}
	cgo.NonBlocking((*byte)(C.do_wasm_functype_new), uintptr(unsafe.Pointer(&args)), 0)
	return (*FuncType)(unsafe.Pointer(args.result))
}

func (ft *FuncType) Delete() {
	cgo.NonBlocking((*byte)(C.do_wasm_functype_delete), uintptr(unsafe.Pointer(ft)), 0)
}

// Params returns the parameter types of this function type
func (ft *FuncType) Params() *ValTypeVec {
	var vec *ValTypeVec
	cgo.NonBlocking((*byte)(C.do_wasm_functype_params), uintptr(unsafe.Pointer(ft)), uintptr(unsafe.Pointer(&vec)))
	return vec
}

// Results returns the result types of this function type
func (ft *FuncType) Results() *ValTypeVec {
	var vec *ValTypeVec
	cgo.NonBlocking((*byte)(C.do_wasm_functype_results), uintptr(unsafe.Pointer(ft)), uintptr(unsafe.Pointer(&vec)))
	return vec
}

// AsExternType converts this type to an instance of `ExternType`
func (ft *FuncType) AsExternType() *ExternType {
	var ty *ExternType
	cgo.NonBlocking((*byte)(C.do_wasm_functype_as_externtype_const), uintptr(unsafe.Pointer(ft)), uintptr(unsafe.Pointer(&ty)))
	return ty
}

func NewFuncTypeZeroZero() *FuncType {
	return NewFuncType(ValTypeVec{}, ValTypeVec{})
}

func NewFuncTypeZeroOne(result ValKind) *FuncType {
	results := NewValTypeVec(1)
	results.Set(0, NewValType(result))
	return NewFuncType(ValTypeVec{}, results)
}

func NewFuncTypeOneZero(param ValKind) *FuncType {
	params := NewValTypeVec(1)
	params.Set(0, NewValType(param))
	return NewFuncType(params, ValTypeVec{})
}

func NewFuncTypeTwoZero(param, param2 ValKind) *FuncType {
	params := NewValTypeVec(2)
	params.Set(0, NewValType(param))
	params.Set(1, NewValType(param2))
	return NewFuncType(params, ValTypeVec{})
}

func NewFuncTypeThreeZero(param, param2, param3 ValKind) *FuncType {
	params := NewValTypeVec(3)
	params.Set(0, NewValType(param))
	params.Set(1, NewValType(param2))
	params.Set(2, NewValType(param3))
	return NewFuncType(params, ValTypeVec{})
}

func NewFuncTypeFourZero(param, param2, param3, param4 ValKind) *FuncType {
	params := NewValTypeVec(4)
	params.Set(0, NewValType(param))
	params.Set(1, NewValType(param2))
	params.Set(2, NewValType(param3))
	params.Set(3, NewValType(param4))
	return NewFuncType(params, ValTypeVec{})
}

func NewFuncTypeOneOne(param, result ValKind) *FuncType {
	params := NewValTypeVec(1)
	params.Set(0, NewValType(param))
	results := NewValTypeVec(1)
	results.Set(0, NewValType(result))
	return NewFuncType(params, results)
}

func NewFuncTypeTwoOne(param, param2, result ValKind) *FuncType {
	params := NewValTypeVec(2)
	params.Set(0, NewValType(param))
	params.Set(1, NewValType(param2))
	results := NewValTypeVec(1)
	results.Set(0, NewValType(result))
	return NewFuncType(params, results)
}

func NewFuncTypeThreeOne(param, param2, param3, result ValKind) *FuncType {
	params := NewValTypeVec(3)
	params.Set(0, NewValType(param))
	params.Set(1, NewValType(param2))
	params.Set(2, NewValType(param3))
	results := NewValTypeVec(1)
	results.Set(0, NewValType(result))
	return NewFuncType(params, results)
}

func NewFuncTypeFourOne(param, param2, param3, param4, result ValKind) *FuncType {
	params := NewValTypeVec(4)
	params.Set(0, NewValType(param))
	params.Set(1, NewValType(param2))
	params.Set(2, NewValType(param3))
	params.Set(3, NewValType(param4))
	results := NewValTypeVec(1)
	results.Set(0, NewValType(result))
	return NewFuncType(params, results)
}
