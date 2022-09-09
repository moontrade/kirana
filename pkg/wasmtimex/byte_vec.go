package wasmtimex

/*
#include <wasmtime.h>

void do_wasm_byte_vec_delete(size_t arg0, size_t arg1) {
	wasm_byte_vec_delete((wasm_byte_vec_t*)(void*)arg0);
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"reflect"
	"unsafe"
)

type ByteVec struct {
	vec C.wasm_byte_vec_t
}

func (bv *ByteVec) Delete() {
	if bv.vec.data == nil {
		return
	}
	cgo.NonBlocking((*byte)(C.do_wasm_byte_vec_delete), uintptr(unsafe.Pointer(&bv.vec)), 0)
	bv.vec.data = nil
}

func (bv *ByteVec) Data() unsafe.Pointer {
	return unsafe.Pointer(bv.vec.data)
}

func (bv *ByteVec) Size() int {
	return int(bv.vec.size)
}

func (bv *ByteVec) Unsafe() []byte {
	if bv.vec.data == nil {
		return nil
	}
	size := int(bv.vec.size)
	if size < 1 {
		return nil
	}
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(bv.vec.data)),
		Len:  size,
		Cap:  size,
	}))
}

func (bv *ByteVec) UnsafeString() string {
	if bv.vec.data == nil {
		return ""
	}
	size := int(bv.vec.size)
	if size < 1 {
		return ""
	}
	if *(*byte)(unsafe.Add(unsafe.Pointer(bv.vec.data), size-1)) == 0 {
		size -= 1
	}
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(bv.vec.data)),
		Len:  size,
	}))
}

func (bv *ByteVec) Bytes() []byte {
	if bv.vec.data == nil {
		return nil
	}
	v := bv.Unsafe()
	if len(v) == 0 {
		return nil
	}
	n := make([]byte, len(v))
	copy(n, v)
	return n
}

//func (bv *ByteVec) Append(n []byte) []byte {
//	if bv.vec.data == nil {
//		return nil
//	}
//	v := bv.Unsafe()
//	if len(v) == 0 {
//		return nil
//	}
//	if len(n) < len(v) {
//		n = make([]byte, len(v))
//	} else {
//		n = n[0:len(v)]
//	}
//	copy(n, v)
//	return n
//}

func (bv *ByteVec) String() string {
	if bv.vec.data == nil {
		return ""
	}
	v := bv.UnsafeString()
	if len(v) == 0 {
		return ""
	}
	n := make([]byte, len(v))
	copy(n, v)
	return *(*string)(unsafe.Pointer(&n))
}

func (bv *ByteVec) ToOwned() string {
	if bv.vec.data == nil {
		return ""
	}
	v := bv.UnsafeString()
	if len(v) == 0 {
		return ""
	}
	n := make([]byte, len(v))
	copy(n, v)
	bv.Delete()
	return *(*string)(unsafe.Pointer(&n))
}
