package pool

import (
	"reflect"
	"unsafe"
)

// memclrNoHeapPointers clears n bytes starting at ptr.
//
// Usually you should use typedmemclr. memclrNoHeapPointers should be
// used only when the caller knows that *ptr contains no heap pointers
// because either:
//
// *ptr is initialized memory and its type is pointer-free, or
//
// *ptr is uninitialized memory (e.g., memory that's being reused
// for a new allocation) and hence contains only "junk".
//
// memclrNoHeapPointers ensures that if ptr is pointer-aligned, and n
// is a multiple of the pointer size, then any pointer-aligned,
// pointer-sized portion is cleared atomically. Despite the function
// name, this is necessary because this function is the underlying
// implementation of typedmemclr and memclrHasPointers. See the doc of
// memmove for more details.
//
// The (CPU-specific) implementations of this function are in memclr_*.s.
//
//go:linkname memclrNoHeapPointers runtime.memclrNoHeapPointers
func memclrNoHeapPointers(ptr unsafe.Pointer, n uintptr)

// memclrHasPointers clears n bytes of typed memory starting at ptr.
// The caller must ensure that the type of the object at ptr has
// pointers, usually by checking typ.ptrdata. However, ptr
// does not have to point to the start of the allocation.
//
//go:linkname memclrHasPointers runtime.memclrHasPointers
func memclrHasPointers(ptr unsafe.Pointer, n uintptr)

func ptrdataOf(v any) uintptr {
	t := reflect.TypeOf(v)
	typ := (*emptyInterface)(unsafe.Pointer(&t))
	return typ.value.ptrdata
}

// emptyInterface is the header for an interface{} value.
type emptyInterface struct {
	typ   unsafe.Pointer
	value *rtype
}

type rtype struct {
	size    uintptr
	ptrdata uintptr
}
