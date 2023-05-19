package main

import (
	"github.com/moontrade/unsafe/memory"
	"reflect"
	"unsafe"
)

var OffHeap offHeap

type offHeap struct{}

func (offHeap) Allocate(size int) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(memory.Alloc(uintptr(size))),
		Len:  size,
		Cap:  size,
	}))
}

func (offHeap) Reallocate(size int, b []byte) []byte {
	if len(b) < 1 {
		if size < 1 {
			return nil
		}
		return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
			Data: uintptr(memory.Alloc(uintptr(size))),
			Len:  size,
			Cap:  size,
		}))
	}
	newAlloc := memory.Realloc(memory.Pointer(unsafe.Pointer(&b[0])), uintptr(size))
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(newAlloc),
		Len:  size,
		Cap:  size,
	}))
}

func (offHeap) Free(b []byte) {
	if cap(b) == 0 {
		return
	}
	memory.Free(memory.Pointer(unsafe.Pointer(&b[0])))
}
