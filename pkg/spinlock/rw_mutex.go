package spinlock

import (
	"github.com/moontrade/wormhole/pkg/atomicx"
	"runtime"
	"sync"
	"sync/atomic"
)

// An RWMutex is a reader/writer mutual exclusion lock.
// The lock can be held by an arbitrary number of readers
// or a single writer.
// RWMutexes can be created as part of other
// structures; the zero value for a RWMutex is
// an unlocked mutex.
type RWMutex struct {
	state uint32
}

const (
	rwmutexUnlocked            = 0
	rwmutexWrite               = 1 << 0 // Bit 1 is used as a flag for write mode
	rwmutexReadOffset          = 1 << 1 // Bits 2-32 store the number of readers
	rwmutexUnderflow           = ^uint32(rwmutexWrite)
	rwmutexWriterUnset         = ^uint32(rwmutexWrite - 1)
	rwmutexWriterUnsetInt32    = ^int32(rwmutexWrite - 1)
	rwmutexReaderDecrease      = ^uint32(rwmutexReadOffset - 1)
	rwmutexReaderDecreaseInt32 = ^int32(rwmutexReadOffset - 1)
)

// RLock locks rw for reading.
func (rw *RWMutex) RLock() {
	// Increase the number of readers by 1
	//state := atomic.AddUint32(&rw.state, rwmutexReadOffset)
	state := atomicx.Xadd(&rw.state, rwmutexReadOffset)

	// If no write bits are set, the read lock was successfully acquired
	if state&rwmutexWrite == 0 {
		return
	}

	// Otherwise we have to wait until the write bits become unset.
	// Afterwards the RWMutex is in read mode.
	for {
		if state := atomic.LoadUint32(&rw.state); state&rwmutexWrite == 0 {
			return
		}
		runtime.Gosched()
	}
}

// TryRLock tries to lock rw for reading.
// If a lock for reading can not be acquired immediately, false is returned.
func (rw *RWMutex) TryRLock() bool {
	// Increase the number of readers by 1
	//state := atomic.AddUint32(&rw.state, rwmutexReadOffset)
	state := atomicx.Xadd(&rw.state, rwmutexReadOffset)

	// If no write bits are set, the read lock was successfully acquired
	if state&rwmutexWrite == 0 {
		return true
	}

	// Undo
	atomicx.Xadd(&rw.state, rwmutexReaderDecreaseInt32)
	//atomic.AddUint32(&rw.state, rwmutexReaderDecrease)
	return false
}

// RUnlock undoes a single RLock call;
// it does not affect other simultaneous readers.
// It is a run-time error if rw is not locked for reading
// on entry to RUnlock.
func (rw *RWMutex) RUnlock() {
	// Decrease the number of readers by 1
	state := atomicx.Xadd(&rw.state, rwmutexReaderDecreaseInt32)
	//state := atomic.AddUint32(&rw.state, rwmutexReaderDecrease)

	// Check for underflow
	if state&rwmutexUnderflow == rwmutexUnderflow {
		panic("spinlock: RUnlock of unlocked RWMutex")
	}
}

// Lock locks rw for writing.
// If the lock is already locked for reading or writing,
// Lock blocks until the lock is available.
func (rw *RWMutex) Lock() {
	for !atomicx.Cas(&rw.state, rwmutexUnlocked, rwmutexWrite) {
		//for !atomic.CompareAndSwapUint32(&rw.state, rwmutexUnlocked, rwmutexWrite) {
		runtime.Gosched()
	}
}

// TryLock tries to lock rw for writing.
// If the lock for writing can not be acquired immediately, false is returned.
func (rw *RWMutex) TryLock() bool {
	//return atomic.CompareAndSwapUint32(&rw.state, rwmutexUnlocked, rwmutexWrite)
	return atomicx.Cas(&rw.state, rwmutexUnlocked, rwmutexWrite)
}

// Unlock unlocks rw for writing.  It is a run-time error if rw is
// not locked for writing on entry to Unlock.
//
// As with Mutexes, a locked RWMutex is not associated with a particular
// goroutine.  One goroutine may RLock (Lock) an RWMutex and then
// arrange for another goroutine to RUnlock (Unlock) it.
func (rw *RWMutex) Unlock() {
	// Unset the Write bit
	atomicx.Xadd(&rw.state, rwmutexWriterUnsetInt32)
	//atomic.AddUint32(&rw.state, rwmutexWriterUnset)
	//state := atomic.AddUint32(&rw.state, rwmutexWriterUnset)
	//if state&rwmutexWrite > 0 {
	//	panic("sync: Unlock of unlocked RWMutex")
	//}
}

// RLocker returns a Locker interface that implements
// the Lock and Unlock methods by calling rw.RLock and rw.RUnlock.
func (rw *RWMutex) RLocker() sync.Locker {
	return (*rlocker)(rw)
}

type rlocker RWMutex

func (r *rlocker) Lock()   { (*RWMutex)(r).RLock() }
func (r *rlocker) Unlock() { (*RWMutex)(r).RUnlock() }
