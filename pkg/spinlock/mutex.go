package spinlock

import (
	"github.com/moontrade/wormhole/pkg/atomicx"
	"runtime"
	"sync/atomic"
)

// Mutex is used as a lock for fast critical sections.
type Mutex struct {
	lock uint32
}

// Lock locks the SpinLock.
func (sl *Mutex) Lock() {
	for !sl.TryLock() {
		runtime.Gosched()
	}
}

// TryLock tries to lock the SpinLock.
func (sl *Mutex) TryLock() bool {
	return atomicx.Cas(&sl.lock, 0, 1)
	//return atomic.CompareAndSwapUint32(&sl.lock, 0, 1)
}

// Unlock unlocks the SpinLock.
func (sl *Mutex) Unlock() {
	atomic.StoreUint32(&sl.lock, 0)
}
