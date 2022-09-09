package wasmtimex

/*
#include <stdlib.h>
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasmtime_trap_code_t {
	size_t trap;
	size_t code;
	size_t result;
} do_wasmtime_trap_code_t;

void do_wasmtime_trap_code(size_t arg0, size_t arg1) {
	do_wasmtime_trap_code_t* args = (do_wasmtime_trap_code_t*)(void*)arg0;
	args->result = wasmtime_trap_code(
		(wasm_trap_t*)(void*)args->trap,
		(uint8_t*)(void*)args->code) ? 1 : 0;
}

typedef struct do_wasm_trap_message_t {
	size_t trap;
	size_t message;
} do_wasm_trap_message_t;

void do_wasm_trap_message(size_t arg0, size_t arg1) {
	do_wasm_trap_message_t* args = (do_wasm_trap_message_t*)(void*)arg0;
	wasm_trap_message((wasm_trap_t*)(void*)args->trap, (wasm_byte_vec_t*)(void*)args->message);
}

void do_wasm_trap_delete(size_t arg0, size_t arg1) {
	wasm_trap_delete((wasm_trap_t*)(void*)arg0);
}

void do_wasm_frame_vec_delete(size_t arg0, size_t arg1) {
	wasm_frame_vec_delete((wasm_frame_vec_t*)(void*)arg0);
}

void do_wasm_frame_func_index(size_t arg0, size_t arg1) {
	*((uint32_t*)arg1) = wasm_frame_func_index(
		(wasm_frame_t*)(void*)arg0
	);
}

void do_wasm_frame_module_offset(size_t arg0, size_t arg1) {
	*((uint64_t*)arg1) = (uint64_t)wasm_frame_module_offset(
		(wasm_frame_t*)(void*)arg0
	);
}

void do_wasm_frame_func_offset(size_t arg0, size_t arg1) {
	*((uint64_t*)arg1) = (uint64_t)wasm_frame_func_offset(
		(wasm_frame_t*)(void*)arg0
	);
}

void do_wasmtime_frame_func_name(size_t arg0, size_t arg1) {
	*((const wasm_name_t**)arg1) = wasmtime_frame_func_name(
		(wasm_frame_t*)(void*)arg0
	);
}

void do_wasmtime_frame_module_name(size_t arg0, size_t arg1) {
	*((const wasm_name_t**)arg1) = wasmtime_frame_module_name(
		(wasm_frame_t*)(void*)arg0
	);
}

void do_wasm_trap_trace(size_t arg0, size_t arg1) {
	wasm_trap_trace(
		(const wasm_trap_t*)(void*)arg0,
		(wasm_frame_vec_t*)(void*)arg1
	);
}

void do_wasm_trap_origin(size_t arg0, size_t arg1) {
	*((const wasm_frame_t**)arg1) = wasm_trap_origin(
		(const wasm_trap_t*)(void*)arg0
	);
}

typedef struct do_wasmtime_trap_exit_status_t {
	size_t trap;
	int32_t status;
	int32_t ok;
} do_wasmtime_trap_exit_status_t;

void do_wasmtime_trap_exit_status(size_t arg0, size_t arg1) {
	do_wasmtime_trap_exit_status_t* args = (do_wasmtime_trap_exit_status_t*)(void*)arg0;
	args->ok = wasmtime_trap_exit_status(
		(const wasm_trap_t*)(void*)args->trap,
		(int*)(void*)&args->status
	) ? 1 : 0;
}

*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"reflect"
	"unsafe"
)

// Trap is the trap instruction which represents the occurrence of a trap.
// Traps are bubbled up through nested instruction sequences, ultimately reducing the entire program to a single trap instruction, signalling abrupt termination.
type Trap C.wasm_trap_t

// Frame is one of activation frames which carry the return arity n of the respective function,
// hold the values of its locals (including arguments) in the order corresponding to their static local indices,
// and a reference to the function’s own module instance
type Frame C.wasm_frame_t

// TrapCode is the code of an instruction trap.
type TrapCode uint8

const (
	// StackOverflow the current stack space was exhausted.
	StackOverflow = TrapCode(C.WASMTIME_TRAP_CODE_STACK_OVERFLOW)
	// MemoryOutOfBounds out-of-bounds memory access.
	MemoryOutOfBounds = TrapCode(C.WASMTIME_TRAP_CODE_MEMORY_OUT_OF_BOUNDS)
	// HeapMisaligned a wasm atomic operation was presented with a not-naturally-aligned linear-memory address.
	HeapMisaligned = TrapCode(C.WASMTIME_TRAP_CODE_HEAP_MISALIGNED)
	// TableOutOfBounds out-of-bounds access to a table.
	TableOutOfBounds = TrapCode(C.WASMTIME_TRAP_CODE_TABLE_OUT_OF_BOUNDS)
	// IndirectCallToNull indirect call to a null table entry.
	IndirectCallToNull = TrapCode(C.WASMTIME_TRAP_CODE_INDIRECT_CALL_TO_NULL)
	// BadSignature signature mismatch on indirect call.
	BadSignature = TrapCode(C.WASMTIME_TRAP_CODE_BAD_SIGNATURE)
	// IntegerOverflow an integer arithmetic operation caused an overflow.
	IntegerOverflow = TrapCode(C.WASMTIME_TRAP_CODE_INTEGER_OVERFLOW)
	// IntegerDivisionByZero integer division by zero.
	IntegerDivisionByZero = TrapCode(C.WASMTIME_TRAP_CODE_INTEGER_DIVISION_BY_ZERO)
	// BadConversionToInteger failed float-to-int conversion.
	BadConversionToInteger = TrapCode(C.WASMTIME_TRAP_CODE_BAD_CONVERSION_TO_INTEGER)
	// UnreachableCodeReached code that was supposed to have been unreachable was reached.
	UnreachableCodeReached = TrapCode(C.WASMTIME_TRAP_CODE_UNREACHABLE_CODE_REACHED)
	// Interrupt execution has been interrupted.
	Interrupt = TrapCode(C.WASMTIME_TRAP_CODE_INTERRUPT)
)

// NewTrap creates a new `Trap` with the `name` and the type provided.
func NewTrap(message string) *Trap {
	return (*Trap)(unsafe.Pointer(C.wasmtime_trap_new(C._GoStringPtr(message), C._GoStringLen(message))))
}

func (t *Trap) IsError() bool {
	return t != nil
}

func (t *Trap) Delete() {
	cgo.NonBlocking((*byte)(C.do_wasm_trap_delete), uintptr(unsafe.Pointer(t)), 0)
}

func (t *Trap) Message() (message ByteVec) {
	args := struct {
		trap    uintptr
		message uintptr
	}{
		trap:    uintptr(unsafe.Pointer(t)),
		message: uintptr(unsafe.Pointer(&message.vec)),
	}
	cgo.NonBlocking((*byte)(C.do_wasm_trap_message), uintptr(unsafe.Pointer(&args)), 0)
	return
}

// Code attempts to extract the trap code from this trap.
//
// Returns `true` if the trap is an instruction trap triggered while
// executing Wasm. If `true` is returned then the trap code is returned
// through the `code` pointer. If `false` is returned then this is not
// an instruction trap -- traps can also be created using wasm_trap_new,
// or occur with WASI modules exiting with a certain exit code.
func (t *Trap) Code() (code TrapCode, ok bool) {
	args := struct {
		trap   uintptr
		code   uintptr
		result uintptr
	}{
		trap: uintptr(unsafe.Pointer(t)),
		code: uintptr(unsafe.Pointer(&code)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_trap_code), uintptr(unsafe.Pointer(&args)), 0)
	return code, args.result != 0
}

func (t *Trap) Error() string {
	m := t.Message()
	return m.ToOwned()
}

// Frames returns the wasm function frames that make up this trap
func (t *Trap) Frames() (frames FrameVec) {
	cgo.NonBlocking((*byte)(C.do_wasm_trap_trace), uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(&frames)))
	return
}

func (t *Trap) Origin() (origin *Frame) {
	cgo.NonBlocking((*byte)(C.do_wasm_trap_origin), uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(&origin)))
	return
}

// ExitStatus attempts to extract a WASI-specific exit status from this trap.
//
// Returns `true` if the trap is a WASI "exit" trap and has a return status. If
// `true` is returned then the exit status is returned through the `status`
// pointer. If `false` is returned then this is not a wasi exit trap.
func (t *Trap) ExitStatus() (status int32, ok bool) {
	args := struct {
		trap   uintptr
		status int32
		ok     int32
	}{
		trap: uintptr(unsafe.Pointer(t)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_trap_exit_status), uintptr(unsafe.Pointer(&args)), 0)
	return args.status, args.ok != 0
}

type FrameVec struct {
	vec C.wasm_frame_vec_t
}

func (fv *FrameVec) Delete() {
	cgo.NonBlocking((*byte)(C.do_wasm_frame_vec_delete), uintptr(unsafe.Pointer(&fv.vec)), 0)
}

func (fv *FrameVec) Size() int {
	return int(fv.vec.size)
}

func (fv *FrameVec) At(index int) *Frame {
	size := fv.Size()
	if index < 0 || index >= size {
		return nil
	}
	return *(**Frame)(unsafe.Add(unsafe.Pointer(fv.vec.data), uintptr(index)*unsafe.Sizeof(uintptr(0))))
}

func (fv *FrameVec) Unsafe() []*Frame {
	return *(*[]*Frame)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(fv.vec.data)),
		Len:  int(fv.vec.size),
		Cap:  int(fv.vec.size),
	}))
}

func (f *Frame) Delete() {
	cgo.NonBlocking((*byte)(C.wasm_frame_delete), uintptr(unsafe.Pointer(f)), 0)
}

// FuncIndex returns the function index in the wasm module that this frame represents
func (f *Frame) FuncIndex() (index uint32) {
	cgo.NonBlocking((*byte)(C.do_wasm_frame_func_index), uintptr(unsafe.Pointer(f)), uintptr(unsafe.Pointer(&index)))
	return
}

// FuncName returns the name, if available, for this frame's function
func (f *Frame) FuncName() BorrowedString {
	var ret *C.wasm_name_t
	cgo.NonBlocking((*byte)(C.do_wasmtime_frame_func_name), uintptr(unsafe.Pointer(f)), uintptr(unsafe.Pointer(&ret)))
	return borrowWasmName(ret)
}

// ModuleName returns the name, if available, for this frame's module
func (f *Frame) ModuleName() BorrowedString {
	var ret *C.wasm_name_t
	cgo.NonBlocking((*byte)(C.do_wasmtime_frame_module_name), uintptr(unsafe.Pointer(f)), uintptr(unsafe.Pointer(&ret)))
	return borrowWasmName(ret)
}

// ModuleOffset returns offset of this frame's instruction into the original module
func (f *Frame) ModuleOffset() (offset uint64) {
	cgo.NonBlocking((*byte)(C.do_wasm_frame_module_offset), uintptr(unsafe.Pointer(f)), uintptr(unsafe.Pointer(&offset)))
	return
}

// FuncOffset returns offset of this frame's instruction into the original function
func (f *Frame) FuncOffset() (offset uint64) {
	cgo.NonBlocking((*byte)(C.do_wasm_frame_func_offset), uintptr(unsafe.Pointer(f)), uintptr(unsafe.Pointer(&offset)))
	return
}
