package message

import (
	"reflect"
	"unsafe"
)

//go:linkname mallocgc runtime.mallocgc
func mallocgc(size uintptr, typ unsafe.Pointer, needzero bool) unsafe.Pointer

func Alloc(size uintptr) unsafe.Pointer {
	return mallocgc(size, nil, false)
	//b := make([]byte, size)
	//return unsafe.Pointer(&b[0])
}

func AllocZeroed(size uintptr) unsafe.Pointer {
	return mallocgc(size, nil, true)
}

func Compare(a, b unsafe.Pointer, n uintptr) int {
	return Cmp(*(*string)(unsafe.Pointer(&reflect.StringHeader{Data: uintptr(a), Len: int(n)})),
		*(*string)(unsafe.Pointer(&reflect.StringHeader{Data: uintptr(b), Len: int(n)})))
}

//go:linkname Cmp runtime.cmpstring
func Cmp(a, b string) int

func Copy(dst, src unsafe.Pointer, n uintptr) {
	Move(dst, src, n)
	//memcpySlow(dst, src, n)
}

func memcpySlow(dst, src unsafe.Pointer, n uintptr) {
	//dstB := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
	//	Data: uintptr(dst),
	//	Len:  int(n),
	//	Cap:  int(n),
	//}))
	//srcB := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
	//	Data: uintptr(src),
	//	Len:  int(n),
	//	Cap:  int(n),
	//}))
	//copy(dstB, srcB)
}

// Move copies n bytes from "from" to "to".
//
// Move ensures that any pointer in "from" is written to "to" with
// an indivisible write, so that racy reads cannot observe a
// half-written pointer. This is necessary to prevent the garbage
// collector from observing invalid pointers, and differs from Memmove
// in unmanaged languages. However, Memmove is only required to do
// this if "from" and "to" may contain pointers, which can only be the
// case if "from", "to", and "n" are all be word-aligned.
//
// Implementations are in memmove_*.s.
//
//go:noescape
//go:linkname Move runtime.memmove
func Move(to, from unsafe.Pointer, n uintptr)

// Zero clears n bytes starting at ptr.
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
// Memmove for more details.
//
// The (CPU-specific) implementations of this function are in memclr_*.s.
//
//go:noescape
//go:linkname Zero runtime.memclrNoHeapPointers
func Zero(ptr unsafe.Pointer, n uintptr)

//func Memequal(a, b unsafe.Pointer, n uintptr) bool {
//	if a == nil {
//		return b == nil
//	}
//	return *(*string)(unsafe.Pointer(&reflect.SliceHeader{
//		Data: uintptr(a),
//		Len:  int(n),
//	})) == *(*string)(unsafe.Pointer(&reflect.SliceHeader{
//		Data: uintptr(b),
//		Len:  int(n),
//	}))
//}

//go:linkname Equals runtime.memequal
func Equals(a, b unsafe.Pointer, size uintptr) bool

//go:linkname Fastrand runtime.fastrand
func Fastrand() uint32
