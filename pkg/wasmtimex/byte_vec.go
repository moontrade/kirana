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
	//bv.vec.data = nil
	//bv.vec.size = 0
}

func (bv *ByteVec) Data() unsafe.Pointer {
	return unsafe.Pointer(bv.vec.data)
}

func (bv *ByteVec) Size() int {
	return int(bv.vec.size)
}

func (bv *ByteVec) Unsafe() string {
	if bv.vec.data == nil {
		return ""
	}
	size := int(bv.vec.size)
	if size < 1 {
		return ""
	}
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(bv.vec.data)),
		Len:  size - 1,
	}))
}

func (bv *ByteVec) String() string {
	if bv.vec.data == nil {
		return ""
	}
	v := bv.Unsafe()
	if len(v) == 0 {
		return ""
	}
	n := make([]byte, int(bv.vec.size))
	copy(n, v)
	return *(*string)(unsafe.Pointer(&n))
}

func (bv *ByteVec) ToOwned() string {
	if bv.vec.data == nil {
		return ""
	}
	v := bv.Unsafe()
	if len(v) == 0 {
		return ""
	}
	n := make([]byte, len(v))
	copy(n, v)
	bv.Delete()
	return *(*string)(unsafe.Pointer(&n))
}
