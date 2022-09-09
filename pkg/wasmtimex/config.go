package wasmtimex

// #include <wasm.h>
// #include <wasmtime.h>
// #include <stdlib.h>
import "C"
import (
	"unsafe"
)

func DefaultConfig(cfg *Config) {
	cfg.SetStrategy(StrategyCranelift)
	cfg.SetEpochInterruption(true)
	cfg.SetWasmMultiMemory(true)
	cfg.SetWasmMultiValue(true)
	cfg.SetWasmBulkMemory(true)
	cfg.SetWasmThreads(false)
	cfg.SetWasmMemory64(false)
	cfg.SetWasmReferenceTypes(false)
	cfg.SetWasmSIMD(true)
}

// Strategy is the compilation strategies for wasmtime
type Strategy C.wasmtime_strategy_t

const (
	// StrategyAuto will let wasmtime automatically pick an appropriate compilation strategy
	StrategyAuto Strategy = C.WASMTIME_STRATEGY_AUTO
	// StrategyCranelift will force wasmtime to use the Cranelift backend
	StrategyCranelift Strategy = C.WASMTIME_STRATEGY_CRANELIFT
)

// OptLevel decides what degree of optimization wasmtime will perform on generated machine code
type OptLevel C.wasmtime_opt_level_t

const (
	// OptLevelNone will perform no optimizations
	OptLevelNone OptLevel = C.WASMTIME_OPT_LEVEL_NONE
	// OptLevelSpeed will optimize machine code to be as fast as possible
	OptLevelSpeed OptLevel = C.WASMTIME_OPT_LEVEL_SPEED
	// OptLevelSpeedAndSize will optimize machine code for speed, but also optimize
	// to be small, sometimes at the cost of speed.
	OptLevelSpeedAndSize OptLevel = C.WASMTIME_OPT_LEVEL_SPEED_AND_SIZE
)

// ProfilingStrategy decides what sort of profiling to enable, if any.
type ProfilingStrategy C.wasmtime_profiling_strategy_t

const (
	// ProfilingStrategyNone means no profiler will be used
	ProfilingStrategyNone ProfilingStrategy = C.WASMTIME_PROFILING_STRATEGY_NONE
	// ProfilingStrategyJitdump will use the "jitdump" linux support
	ProfilingStrategyJitdump ProfilingStrategy = C.WASMTIME_PROFILING_STRATEGY_JITDUMP
)

// Config holds options used to create an Engine and customize its behavior.
type Config C.wasm_config_t

// NewConfig creates a new `Config` with all default options configured.
func NewConfig() *Config {
	return (*Config)(unsafe.Pointer(C.wasm_config_new()))
}

func (cfg *Config) Delete() {
	if cfg != nil {
		C.wasm_config_delete((*C.wasm_config_t)(unsafe.Pointer(cfg)))
	}
}

func (cfg *Config) ptr() *C.wasm_config_t {
	return (*C.wasm_config_t)(unsafe.Pointer(cfg))
}

// SetDebugInfo configures whether dwarf debug information for JIT code is enabled
func (cfg *Config) SetDebugInfo(enabled bool) {
	C.wasmtime_config_debug_info_set(cfg.ptr(), C.bool(enabled))
}

// SetWasmThreads configures whether the wasm threads proposal is enabled
func (cfg *Config) SetWasmThreads(enabled bool) {
	C.wasmtime_config_wasm_threads_set(cfg.ptr(), C.bool(enabled))
}

// SetWasmReferenceTypes configures whether the wasm reference types proposal is enabled
func (cfg *Config) SetWasmReferenceTypes(enabled bool) {
	C.wasmtime_config_wasm_reference_types_set(cfg.ptr(), C.bool(enabled))
}

// SetWasmSIMD configures whether the wasm SIMD proposal is enabled
func (cfg *Config) SetWasmSIMD(enabled bool) {
	C.wasmtime_config_wasm_simd_set(cfg.ptr(), C.bool(enabled))
}

// SetWasmBulkMemory configures whether the wasm bulk memory proposal is enabled
func (cfg *Config) SetWasmBulkMemory(enabled bool) {
	C.wasmtime_config_wasm_bulk_memory_set(cfg.ptr(), C.bool(enabled))
}

// SetWasmMultiValue configures whether the wasm multi value proposal is enabled
func (cfg *Config) SetWasmMultiValue(enabled bool) {
	C.wasmtime_config_wasm_multi_value_set(cfg.ptr(), C.bool(enabled))
}

// SetWasmMultiMemory configures whether the wasm multi memory proposal is enabled
func (cfg *Config) SetWasmMultiMemory(enabled bool) {
	C.wasmtime_config_wasm_multi_memory_set(cfg.ptr(), C.bool(enabled))
}

// SetWasmMemory64 configures whether the wasm memory64 proposal is enabled
func (cfg *Config) SetWasmMemory64(enabled bool) {
	C.wasmtime_config_wasm_memory64_set(cfg.ptr(), C.bool(enabled))
}

// SetConsumFuel configures whether fuel is enabled
func (cfg *Config) SetConsumeFuel(enabled bool) {
	C.wasmtime_config_consume_fuel_set(cfg.ptr(), C.bool(enabled))
}

// SetStrategy configures what compilation strategy is used to compile wasm code
func (cfg *Config) SetStrategy(strat Strategy) {
	C.wasmtime_config_strategy_set(cfg.ptr(), C.wasmtime_strategy_t(strat))
}

// SetCraneliftDebugVerifier configures whether the cranelift debug verifier will be active when
// cranelift is used to compile wasm code.
func (cfg *Config) SetCraneliftDebugVerifier(enabled bool) {
	C.wasmtime_config_cranelift_debug_verifier_set(cfg.ptr(), C.bool(enabled))
}

// SetCraneliftOptLevel configures the cranelift optimization level for generated code
func (cfg *Config) SetCraneliftOptLevel(level OptLevel) {
	C.wasmtime_config_cranelift_opt_level_set(cfg.ptr(), C.wasmtime_opt_level_t(level))
}

// SetProfiler configures what profiler strategy to use for generated code
func (cfg *Config) SetProfiler(profiler ProfilingStrategy) {
	C.wasmtime_config_profiler_set(cfg.ptr(), C.wasmtime_profiling_strategy_t(profiler))
}

// CacheConfigLoadDefault enables compiled code caching for this `Config` using the default settings
// configuration can be found.
//
// For more information about caching see
// https://bytecodealliance.github.io/wasmtime/cli-cache.html
func (cfg *Config) CacheConfigLoadDefault() *Error {
	return (*Error)(unsafe.Pointer(C.wasmtime_config_cache_config_load(cfg.ptr(), nil)))
}

// CacheConfigLoad enables compiled code caching for this `Config` using the settings specified
// in the configuration file `path`.
//
// For more information about caching and configuration options see
// https://bytecodealliance.github.io/wasmtime/cli-cache.html
func (cfg *Config) CacheConfigLoad(path string) *Error {
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))
	return (*Error)(unsafe.Pointer(C.wasmtime_config_cache_config_load(cfg.ptr(), cstr)))
}

// SetEpochInterruption enables epoch-based instrumentation of generated code to
// interrupt WebAssembly execution when the current engine epoch exceeds a
// defined threshold.
func (cfg *Config) SetEpochInterruption(enable bool) {
	C.wasmtime_config_epoch_interruption_set(cfg.ptr(), C.bool(enable))
}
