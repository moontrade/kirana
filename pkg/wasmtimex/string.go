package wasmtimex

/*
#include <wasm.h>
#include <wasmtime.h>
*/
import "C"
import (
	"reflect"
	"unsafe"
)

type BorrowedString string

func BorrowedStringOf(data unsafe.Pointer, size int) BorrowedString {
	if data == nil || size <= 0 {
		return ""
	}
	if *(*byte)(unsafe.Add(data, size-1)) == 0 {
		return *(*BorrowedString)(unsafe.Pointer(&reflect.StringHeader{
			Data: 0,
			Len:  size - 1,
		}))
	} else {
		return *(*BorrowedString)(unsafe.Pointer(&reflect.StringHeader{
			Data: 0,
			Len:  size,
		}))
	}
}

func borrowWasmName(n *C.wasm_name_t) BorrowedString {
	if n == nil {
		return ""
	}
	return BorrowedStringOf(unsafe.Pointer(n.data), int(n.size))
}
