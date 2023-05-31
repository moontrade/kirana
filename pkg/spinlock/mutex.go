// Copyright 2019 Andy Pan & Dietoad. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package spinlock

import (
	"runtime"
	"sync/atomic"
)

// Mutex is used as a lock for fast critical sections.
type Mutex uint32

const maxBackoff = 16

// Lock locks the SpinLock.
func (sl *Mutex) Lock() {
	backoff := 1
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		// Leverage the exponential backoff algorithm, see https://en.wikipedia.org/wiki/Exponential_backoff.
		for i := 0; i < backoff; i++ {
			runtime.Gosched()
		}
		if backoff < maxBackoff {
			backoff <<= 1
		}
	}
}

// TryLock tries to lock the SpinLock.
func (sl *Mutex) TryLock() bool {
	return atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1)
}

// Unlock unlocks the SpinLock.
func (sl *Mutex) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}
