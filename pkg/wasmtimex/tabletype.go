package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasm_tabletype_new_t {
	size_t val_type;
	uint32_t min;
	uint32_t max;
	size_t result;
} do_wasm_tabletype_new_t;

void do_wasm_tabletype_new(size_t arg0, size_t arg1) {
	do_wasm_tabletype_new_t* args = (do_wasm_tabletype_new_t*)(void*)arg0;
	wasm_limits_t limits;
	limits.min = args->min;
	limits.max = args->max;
	args->result = (size_t)(void*)wasm_tabletype_new(
		(wasm_valtype_t*)(void*)args->val_type,
		&limits
	);
}

void do_wasm_tabletype_delete(size_t arg0, size_t arg1) {
	wasm_tabletype_delete(
		(wasm_tabletype_t*)(void*)arg0
	);
}

void do_wasm_tabletype_element(size_t arg0, size_t arg1) {
	*((const wasm_valtype_t**)arg1) = wasm_tabletype_element(
		(wasm_tabletype_t*)(void*)arg0
	);
}

typedef struct do_wasm_tabletype_limits_t {
	size_t table_type;
	uint32_t min;
	uint32_t max;
} do_wasm_tabletype_limits_t;

void do_wasm_tabletype_limits(size_t arg0, size_t arg1) {
	do_wasm_tabletype_limits_t* args = (do_wasm_tabletype_limits_t*)(void*)arg0;
	const wasm_limits_t* limits = wasm_tabletype_limits(
		(wasm_tabletype_t*)(void*)args->table_type
	);
	args->min = limits->min;
	args->max = limits->max;
}

void do_wasm_tabletype_as_externtype_const(size_t arg0, size_t arg1) {
	*((const wasm_externtype_t**)arg1) = wasm_tabletype_as_externtype_const(
		(const wasm_tabletype_t*)(void*)arg0
	);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

// TableType is one of table types which classify tables over elements of element types within a size range.
type TableType C.wasm_tabletype_t

// NewTableType creates a new `TableType` with the `element` type provided as
// well as limits on its size.
//
// The `min` value is the minimum size, in elements, of this table. The
// `has_max` boolean indicates whether a maximum size is present, and if so
// `max` is used as the maximum size of the table, in elements.
func NewTableType(element *ValType, min uint32, hasMax bool, max uint32) *TableType {
	if !hasMax {
		max = 0xffffffff
	}
	args := struct {
		valType uintptr
		min     uint32
		max     uint32
		result  uintptr
	}{
		valType: uintptr(unsafe.Pointer(element.Clone())),
		min:     min,
		max:     max,
	}
	cgo.NonBlocking((*byte)(C.do_wasm_tabletype_new), uintptr(unsafe.Pointer(&args)), 0)
	return (*TableType)(unsafe.Pointer(args.result))
}

func (tt *TableType) Delete() {
	cgo.NonBlocking((*byte)(C.do_wasm_tabletype_delete), uintptr(unsafe.Pointer(tt)), 0)
}

// Element returns the type of value stored in this table.
func (tt *TableType) Element() *ValType {
	var ty *ValType
	cgo.NonBlocking((*byte)(C.do_wasm_tabletype_element), uintptr(unsafe.Pointer(tt)), uintptr(unsafe.Pointer(&ty)))
	return ty
}

func (tt *TableType) Limits() (min uint32, hasMax bool, max uint32) {
	args := struct {
		tableType uintptr
		min       uint32
		max       uint32
	}{
		tableType: uintptr(unsafe.Pointer(tt)),
	}
	cgo.NonBlocking((*byte)(C.do_wasm_tabletype_limits), uintptr(unsafe.Pointer(&args)), 0)
	return args.min, args.max != 0xffffffff, args.max
}

// Minimum returns the minimum size, in elements, of this table.
func (tt *TableType) Minimum() uint32 {
	min, _, _ := tt.Limits()
	return min
}

// Maximum returns the maximum size, in elements, of this table.
//
// If no maximum size is listed then `(false, 0)` is returned, otherwise
// `(true, N)` is returned where `N` is the maximum size.
func (tt *TableType) Maximum() (hasMax bool, max uint32) {
	_, hasMax, max = tt.Limits()
	return
}

// AsExternType converts this type to an instance of `ExternType`
func (tt *TableType) AsExternType() *ExternType {
	var ty *ExternType
	cgo.NonBlocking((*byte)(C.do_wasm_tabletype_as_externtype_const), uintptr(unsafe.Pointer(tt)), uintptr(unsafe.Pointer(&ty)))
	return ty
}
