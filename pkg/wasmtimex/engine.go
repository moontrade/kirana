package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>
#include "ffi.h"

void do_wasmtime_engine_delete(size_t arg0, size_t arg1) {
	wasm_engine_delete(
		(wasm_engine_t*)arg0
	);
}

void do_wasmtime_engine_increment_epoch(size_t arg0, size_t arg1) {
	wasmtime_engine_increment_epoch(
		(wasm_engine_t*)arg0
	);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

// Engine is an instance of a wasmtime engine which is used to create a `Store`.
//
// Engines are a form of global configuration for wasm compilations and modules
// and such.
type Engine C.wasm_engine_t

// NewEngine creates a new `Engine` with default configuration.
func NewEngine() *Engine {
	return NewEngineWithConfigBuilder(DefaultConfig)
	//return (*Engine)(unsafe.Pointer(C.wasm_engine_new()))
}

// NewEngineWithConfig creates a new `Engine` with the `Config` provided
//
// Note that once a `Config` is passed to this method it cannot be used again.
func NewEngineWithConfig(config *Config) *Engine {
	if config == nil {
		panic("config already used")
	}
	engine := (*Engine)(unsafe.Pointer(C.wasm_engine_new_with_config((*C.wasm_config_t)(unsafe.Pointer(config)))))
	return engine
}

// NewEngineWithConfigBuilder creates a new `Engine` with the `Config` provided
//
// Note that once a `Config` is passed to this method it cannot be used again.
func NewEngineWithConfigBuilder(fn func(cfg *Config)) *Engine {
	if fn == nil {
		return NewEngine()
	}
	config := NewConfig()
	fn(config)
	engine := (*Engine)(unsafe.Pointer(C.wasm_engine_new_with_config((*C.wasm_config_t)(unsafe.Pointer(config)))))
	return engine
}

func (engine *Engine) Delete() {
	if engine == nil {
		return
	}
	RemoveEpochEngine(engine)
	cgo.NonBlocking((*byte)(C.do_wasmtime_engine_delete), uintptr(unsafe.Pointer(engine)), 0)
}

func (engine *Engine) ptr() *C.wasm_engine_t {
	return (*C.wasm_engine_t)(unsafe.Pointer(engine))
}

// IncrementEpoch will increase the current epoch number by 1 within the
// current engine which will cause any connected stores with their epoch
// deadline exceeded to now be interrupted.
//
// This method is safe to call from any goroutine.
func (engine *Engine) IncrementEpoch() {
	cgo.NonBlocking((*byte)(C.do_wasmtime_engine_increment_epoch), uintptr(unsafe.Pointer(engine.ptr())), 0)
}
