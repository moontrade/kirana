package wasmtimex

// #include <wasmtime.h>
import "C"
import (
	"unsafe"
)

// Engine is an instance of a wasmtime engine which is used to create a `Store`.
//
// Engines are a form of global configuration for wasm compilations and modules
// and such.
type Engine C.wasm_engine_t

// NewEngine creates a new `Engine` with default configuration.
func NewEngine() *Engine {
	return (*Engine)(unsafe.Pointer(C.wasm_engine_new()))
}

// NewEngineWithConfig creates a new `Engine` with the `Config` provided
//
// Note that once a `Config` is passed to this method it cannot be used again.
//func NewEngineWithConfig(config *Config) *Engine {
//	if config == nil {
//		panic("config already used")
//	}
//	engine := (*Engine)(unsafe.Pointer(C.wasm_engine_new_with_config(config.ptr())))
//	return engine
//}

func NewEngineWithConfig(fn func(cfg *Config)) *Engine {
	if fn == nil {
		return NewEngine()
	}
	config := NewConfig()
	fn(config)
	engine := (*Engine)(unsafe.Pointer(C.wasm_engine_new_with_config(config.ptr())))
	return engine
}

func (engine *Engine) Close() error {
	if engine == nil {
		return nil
	}
	C.wasm_engine_delete(engine.ptr())
	return nil
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
	C.wasmtime_engine_increment_epoch(engine.ptr())
}
