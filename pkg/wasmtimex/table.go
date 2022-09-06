package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasmtime_table_new_t {
	size_t context;
	size_t table_type;
	size_t init;
	size_t table;
	size_t error;
} do_wasmtime_table_new_t;

void do_wasmtime_table_new(size_t arg0, size_t arg1) {
	do_wasmtime_table_new_t* args = (do_wasmtime_table_new_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_table_new(
		(wasmtime_context_t*)(void*)args->context,
		(const wasm_tabletype_t*)(void*)args->table_type,
		(const wasmtime_val_t*)(void*)args->init,
		(wasmtime_table_t*)(void*)args->table
	);
}

typedef struct do_wasmtime_table_type_t {
	size_t context;
	size_t table;
	size_t table_type;
} do_wasmtime_table_type_t;

void do_wasmtime_table_type(size_t arg0, size_t arg1) {
	do_wasmtime_table_type_t* args = (do_wasmtime_table_type_t*)(void*)arg0;
	args->table_type = (size_t)(void*)wasmtime_table_type(
		(wasmtime_context_t*)(void*)args->context,
		(wasmtime_table_t*)(void*)args->table
	);
}

typedef struct do_wasmtime_table_get_t {
	size_t context;
	size_t table;
	uint32_t index;
	uint32_t ok;
	size_t val;
} do_wasmtime_table_get_t;

void do_wasmtime_table_get(size_t arg0, size_t arg1) {
	do_wasmtime_table_get_t* args = (do_wasmtime_table_get_t*)(void*)arg0;
	args->ok = wasmtime_table_get(
		(wasmtime_context_t*)(void*)args->context,
		(const wasmtime_table_t*)(void*)args->table,
		args->index,
		(wasmtime_val_t*)(void*)args->val
	) ? 1 : 0;
}

typedef struct do_wasmtime_table_set_t {
	size_t context;
	size_t table;
	uint32_t index;
	uint32_t pad;
	size_t val;
	size_t error;
} do_wasmtime_table_set_t;

void do_wasmtime_table_set(size_t arg0, size_t arg1) {
	do_wasmtime_table_set_t* args = (do_wasmtime_table_set_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_table_set(
		(wasmtime_context_t*)(void*)args->context,
		(const wasmtime_table_t*)(void*)args->table,
		args->index,
		(const wasmtime_val_t*)(void*)args->val
	);
}

typedef struct do_wasmtime_table_size_t {
	size_t context;
	size_t table;
	uint32_t size;
	uint32_t pad;
} do_wasmtime_table_size_t;

void do_wasmtime_table_size(size_t arg0, size_t arg1) {
	do_wasmtime_table_size_t* args = (do_wasmtime_table_size_t*)(void*)arg0;
	args->size = wasmtime_table_size(
		(const wasmtime_context_t*)(void*)args->context,
		(const wasmtime_table_t*)(void*)args->table
	);
}

typedef struct do_wasmtime_table_grow_t {
	size_t context;
	size_t table;
	size_t init;
	uint32_t delta;
	uint32_t prev_size;
	size_t error;
} do_wasmtime_table_grow_t;

void do_wasmtime_table_grow(size_t arg0, size_t arg1) {
	do_wasmtime_table_grow_t* args = (do_wasmtime_table_grow_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_table_grow(
		(wasmtime_context_t*)(void*)args->context,
		(const wasmtime_table_t*)(void*)args->table,
		args->delta,
		(wasmtime_val_t*)(void*)args->init,
		&args->prev_size
	);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

type Table C.wasmtime_table_t

// NewTable Creates a new host-defined wasm table.
//
// \param store the store to create the table within
// \param ty the type of the table to create
// \param init the initial value for this table's elements
// \param table where to store the returned table
//
// This function does not take ownership of its parameters, but yields
// ownership of returned error. This function may return an error if the `init`
// value does not match `ty`, for example.
func NewTable(context *Context, tableType *TableType, init Val) (Table, *Error) {
	var table Table
	args := struct {
		context   uintptr
		tableType uintptr
		init      uintptr
		table     uintptr
		error     uintptr
	}{
		context:   uintptr(unsafe.Pointer(context)),
		tableType: uintptr(unsafe.Pointer(tableType)),
		table:     uintptr(unsafe.Pointer(&table)),
		init:      uintptr(unsafe.Pointer(&init)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_table_new), uintptr(unsafe.Pointer(&args)), 0)
	return table, (*Error)(unsafe.Pointer(args.error))
}

// Type returns the TableType of the Table
func (t *Table) Type(context *Context) *TableType {
	args := struct {
		context   uintptr
		table     uintptr
		tableType uintptr
	}{
		context: uintptr(unsafe.Pointer(context)),
		table:   uintptr(unsafe.Pointer(t)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_table_type), uintptr(unsafe.Pointer(&args)), 0)
	return (*TableType)(unsafe.Pointer(args.tableType))
}

// Get a value in a table.
//
// \param store the store that owns `table`
// \param table the table to access
// \param index the table index to access
// \param val where to store the table's value
//
// This function will attempt to access a table element. If a nonzero value is
// returned then `val` is filled in and is owned by the caller. Otherwise, zero
// is returned because the `index` is out-of-bounds.
func (t *Table) Get(context *Context, index uint32) (ok bool, val Val) {
	args := struct {
		context uintptr
		table   uintptr
		index   uint32
		ok      uint32
		value   uintptr
	}{
		context: uintptr(unsafe.Pointer(context)),
		table:   uintptr(unsafe.Pointer(t)),
		index:   index,
		value:   uintptr(unsafe.Pointer(&val)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_table_get), uintptr(unsafe.Pointer(&args)), 0)
	return args.ok != 0, val
}

// Set a value in a table.
//
// \param store the store that owns `table`
// \param table the table to write to
// \param index the table index to write
// \param value the value to store.
//
// This function will store `value` into the specified index in the table. This
// does not take ownership of any argument but yields ownership of the error.
// This function can fail if `value` has the wrong type for the table, or if
// `index` is out of bounds.
func (t *Table) Set(context *Context, index uint32, val Val) *Error {
	args := struct {
		context uintptr
		table   uintptr
		index   uint32
		_       uint32
		value   uintptr
		error   uintptr
	}{
		context: uintptr(unsafe.Pointer(context)),
		table:   uintptr(unsafe.Pointer(t)),
		index:   index,
		value:   uintptr(unsafe.Pointer(&val)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_table_set), uintptr(unsafe.Pointer(&args)), 0)
	return (*Error)(unsafe.Pointer(args.error))
}

// Size returns the size, in elements, of this table.
func (t *Table) Size(context *Context) uint32 {
	args := struct {
		context uintptr
		table   uintptr
		size    uint32
		_       uint32
	}{
		context: uintptr(unsafe.Pointer(context)),
		table:   uintptr(unsafe.Pointer(t)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_table_size), uintptr(unsafe.Pointer(&args)), 0)
	return args.size
}

// Grow a table.
//
// \param store the store that owns `table`
// \param table the table to grow
// \param delta the number of elements to grow the table by
// \param init the initial value for new table element slots
// \param prev_size where to store the previous size of the table before growth
//
// This function will attempt to grow the table by `delta` table elements. This
// can fail if `delta` would exceed the maximum size of the table or if `init`
// is the wrong type for this table. If growth is successful then `NULL` is
// returned and `prev_size` is filled in with the previous size of the table, in
// elements, before the growth happened.
//
// This function does not take ownership of its arguments.
func (t *Table) Grow(context *Context, delta uint32, init Val) (prevSize uint32, err *Error) {
	args := struct {
		context  uintptr
		table    uintptr
		init     uintptr
		delta    uint32
		prevSize uint32
		error    uintptr
	}{
		context: uintptr(unsafe.Pointer(context)),
		table:   uintptr(unsafe.Pointer(t)),
		init:    uintptr(unsafe.Pointer(&init)),
		delta:   delta,
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_table_grow), uintptr(unsafe.Pointer(&args)), 0)
	return args.prevSize, (*Error)(unsafe.Pointer(args.error))
}

func (t *Table) AsExtern() Extern {
	var ret Extern
	ret.SetTable(t)
	return ret
}
