#include "_cgo_export.h"
#include "ffi.h"

#include <wasm.h>
#include <wasmtime.h>
#include <assert.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <unistd.h>
#include <thread>
#include <chrono>
#include <iostream>

static bool wasmtime_epoch_thread_stopping = false;
static std::unique_ptr<std::thread> epoch_thread = nullptr;

static void wasmtime_epoch_thread_run(void* engine_list, size_t engine_count, size_t nanos) {
    struct timespec ts;
    ts.tv_sec = 0;
    ts.tv_nsec = nanos;
//    if (nanos < 20000) nanos = 20;
//    else nanos /= 1000;

    if (engine_count == 1) {
        wasm_engine_t* engine = (wasm_engine_t*)engine_list;
        while (!wasmtime_epoch_thread_stopping) {
//            usleep(nanos);
            nanosleep(&ts, nullptr);
//            std::this_thread::sleep_for((std::chrono::nanoseconds)nanos);
            wasmtime_engine_increment_epoch(engine);
//            ts.tv_nsec = nanos;
//            std::cout << "hi\n";
        }
    } else if (engine_count > 1) {
        wasm_engine_t** engines = (wasm_engine_t**)engine_list;
        size_t i = 0;
        while (!wasmtime_epoch_thread_stopping) {
            nanosleep(&ts, nullptr);
//            usleep(nanos);
//            std::this_thread::sleep_for((std::chrono::nanoseconds)nanos);
            for (i = 0; i < engine_count; i++) {
                wasmtime_engine_increment_epoch(engines[i]);
            }
//            ts.tv_nsec = nanos;
//            std::cout << "hi\n" << std::endl;
        }
        free(engine_list);
    }
}

void wasmtime_epoch_thread_start(wasm_engine_t* engine, size_t nanos) {
    if (engine == nullptr) return;
    wasmtime_epoch_thread_stopping = false;
    epoch_thread = std::make_unique<std::thread>(wasmtime_epoch_thread_run, (void*)engine, 1, nanos);
}

void wasmtime_epoch_thread_start_multiple(wasm_engine_t* engines[], size_t engine_count, size_t nanos) {
    if (engines == nullptr || engine_count == 0) return;
    wasmtime_epoch_thread_stopping = false;

    if (engine_count == 1) {
        epoch_thread = std::make_unique<std::thread>(wasmtime_epoch_thread_run, (void*)engines[0], 1, nanos);
    } else {
        wasm_engine_t** copy = (wasm_engine_t**)malloc(engine_count*sizeof(wasm_engine_t*));
        memcpy((void*)copy, (void*)engines, engine_count*sizeof(wasm_engine_t*));
        epoch_thread = std::make_unique<std::thread>(wasmtime_epoch_thread_run, (void*)copy, engine_count, nanos);
    }
}

void wasmtime_epoch_thread_stop() {
    if (!epoch_thread) return;
    wasmtime_epoch_thread_stopping = true;
    epoch_thread->join();
    epoch_thread = nullptr;
    wasmtime_epoch_thread_stopping = false;
}

void wasmtime_sleep(size_t arg0, size_t arg1) {
	std::this_thread::sleep_for((std::chrono::nanoseconds)arg0);
}
