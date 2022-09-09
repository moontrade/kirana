package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>

typedef struct do_wasmtime_store_new_t {
	size_t engine;
	size_t data;
	size_t finalizer;
	size_t store;
} do_wasmtime_store_new_t;

void do_wasmtime_store_new(size_t arg0, size_t arg1) {
	do_wasmtime_store_new_t* args = (do_wasmtime_store_new_t*)(void*)arg0;
	args->store = (size_t)(void*)wasmtime_store_new(
		(wasm_engine_t*)(void*)args->engine,
		(void*)args->data,
		(void (*)(void*))(void*)args->finalizer
	);
}

void do_wasmtime_store_context(size_t arg0, size_t arg1) {
	*((wasmtime_context_t**)arg1) = wasmtime_store_context(
		(wasmtime_store_t*)(void*)arg0
	);
}

void do_wasmtime_store_delete(size_t arg0, size_t arg1) {
	wasmtime_store_delete(
		(wasmtime_store_t*)(void*)arg0
	);
}

void do_wasmtime_context_get_data(size_t arg0, size_t arg1) {
	*((void**)arg1) = wasmtime_context_get_data(
		(wasmtime_context_t*)(void*)arg0
	);
}

void do_wasmtime_context_set_data(size_t arg0, size_t arg1) {
	wasmtime_context_set_data(
		(wasmtime_context_t*)(void*)arg0,
		(void*)arg1
	);
}

void do_wasmtime_context_gc(size_t arg0, size_t arg1) {
	wasmtime_context_gc(
		(wasmtime_context_t*)(void*)arg0
	);
}

typedef struct do_wasmtime_context_add_fuel_t {
	size_t context;
	uint64_t fuel;
	size_t error;
} do_wasmtime_context_add_fuel_t;

void do_wasmtime_context_add_fuel(size_t arg0, size_t arg1) {
	do_wasmtime_context_add_fuel_t* args = (do_wasmtime_context_add_fuel_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_context_add_fuel(
		(wasmtime_context_t*)(void*)args->context,
		(uint64_t)args->fuel
	);
}

typedef struct do_wasmtime_context_fuel_consumed_t {
	size_t context;
	uint64_t fuel;
	bool result;
} do_wasmtime_context_fuel_consumed_t;

void do_wasmtime_context_fuel_consumed(size_t arg0, size_t arg1) {
	do_wasmtime_context_fuel_consumed_t* args = (do_wasmtime_context_fuel_consumed_t*)(void*)arg0;
	args->result = wasmtime_context_fuel_consumed(
		(wasmtime_context_t*)(void*)args->context,
		&args->fuel
	);
}

typedef struct do_wasmtime_context_consume_fuel_t {
	size_t context;
	uint64_t fuel;
	uint64_t remaining;
	size_t error;
} do_wasmtime_context_consume_fuel_t;

void do_wasmtime_context_consume_fuel(size_t arg0, size_t arg1) {
	do_wasmtime_context_consume_fuel_t* args = (do_wasmtime_context_consume_fuel_t*)(void*)arg0;
	args->error = (size_t)(void*)wasmtime_context_consume_fuel(
		(wasmtime_context_t*)(void*)args->context,
		args->fuel,
		&args->remaining
	);
}

void do_wasmtime_context_set_epoch_deadline(size_t arg0, size_t arg1) {
	wasmtime_context_set_epoch_deadline(
		(wasmtime_context_t*)(void*)arg0,
		(uint64_t)arg1
	);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

type Finalizer *[0]byte

// Store
//
//	\typedef wasmtime_store_t
//	\brief Convenience alias for #wasmtime_store_t
//
//	\struct wasmtime_store
//	\brief Storage of WebAssembly objects
//
//	A store is the unit of isolation between WebAssembly instances in an
//	embedding of Wasmtime. Values in one #wasmtime_store_t cannot flow into
//	another #wasmtime_store_t. Stores are cheap to create and cheap to dispose.
//	It's expected that one-off stores are common in embeddings.
//
//	Objects stored within a #wasmtime_store_t are referenced with integer handles
//	rather than interior pointers. This means that most APIs require that the
//	store be explicitly passed in, which is done via #wasmtime_context_t. It is
//	safe to move a #wasmtime_store_t to any thread at any time. A store generally
//	cannot be concurrently used, however.
type Store C.wasmtime_store_t

// Context
// \typedef wasmtime_context_t
// \brief Convenience alias for #wasmtime_context
//
// \struct wasmtime_context
// \brief An interior pointer into a #wasmtime_store_t which is used as
// "context" for many functions.
//
// This context pointer is used pervasively throughout Wasmtime's API. This can be
// acquired from #wasmtime_store_context or #wasmtime_caller_context. The
// context pointer for a store is the same for the entire lifetime of a store,
// so it can safely be stored adjacent to a #wasmtime_store_t itself.
//
// Usage of a #wasmtime_context_t must not outlive the original
// #wasmtime_store_t. Additionally, #wasmtime_context_t can only be used in
// situations where it has explicitly been granted access to doing so. For
// example finalizers cannot use #wasmtime_context_t because they are not given
// access to it.
type Context C.wasmtime_context_t

// Data returns the user-specified data associated with the specified store
func (c *Context) Data() uintptr {
	var data uintptr
	cgo.NonBlocking((*byte)(C.do_wasmtime_context_get_data), uintptr(unsafe.Pointer(c)), uintptr(unsafe.Pointer(&data)))
	return data
}

// SetData overwrites the user-specified data associated with this store.
//
// Note that this does not execute the original finalizer for the provided data,
// and the original finalizer will be executed for the provided data when the
// store is deleted.
func (c *Context) SetData(data uintptr) {
	cgo.NonBlocking((*byte)(C.do_wasmtime_context_set_data), uintptr(unsafe.Pointer(c)), data)
}

// GC performs a garbage collection within the given context.
//
// Garbage collects `externref`s that are used within this store. Any
// `externref`s that are discovered to be unreachable by other code or objects
// will have their finalizers run.
//
// The `context` argument must not be NULL.
func (c *Context) GC(data uintptr) {
	cgo.NonBlocking((*byte)(C.do_wasmtime_context_gc), uintptr(unsafe.Pointer(c)), data)
}

// AddFuel adds fuel to this context's store for wasm to consume while executing.
//
// For this method to work fuel consumption must be enabled via
// #wasmtime_config_consume_fuel_set. By default a store starts with 0 fuel
// for wasm to execute with (meaning it will immediately trap).
// This function must be called for the store to have
// some fuel to allow WebAssembly to execute.
//
// Note that at this time when fuel is entirely consumed it will cause
// wasm to trap. More usages of fuel are planned for the future.
//
// If fuel is not enabled within this store then an error is returned. If fuel
// is successfully added then NULL is returned.
func (c *Context) AddFuel(fuel uint64) *Error {
	args := struct {
		context uintptr
		fuel    uint64
		error   uintptr
	}{
		context: uintptr(unsafe.Pointer(c)),
		fuel:    fuel,
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_context_add_fuel), uintptr(unsafe.Pointer(&args)), 0)
	return (*Error)(unsafe.Pointer(args.error))
}

// FuelConsumed returns the amount of fuel consumed by this context's store execution
//
//	so far.
//
//	If fuel consumption is not enabled via #wasmtime_config_consume_fuel_set
//	then this function will return false. Otherwise, true is returned and the
//	fuel parameter is filled in with fuel consuemd so far.
//
//	Also note that fuel, if enabled, must be originally configured via
//	#wasmtime_context_add_fuel.
func (c *Context) FuelConsumed() (uint64, bool) {
	args := struct {
		context uintptr
		fuel    uint64
		result  bool
	}{
		context: uintptr(unsafe.Pointer(c)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_context_fuel_consumed), uintptr(unsafe.Pointer(&args)), 0)
	return args.fuel, args.result
}

// ConsumeFuel attempts to manually consume fuel from the store.
//
//	If fuel consumption is not enabled via #wasmtime_config_consume_fuel_set then
//	this function will return an error. Otherwise this will attempt to consume
//	the specified amount of `fuel` from the store. If successful the remaining
//	amount of fuel is stored into `remaining`. If `fuel` couldn't be consumed
//	then an error is returned.
//
//	Also note that fuel, if enabled, must be originally configured via
//	#wasmtime_context_add_fuel.
func (c *Context) ConsumeFuel(fuel uint64) (remaining uint64, err *Error) {
	args := struct {
		context   uintptr
		fuel      uint64
		remaining uint64
		error     uintptr
	}{
		context: uintptr(unsafe.Pointer(c)),
		fuel:    fuel,
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_context_consume_fuel), uintptr(unsafe.Pointer(&args)), 0)
	return args.remaining, (*Error)(unsafe.Pointer(args.error))
}

// SetEpochDeadline
// Configures the relative deadline at which point WebAssembly code will trap.
//
//	This function configures the store-local epoch deadline after which point
//	WebAssembly code will trap.
//
//	See also #wasmtime_config_epoch_interruption_set.
func (c *Context) SetEpochDeadline(ticksBeyondCurrent uint64) {
	cgo.NonBlocking(
		(*byte)(C.wasmtime_context_set_epoch_deadline),
		uintptr(unsafe.Pointer(c)),
		uintptr(ticksBeyondCurrent),
	)
}

// NewStore
//
//	\brief Creates a new store within the specified engine.
//
//	\param engine the compilation environment with configuration this store is
//	connected to
//	\param data user-provided data to store, can later be acquired with
//	#wasmtime_context_get_data.
//	\param finalizer an optional finalizer for `data`
//
//	This function creates a fresh store with the provided configuration settings.
//	The returned store must be deleted with #wasmtime_store_delete.
func NewStore(engine *Engine, data uintptr, finalizer Finalizer) *Store {
	args := struct {
		engine    uintptr
		data      uintptr
		finalizer uintptr
		store     uintptr
	}{
		engine:    uintptr(unsafe.Pointer(engine)),
		data:      data,
		finalizer: uintptr(unsafe.Pointer(finalizer)),
	}
	cgo.NonBlocking((*byte)(C.do_wasmtime_store_new), uintptr(unsafe.Pointer(&args)), 0)
	return (*Store)(unsafe.Pointer(args.store))
}

// Context returns the interior #wasmtime_context_t pointer to this store
func (s *Store) Context() *Context {
	var ctx *Context
	cgo.NonBlocking((*byte)(C.do_wasmtime_store_context), uintptr(unsafe.Pointer(s)), uintptr(unsafe.Pointer(&ctx)))
	return ctx
}

func (s *Store) Delete() {
	cgo.NonBlocking((*byte)(C.do_wasmtime_store_delete), uintptr(unsafe.Pointer(s)), 0)
}
