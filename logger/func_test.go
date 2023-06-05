package logger

import (
	"reflect"
	"testing"
	"unsafe"
)

func BenchmarkFuncInfo(b *testing.B) {
	fn := func() {}
	m := NewFuncInfoMap()
	m.GetForFunc(fn)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetForFunc(fn)
	}
}

type Object struct{}

func (o *Object) Run() {
	runMe()
}

type Job struct{}

func (o Job) Run() {
	runMe()
}

func runMe() {

}

func BenchmarkMethodInfo(b *testing.B) {
	m := NewFuncInfoMap()
	o := &Object{}
	fn := o.Run
	ty := reflect.TypeOf(o)
	_ = ty
	method, _ := ty.MethodByName("Run")
	methodFunc := method.Func
	vv := reflect.ValueOf(fn)

	_ = methodFunc
	_ = vv
	//mmm := vv.Kind()
	mmm := methodFunc.Type().NumIn()
	first := methodFunc.Type().In(0)
	_ = first
	_ = mmm
	el := ty.Name()
	_ = el
	pc := FuncToPCUnsafe(*(*unsafe.Pointer)(unsafe.Pointer(&fn)))
	fi := m.GetForPC(pc)
	_ = fi

	m.GetMethodSlow(o, pc, "Run")
	//info := m.GetMethod(pc)
	//if info == nil {
	//	b.Fatal("info not found")
	//}

	fi2 := m.GetForPC(FuncToPC(runMe))
	_ = fi2

	me := func() {}

	fi3 := m.GetForPC(FuncToPC(func() {
		runMe()
	}))
	_ = fi3

	fi4 := m.GetForPC(FuncToPC(me))
	_ = fi4

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetMethod(FuncToPCUnsafe(*(*unsafe.Pointer)(unsafe.Pointer(&fn))))
	}
}

func BenchmarkMethodInfoSlow(b *testing.B) {
	m := NewFuncInfoMap()
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
