package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

void do_wasm_externtype_as_functype(size_t arg0, size_t arg1) {
	*((const wasm_functype_t**)arg1) = wasm_externtype_as_functype(
		(wasm_externtype_t*)(void*)arg0
	);
}

void do_wasm_externtype_as_globaltype(size_t arg0, size_t arg1) {
	*((const wasm_globaltype_t**)arg1) = wasm_externtype_as_globaltype(
		(wasm_externtype_t*)(void*)arg0
	);
}

void do_wasm_externtype_as_tabletype(size_t arg0, size_t arg1) {
	*((const wasm_tabletype_t**)arg1) = wasm_externtype_as_tabletype(
		(wasm_externtype_t*)(void*)arg0
	);
}

void do_wasm_externtype_as_memorytype(size_t arg0, size_t arg1) {
	*((const wasm_memorytype_t**)arg1) = wasm_externtype_as_memorytype(
		(wasm_externtype_t*)(void*)arg0
	);
}

void do_wasm_externtype_copy(size_t arg0, size_t arg1) {
	*((const wasm_externtype_t**)arg1) = wasm_externtype_copy(
		(wasm_externtype_t*)(void*)arg0
	);
}

*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

// AsExternType is an interface for all types which can be ExternType.
type AsExternType interface {
	AsExternType() *ExternType
}

type ExternType C.wasm_externtype_t

// FuncType returns the underlying `FuncType` for this `ExternType` if it's a function  type.
// Otherwise, returns `nil`.
func (et *ExternType) FuncType() *FuncType {
	var ty *FuncType
	cgo.NonBlocking((*byte)(C.do_wasm_externtype_as_functype), uintptr(unsafe.Pointer(et)), uintptr(unsafe.Pointer(&ty)))
	return ty
}

// GlobalType returns the underlying `GlobalType` for this `ExternType` if it's a *global* type.
// Otherwise, returns `nil`.
func (et *ExternType) GlobalType() *GlobalType {
	var ty *GlobalType
	cgo.NonBlocking((*byte)(C.do_wasm_externtype_as_globaltype), uintptr(unsafe.Pointer(et)), uintptr(unsafe.Pointer(&ty)))
	return ty
}

// TableType returns the underlying `TableType` for this `ExternType` if it's a *table* type.
// Otherwise, returns `nil`.
func (et *ExternType) TableType() *TableType {
	var ty *TableType
	cgo.NonBlocking((*byte)(C.do_wasm_externtype_as_tabletype), uintptr(unsafe.Pointer(et)), uintptr(unsafe.Pointer(&ty)))
	return ty
}

// MemoryType returns the underlying `MemoryType` for this `ExternType` if it's a *memory* type.
// Otherwise returns `nil`.
func (et *ExternType) MemoryType() *MemoryType {
	var ty *MemoryType
	cgo.NonBlocking((*byte)(C.do_wasm_externtype_as_memorytype), uintptr(unsafe.Pointer(et)), uintptr(unsafe.Pointer(&ty)))
	return ty
}

func (et *ExternType) Copy() *ExternType {
	var ret *ExternType
	cgo.NonBlocking((*byte)(C.do_wasm_externtype_copy), uintptr(unsafe.Pointer(et)), uintptr(unsafe.Pointer(&ret)))
	return ret
}
