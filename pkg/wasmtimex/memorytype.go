package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasmtime_memorytype_new_t {
	uint64_t min;
	uint64_t max;
	uint32_t max_present;
	uint32_t is_64;
	size_t result;
} do_wasmtime_memorytype_new_t;

void do_wasmtime_memorytype_new(size_t arg0, size_t arg1) {
	do_wasmtime_memorytype_new_t* args = (do_wasmtime_memorytype_new_t*)(void*)arg0;
	args->result = (size_t)(void*)wasmtime_memorytype_new(
		args->min,
		args->max_present != 0,
		args->max,
		args->is_64 != 0
	);
}

void do_wasm_memorytype_delete(size_t arg0, size_t arg1) {
	wasm_memorytype_delete(
		(wasm_memorytype_t*)(void*)arg0
	);
}

void do_wasmtime_memorytype_minimum(size_t arg0, size_t arg1) {
	*((uint64_t*)arg1) = wasmtime_memorytype_minimum(
		(wasm_memorytype_t*)(void*)arg0
	);
}

void do_wasmtime_memorytype_maximum(size_t arg0, size_t arg1) {
	if (!wasmtime_memorytype_maximum(
		(wasm_memorytype_t*)(void*)arg0,
		(uint64_t*)(void*)arg1
	)) {
		*(uint64_t*)(void*)arg1 = 0;
	}
}

void do_wasmtime_memorytype_is64(size_t arg0, size_t arg1) {
	*((size_t*)arg1) = wasmtime_memorytype_is64(
		(wasm_memorytype_t*)(void*)arg0
	) ? 1 : 0;
}

void do_wasm_memorytype_as_externtype_const(size_t arg0, size_t arg1) {
	*((const wasm_externtype_t**)arg1) = wasm_memorytype_as_externtype_const(
		(const wasm_memorytype_t*)(void*)arg0
	);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

type MemoryType C.wasm_memorytype_t

func NewMemoryType(min uint64, maxPresent bool, max uint64, is64 bool) *MemoryType {
	args := struct {
		min        uint64
		max        uint64
		maxPresent uint32
		is64       uint32
		result     uintptr
	}{
		min:        min,
		max:        max,
		maxPresent: boolToU32(maxPresent),
		is64:       boolToU32(is64),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_memorytype_new), uintptr(unsafe.Pointer(&args)), 0)
	return (*MemoryType)(unsafe.Pointer(args.result))
}

func (mt *MemoryType) Delete() {
	cgo.NonBlocking((*byte)(C.do_wasm_memorytype_delete), uintptr(unsafe.Pointer(mt)), 0)
}

// Minimum returns the minimum size, in pages, of the specified memory type.
//
// Note that this function is preferred over #wasm_memorytype_limits for
// compatibility with the memory64 proposal.
func (mt *MemoryType) Minimum() uint64 {
	var min uint64
	cgo.NonBlocking((*byte)(C.do_wasmtime_memorytype_minimum), uintptr(unsafe.Pointer(mt)), uintptr(unsafe.Pointer(&min)))
	return min
}

// Maximum returns the maximum size, in pages, of the specified memory type.
//
// If this memory type doesn't have a maximum size listed then `0` is
// returned. Otherwise, returns the maximum size in pages.
//
// Note that this function is preferred over #wasm_memorytype_limits for
// compatibility with the memory64 proposal.
func (mt *MemoryType) Maximum() uint64 {
	var max uint64
	cgo.NonBlocking((*byte)(C.do_wasmtime_memorytype_maximum), uintptr(unsafe.Pointer(mt)), uintptr(unsafe.Pointer(&max)))
	return max
}

// Is64 returns whether this type of memory represents a 64-bit memory.
func (mt *MemoryType) Is64() bool {
	var is64 uintptr
	cgo.NonBlocking((*byte)(C.do_wasmtime_memorytype_maximum), uintptr(unsafe.Pointer(mt)), uintptr(unsafe.Pointer(&is64)))
	return is64 != 0
}

// AsExternType converts this type to an instance of `ExternType`
func (mt *MemoryType) AsExternType() *ExternType {
	var ty *ExternType
	cgo.NonBlocking((*byte)(C.do_wasm_memorytype_as_externtype_const), uintptr(unsafe.Pointer(mt)), uintptr(unsafe.Pointer(&ty)))
	return ty
}
