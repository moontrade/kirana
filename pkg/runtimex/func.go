package runtimex

import (
	"github.com/moontrade/kirana/pkg/hashmap"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"unsafe"
)

type emptyInterface struct {
	typ  *rtype
	word unsafe.Pointer
}

type Runnable interface {
	Run()
}

var funcInfoMap = NewFuncInfoMap[any]()

func FuncToPointer(fn func()) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&fn))
}

func funcToPointer(fn func(uintptr, uintptr)) unsafe.Pointer {
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

func GetRunnable(runnable Runnable) *FuncInfo[any] {
	eface := *(*emptyInterface)(unsafe.Pointer(&runnable))
	typ := reflect.TypeOf(runnable)
	_ = typ
	println(eface.typ.ptrdata)
	return nil
}

func GetMethodSlow(
	object interface{},
	methodWrapperPC uintptr,
	methodName string,
) *FuncInfo[any] {
	return funcInfoMap.GetMethodSlow(object, methodWrapperPC, methodName)
}

type FuncInfoMap[T any] struct {
	data    *hashmap.Map[uintptr, *FuncInfo[T]]
	methods *hashmap.Map[uintptr, *FuncInfo[T]]
	mu      sync.Mutex
}

func NewFuncInfoMap[T any]() *FuncInfoMap[T] {
	return &FuncInfoMap[T]{
		data:    hashmap.New[uintptr, *FuncInfo[T]](1024, hashmap.HashUintptr),
		methods: hashmap.New[uintptr, *FuncInfo[T]](1024, hashmap.HashUintptr),
		mu:      sync.Mutex{},
	}
}

func (fip *FuncInfoMap[T]) GetMethod(methodWrapperPC uintptr) *FuncInfo[T] {
	return fip.methods.GetValue(methodWrapperPC)
}

func (fip *FuncInfoMap[T]) GetMethodSlow(
	object interface{},
	methodWrapperPC uintptr,
	methodName string,
) *FuncInfo[T] {
	method, _ := fip.methods.Get(methodWrapperPC)
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

	fi := findfunc(pc)
	funk := fi._Func()
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
	newMethods := fip.methods.Copy()
	newData := fip.data.Copy()
	newMethods.Set(methodWrapperPC, info)
	newMethods.Set(pc, info)
	newData.Set(methodWrapperPC, info)
	newData.Set(pc, info)
	fip.methods = newMethods
	fip.data = newData
	fip.mu.Unlock()
	return info
}

func (fip *FuncInfoMap[T]) Get(fn func()) *FuncInfo[T] {
	return fip.data.GetValue(FuncToPC(fn))
}

func (fip *FuncInfoMap[T]) GetPC(pc uintptr) *FuncInfo[T] {
	return fip.data.GetValue(pc)
}

func (fip *FuncInfoMap[T]) GetSlow(pc uintptr) *FuncInfo[T] {
	info := fip.data.GetValue(pc)
	if info != nil {
		return info
	}

	fn := findfunc(pc)
	funk := fn._Func()
	info = &FuncInfo[T]{
		pc:    pc,
		name:  funk.Name(),
		entry: funk.Entry(),
	}

	info.file, info.line = funk.FileLine(pc)
	info.formatted = info.file + ":" + strconv.FormatInt(int64(fn.startLine), 10)
	//info.formatted = info.file + ":" + strconv.FormatInt(int64(info.line), 10)
	info.debug = info.name + " -> " + info.formatted

	fip.mu.Lock()

	old := fip.data
	if existing := old.GetValue(pc); existing != nil {
		fip.mu.Unlock()
		return existing
	}
	data := old.Copy()
	data.Set(pc, info)
	fip.data = data
	fip.mu.Unlock()

	return info
}

func raw(f *runtime.Func) *_func {
	return (*_func)(unsafe.Pointer(f))
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
	formatted    string
	debug        string
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
func (f *FuncInfo[T]) Formatted() string      { return f.formatted }
func (f *FuncInfo[T]) Data() T                { return f.data }
func (f *FuncInfo[T]) SetData(data T)         { f.data = data }
func (f *FuncInfo[T]) String() string         { return f.debug }

type nameOff int32
type typeOff int32
type tflag uint8

// rtype is the common implementation of most values.
// It is embedded in other struct types.
//
// rtype must be kept in sync with ../runtime/type.go:/^type._type.
type rtype struct {
	size       uintptr
	ptrdata    uintptr // number of bytes in the type that can contain pointers
	hash       uint32  // hash of type; avoids computation in hash tables
	tflag      tflag   // extra type information flags
	align      uint8   // alignment of variable with this type
	fieldAlign uint8   // alignment of struct field with this type
	kind       uint8   // enumeration for C
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	equal     func(unsafe.Pointer, unsafe.Pointer) bool
	gcdata    *byte   // garbage collection data
	str       nameOff // string form
	ptrToThis typeOff // type for pointer to this type, may be zero
}
