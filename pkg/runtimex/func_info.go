package runtimex

import (
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

var funcInfoMap = NewFuncInfoMap[any]()

func FuncToPointer(fn func()) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&fn))
}

func FuncFromPointer(p unsafe.Pointer) func() {
	return *(*func())(unsafe.Pointer(&p))
}

func RuntimeFuncOf(fn func()) *runtime.Func {
	return runtime.FuncForPC(FuncToPC(fn))
}

func FuncToPC(fn func()) uintptr {
	return uintptr(*(*unsafe.Pointer)(FuncToPointer(fn)))
}

func FuncToPCUnsafe(p unsafe.Pointer) uintptr {
	return uintptr(*(*unsafe.Pointer)(p))
}

func GetFuncInfo(fn func()) *FuncInfo[any] {
	return funcInfoMap.Get(fn)
}

func GetFuncInfoUnsafe(pc uintptr) *FuncInfo[any] {
	return funcInfoMap.GetSlow(pc)
}

func GetMethod(pc uintptr) *FuncInfo[any] {
	return funcInfoMap.GetMethod(pc)
}

func GetMethodSlow(
	object interface{},
	methodWrapperPC uintptr,
	methodName string,
) *FuncInfo[any] {
	return funcInfoMap.GetMethodSlow(object, methodWrapperPC, methodName)
}

type FuncInfoMap[T any] struct {
	data    map[uintptr]*FuncInfo[T]
	methods map[uintptr]*FuncInfo[T]
	mu      sync.Mutex
}

func NewFuncInfoMap[T any]() *FuncInfoMap[T] {
	return &FuncInfoMap[T]{
		data:    make(map[uintptr]*FuncInfo[T]),
		methods: make(map[uintptr]*FuncInfo[T]),
		mu:      sync.Mutex{},
	}
}

func (fip *FuncInfoMap[T]) GetMethod(methodWrapperPC uintptr) *FuncInfo[T] {
	return fip.methods[methodWrapperPC]
}

func (fip *FuncInfoMap[T]) GetMethodSlow(
	object interface{},
	methodWrapperPC uintptr,
	methodName string,
) *FuncInfo[T] {
	method := fip.methods[methodWrapperPC]
	if method != nil {
		return method
	}

	t := reflect.TypeOf(object)
	m, ok := t.MethodByName(methodName)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
		if !ok {
			m, ok = t.MethodByName(methodName)
		}
	}
	if !ok {
		m, ok = t.MethodByName(methodName)
	}
	if !ok {
		return nil
	}

	pc := m.Func.Pointer()
	if pc == 0 {
		return nil
	}

	funk := runtime.FuncForPC(pc)
	info := &FuncInfo[T]{
		pc:           pc,
		name:         funk.Name(),
		entry:        funk.Entry(),
		methodName:   m.Name,
		method:       m,
		numIn:        m.Type.NumIn(),
		numOut:       m.Type.NumOut(),
		receiver:     t.Name(),
		receiverSize: t.Size(),
	}
	info.file, info.line = funk.FileLine(info.entry)

	if info.numIn > 0 {
		info.in = make([]reflect.Type, info.numIn)
		for i := 0; i < info.numIn; i++ {
			info.in[i] = m.Type.In(i)
		}
	}
	if info.numOut > 0 {
		info.out = make([]reflect.Type, info.numOut)
		for i := 0; i < info.numOut; i++ {
			info.out[i] = m.Type.Out(i)
		}
	}

	fip.mu.Lock()
	defer fip.mu.Unlock()

	newMethods := make(map[uintptr]*FuncInfo[T])
	newData := make(map[uintptr]*FuncInfo[T])

	for k, v := range fip.methods {
		newMethods[k] = v
	}
	for k, v := range fip.data {
		newData[k] = v
	}

	newMethods[methodWrapperPC] = info
	newMethods[pc] = info
	newData[methodWrapperPC] = info
	newData[pc] = info
	fip.methods = newMethods
	fip.data = newData
	return info
}

func (fip *FuncInfoMap[T]) Get(fn func()) *FuncInfo[T] {
	return fip.data[FuncToPC(fn)]
}

func (fip *FuncInfoMap[T]) GetPC(pc uintptr) *FuncInfo[T] {
	return fip.data[pc]
}

func (fip *FuncInfoMap[T]) GetSlow(pc uintptr) *FuncInfo[T] {
	info := fip.data[pc]
	if info != nil {
		return info
	}

	funk := runtime.FuncForPC(pc)
	info = &FuncInfo[T]{
		pc:    pc,
		name:  funk.Name(),
		entry: funk.Entry(),
	}
	info.file, info.line = funk.FileLine(info.entry)

	fip.mu.Lock()
	defer fip.mu.Unlock()
	old := fip.data
	if existing := old[pc]; existing != nil {
		return existing
	}
	data := make(map[uintptr]*FuncInfo[T])
	for k, v := range old {
		data[k] = v
	}
	data[pc] = info
	fip.data = data

	return info
}

type FuncInfo[T any] struct {
	pc           uintptr
	entry        uintptr
	receiver     string
	receiverSize uintptr
	method       reflect.Method
	methodName   string
	numIn        int
	in           []reflect.Type
	numOut       int
	out          []reflect.Type
	name         string
	file         string
	line         int
	data         T
}

func (f *FuncInfo[T]) PC() uintptr            { return f.pc }
func (f *FuncInfo[T]) Entry() uintptr         { return f.entry }
func (f *FuncInfo[T]) Name() string           { return f.name }
func (f *FuncInfo[T]) MethodName() string     { return f.methodName }
func (f *FuncInfo[T]) Method() reflect.Method { return f.method }
func (f *FuncInfo[T]) NumIn() int             { return f.numIn }
func (f *FuncInfo[T]) NumOut() int            { return f.numOut }
func (f *FuncInfo[T]) Receiver() string       { return f.receiver }
func (f *FuncInfo[T]) ReceiverSize() uintptr  { return f.receiverSize }
func (f *FuncInfo[T]) File() string           { return f.file }
func (f *FuncInfo[T]) Line() int              { return f.line }
func (f *FuncInfo[T]) Data() T                { return f.data }
func (f *FuncInfo[T]) SetData(data T)         { f.data = data }
