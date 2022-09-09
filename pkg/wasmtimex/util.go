package wasmtimex

import (
	"reflect"
	"unsafe"
)

func boolToU32(v bool) uint32 {
	if v {
		return 1
	}
	return 0
}

func boolToU64(v bool) uint64 {
	if v {
		return 1
	}
	return 0
}

func dataPtr[T any](v []T) uintptr {
	return (*(*reflect.SliceHeader)(unsafe.Pointer(&v))).Data
}

func strDataPtr(s string) uintptr {
	return (*(*reflect.StringHeader)(unsafe.Pointer(&s))).Data
}
