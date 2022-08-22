package runtimex

import (
	"testing"
	"unsafe"
)

func BenchmarkFuncInfo(b *testing.B) {
	fn := func() {}
	m := NewFuncInfoMap[any]()
	m.Get(fn)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(fn)
	}
}

type Object struct{}

func (o *Object) Run() {}

func BenchmarkMethodInfo(b *testing.B) {
	m := NewFuncInfoMap[any]()
	o := &Object{}
	fn := o.Run
	pc := FuncToPCUnsafe(*(*unsafe.Pointer)(unsafe.Pointer(&fn)))
	m.GetMethodSlow(o, pc, "Run")
	info := m.GetMethod(pc)
	if info == nil {
		b.Fatal("info not found")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetMethod(FuncToPCUnsafe(*(*unsafe.Pointer)(unsafe.Pointer(&fn))))
	}
}

func BenchmarkMethodInfoSlow(b *testing.B) {
	m := NewFuncInfoMap[any]()
	o := &Object{}
	fn := o.Run
	pc := FuncToPCUnsafe(*(*unsafe.Pointer)(unsafe.Pointer(&fn)))
	m.GetMethodSlow(o, pc, "Run")
	info := m.GetMethod(pc)
	if info == nil {
		b.Fatal("info not found")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetMethodSlow(o, pc, "Run")
		//m.GetMethod(FuncToPCUnsafe(*(*unsafe.Pointer)(unsafe.Pointer(&fn))))
	}
}
