package atomicx

import (
	"sync/atomic"
	"unsafe"
)

// A Pointer is an atomic pointer of type *T. The zero value is a nil *T.
type Pointer[T any] struct {
	_ noCopy
	v unsafe.Pointer
}

// Get non-atomically loads value stored in x.
func (x *Pointer[T]) Get() *T { return (*T)(x.v) }

// Load atomically loads and returns the value stored in x.
func (x *Pointer[T]) Load() *T { return (*T)(atomic.LoadPointer(&x.v)) }

// Store atomically stores val into x.
func (x *Pointer[T]) Store(val *T) { atomic.StorePointer(&x.v, unsafe.Pointer(val)) }

// Swap atomically stores new into x and returns the previous value.
func (x *Pointer[T]) Swap(new *T) (old *T) {
	return (*T)(atomic.SwapPointer(&x.v, unsafe.Pointer(new)))
}

// Xchg atomically stores new into x and returns the previous value.
func (x *Pointer[T]) Xchg(new *T) (old *T) {
	return (*T)(unsafe.Pointer(Xchguintptr((*uintptr)(unsafe.Pointer(&x.v)), uintptr(unsafe.Pointer(new)))))
}

// CompareAndSwap executes the compare-and-swap operation for x.
func (x *Pointer[T]) CompareAndSwap(old, new *T) (swapped bool) {
	return atomic.CompareAndSwapPointer(&x.v, unsafe.Pointer(old), unsafe.Pointer(new))
}

// CAS executes the compare-and-swap operation for x.
func (x *Pointer[T]) CAS(old, new *T) (swapped bool) {
	return Casp1(&x.v, unsafe.Pointer(old), unsafe.Pointer(new))
}
