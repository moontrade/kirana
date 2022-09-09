#ifndef WASMTIME_FFI_H
#define WASMTIME_FFI_H

#include <wasm.h>
#include <wasmtime.h>

#ifdef __cplusplus
extern "C" {
#endif

void wasmtime_epoch_thread_start(wasm_engine_t* engine, size_t nanos);

void wasmtime_epoch_thread_start_multiple(wasm_engine_t* engines[], size_t engine_count, size_t nanos);

void wasmtime_epoch_thread_stop();

#ifdef __cplusplus
}  // extern "C"
#endif

#endif // WASMTIME_FFI_H
