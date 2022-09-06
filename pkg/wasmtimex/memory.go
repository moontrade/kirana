package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasmtime_memory_new_t {
	size_t context;
	size_t memory_type;
	size_t memory;
	size_t error;
} do_wasmtime_memory_new_t;

void do_wasmtime_memory_new(size_t arg0, size_t arg1) {
	do_wasmtime_memory_new_t* args = (do_wasmtime_memory_new_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_memory_new(
		(wasmtime_context_t*)(void*)args->context,
		(const wasm_memorytype_t*)(void*)args->memory_type,
		(wasmtime_memory_t*)(void*)args->memory
	);
}

typedef struct do_wasmtime_memory_type_t {
	size_t context;
	size_t memory;
	size_t memory_type;
} do_wasmtime_memory_type_t;

void do_wasmtime_memory_type(size_t arg0, size_t arg1) {
	do_wasmtime_memory_type_t* args = (do_wasmtime_memory_type_t*)(void*)arg0;
	args->memory_type = (size_t)(void*)wasmtime_memory_type(
		(wasmtime_context_t*)(void*)args->context,
		(wasmtime_memory_t*)(void*)args->memory
	);
}

typedef struct do_wasmtime_memory_data_t {
	size_t context;
	size_t memory;
	size_t data;
	size_t data_size;
} do_wasmtime_memory_data_t;

void do_wasmtime_memory_data(size_t arg0, size_t arg1) {
	do_wasmtime_memory_data_t* args = (do_wasmtime_memory_data_t*)(void*)arg0;
	args->data = (size_t)(void*)wasmtime_memory_data(
		(wasmtime_context_t*)(void*)args->context,
		(wasmtime_memory_t*)(void*)args->memory
	);
	args->data_size = (size_t)(void*)wasmtime_memory_data_size(
		(wasmtime_context_t*)(void*)args->context,
		(wasmtime_memory_t*)(void*)args->memory
	);
}

typedef struct do_wasmtime_memory_size_t {
	size_t context;
	size_t memory;
	uint64_t size;
} do_wasmtime_memory_size_t;

void do_wasmtime_memory_size(size_t arg0, size_t arg1) {
	do_wasmtime_memory_size_t* args = (do_wasmtime_memory_size_t*)(void*)arg0;
	args->size = wasmtime_memory_size(
		(wasmtime_context_t*)(void*)args->context,
		(wasmtime_memory_t*)(void*)args->memory
	);
}

typedef struct do_wasmtime_memory_grow_t {
	size_t context;
	size_t memory;
	uint64_t delta;
	uint64_t prev_size;
	size_t error;
} do_wasmtime_memory_grow_t;

void do_wasmtime_memory_grow(size_t arg0, size_t arg1) {
	do_wasmtime_memory_grow_t* args = (do_wasmtime_memory_grow_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_memory_grow(
		(wasmtime_context_t*)(void*)args->context,
		(wasmtime_memory_t*)(void*)args->memory,
		args->delta,
		&args->prev_size
	);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"reflect"
	"unsafe"
)

type Memory C.wasmtime_memory_t

// NewMemory creates a new WebAssembly linear memory
//
//	\param store the store to create the memory within
//	\param ty the type of the memory to create
//	\param ret where to store the returned memory
//
//	If an error happens when creating the memory it's returned and owned by the
//	caller. If an error happens then `ret` is not filled in.
func NewMemory(context *Context, memoryType *MemoryType) (Memory, *Error) {
	var memory Memory
	args := struct {
		context    uintptr
		memoryType uintptr
		memory     uintptr
		error      uintptr
	}{
		context:    uintptr(unsafe.Pointer(context)),
		memoryType: uintptr(unsafe.Pointer(memoryType)),
		memory:     uintptr(unsafe.Pointer(&memory)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_memory_new), uintptr(unsafe.Pointer(&args)), 0)
	return memory, (*Error)(unsafe.Pointer(args.error))
}

// Type returns the MemoryType of the Memory
func (m *Memory) Type(context *Context) *MemoryType {
	args := struct {
		context    uintptr
		memory     uintptr
		memoryType uintptr
	}{
		context: uintptr(unsafe.Pointer(context)),
		memory:  uintptr(unsafe.Pointer(m)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_memory_type), uintptr(unsafe.Pointer(&args)), 0)
	return (*MemoryType)(unsafe.Pointer(args.memoryType))
}

// Data returns a byte slice starting where the linear memory starts
func (m *Memory) Data(context *Context) []byte {
	args := struct {
		context  uintptr
		memory   uintptr
		data     uintptr
		dataSize uintptr
	}{
		context: uintptr(unsafe.Pointer(context)),
		memory:  uintptr(unsafe.Pointer(m)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_memory_data), uintptr(unsafe.Pointer(&args)), 0)
	if args.data == 0 {
		return nil
	}
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: args.data,
		Len:  int(args.dataSize),
		Cap:  int(args.dataSize),
	}))
}

// Size returns the length, in WebAssembly pages, of this linear memory.
func (m *Memory) Size(context *Context) (pages uint64) {
	args := struct {
		context uintptr
		memory  uintptr
		size    uint64
	}{
		context: uintptr(unsafe.Pointer(context)),
		memory:  uintptr(unsafe.Pointer(m)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_memory_size), uintptr(unsafe.Pointer(&args)), 0)
	return args.size
}

// Grow attempts to grow the specified memory by `delta` pages.
//
//	\param store the store that owns `memory`
//	\param memory the memory to grow
//	\param delta the number of pages to grow by
//	\param prev_size where to store the previous size of memory
//
//	If memory cannot be grown then `prev_size` is left unchanged and an error is
//	returned. Otherwise, `prev_size` is set to the previous size of the memory, in
//	WebAssembly pages, and `NULL` is returned.
func (m *Memory) Grow(context *Context, delta uint64) (prevSize uint64, err *Error) {
	args := struct {
		context  uintptr
		memory   uintptr
		delta    uint64
		prevSize uint64
		error    uintptr
	}{
		context: uintptr(unsafe.Pointer(context)),
		memory:  uintptr(unsafe.Pointer(m)),
		delta:   delta,
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_memory_grow), uintptr(unsafe.Pointer(&args)), 0)
	return args.prevSize, (*Error)(unsafe.Pointer(args.error))
}

func (m *Memory) AsExtern() Extern {
	var ret Extern
	ret.SetMemory(m)
	return ret
}
