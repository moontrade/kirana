package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>
#include "ffi.h"
*/
import "C"
import (
	"sync"
	"time"
	"unsafe"
)

var (
	epochThreadMu      sync.Mutex
	epochInterval      = time.Millisecond
	epochThreadRunning = false
	epochEngines       []*Engine
)

func NewEpochDeadline(duration time.Duration) uint64 {
	if duration <= 0 {
		return 0
	}
	if duration < epochInterval {
		return 1
	}
	return uint64(duration / epochInterval)
}

func IsEpochThreadRunning() bool {
	return epochThreadRunning
}

func EpochInterval() time.Duration {
	return epochInterval
}

// StartEpochThread starts a native thread that increments the epoch at an interval specified
// by the supplied duration. If a thread is active, then it will stop that thread and start
// a new thread with the specified duration.
func StartEpochThread(duration time.Duration, engine *Engine) {
	epochThreadMu.Lock()
	defer epochThreadMu.Unlock()
	if engine == nil {
		epochEngines = nil
		epochThreadRunning = false
		C.wasmtime_epoch_thread_stop()
		return
	}
	if epochEngines == nil {
		epochEngines = append(epochEngines, engine)
	} else {
		found := false
		for _, e := range epochEngines {
			if e == engine {
				found = true
				break
			}
		}
		if !found {
			epochEngines = append(epochEngines, engine)
		}
	}
	if duration < time.Microsecond*50 {
		duration = time.Microsecond * 50
	}
	epochInterval = duration
	C.wasmtime_epoch_thread_stop()

	if len(epochEngines) > 1 {
		C.wasmtime_epoch_thread_start_multiple(
			(**C.wasm_engine_t)(unsafe.Pointer(dataPtr(epochEngines))),
			C.size_t(len(epochEngines)),
			C.size_t(duration))
	} else {
		C.wasmtime_epoch_thread_start(engine.ptr(), C.size_t(duration))
	}
	epochThreadRunning = true
}

func StartEpochThreadMultiple(duration time.Duration, engines ...*Engine) {
	epochThreadMu.Lock()
	defer epochThreadMu.Unlock()
	if duration < time.Microsecond*50 {
		duration = time.Microsecond * 50
	}
	if len(engines) == 0 {
		epochEngines = nil
		epochThreadRunning = false
		C.wasmtime_epoch_thread_stop()
		return
	}
	if len(epochEngines) == 0 {
		epochEngines = append(epochEngines, engines...)
	} else {
		for _, engine := range engines {
			found := false
			for _, e := range epochEngines {
				if e == engine {
					found = true
					break
				}
			}
			if !found {
				epochEngines = append(epochEngines, engine)
			}
		}
	}
	if duration < time.Microsecond*50 {
		duration = time.Microsecond * 50
	}
	epochInterval = duration
	C.wasmtime_epoch_thread_stop()
	if len(epochEngines) > 1 {
		C.wasmtime_epoch_thread_start_multiple(
			(**C.wasm_engine_t)(unsafe.Pointer(dataPtr(epochEngines))),
			C.size_t(len(epochEngines)),
			C.size_t(duration))
	} else {
		C.wasmtime_epoch_thread_start(epochEngines[0].ptr(), C.size_t(duration))
	}
	epochThreadRunning = true
}

func RemoveEpochEngine(engine *Engine) {
	epochThreadMu.Lock()
	defer epochThreadMu.Unlock()
	if len(epochEngines) == 0 {
		return
	}
	if len(epochEngines) == 1 {
		if epochEngines[0] != engine {
			return
		}
		epochEngines = nil
		epochThreadRunning = false
		C.wasmtime_epoch_thread_stop()
		return
	}

	engines := make([]*Engine, len(epochEngines)-1)
	for _, e := range epochEngines {
		if e != engine {
			engines = append(engines, e)
		}
	}

	if len(engines) == len(epochEngines) {
		return
	}

	epochEngines = engines
	C.wasmtime_epoch_thread_stop()

	if len(epochEngines) > 1 {
		C.wasmtime_epoch_thread_start_multiple(
			(**C.wasm_engine_t)(unsafe.Pointer(dataPtr(epochEngines))),
			C.size_t(len(epochEngines)),
			C.size_t(epochInterval))
	} else {
		C.wasmtime_epoch_thread_start(engine.ptr(), C.size_t(epochInterval))
	}
	epochThreadRunning = true
}

// StopEpochThread stops the epoch thread if running.
func StopEpochThread() {
	epochThreadMu.Lock()
	defer epochThreadMu.Unlock()
	if !epochThreadRunning {
		return
	}
	C.wasmtime_epoch_thread_stop()
	epochEngines = nil
	epochThreadRunning = false
}
