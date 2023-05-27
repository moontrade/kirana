package runtimex

import (
	"reflect"
	"runtime"
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
	f := raw(funk)
	_ = f

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

func raw(f *runtime.Func) *_func {
	return (*_func)(unsafe.Pointer(f))
}

// A FuncID identifies particular functions that need to be treated
// specially by the runtime.
// Note that in some situations involving plugins, there may be multiple
// copies of a particular special runtime function.
// Note: this list must match the list in cmd/internal/objabi/funcid.go.
type funcID uint8

const (
	funcID_normal funcID = iota // not a special function
	funcID_abort
	funcID_asmcgocall
	funcID_asyncPreempt
	funcID_cgocallback
	funcID_debugCallV2
	funcID_gcBgMarkWorker
	funcID_goexit
	funcID_gogo
	funcID_gopanic
	funcID_handleAsyncEvent
	funcID_mcall
	funcID_morestack
	funcID_mstart
	funcID_panicwrap
	funcID_rt0_go
	funcID_runfinq
	funcID_runtime_main
	funcID_sigpanic
	funcID_systemstack
	funcID_systemstack_switch
	funcID_wrapper // any autogenerated code (hash/eq algorithms, method wrappers, etc.)
)

// A FuncFlag holds bits about a function.
// This list must match the list in cmd/internal/objabi/funcid.go.
type funcFlag uint8

const (
	// TOPFRAME indicates a function that appears at the top of its stack.
	// The traceback routine stop at such a function and consider that a
	// successful, complete traversal of the stack.
	// Examples of TOPFRAME functions include goexit, which appears
	// at the top of a user goroutine stack, and mstart, which appears
	// at the top of a system goroutine stack.
	funcFlag_TOPFRAME funcFlag = 1 << iota

	// SPWRITE indicates a function that writes an arbitrary value to SP
	// (any write other than adding or subtracting a constant amount).
	// The traceback routines cannot encode such changes into the
	// pcsp tables, so the function traceback cannot safely unwind past
	// SPWRITE functions. Stopping at an SPWRITE function is considered
	// to be an incomplete unwinding of the stack. In certain contexts
	// (in particular garbage collector stack scans) that is a fatal error.
	funcFlag_SPWRITE

	// ASM indicates that a function was implemented in assembly.
	funcFlag_ASM
)

// Layout of in-memory per-function information prepared by linker
// See https://golang.org/s/go12symtab.
// Keep in sync with linker (../cmd/link/internal/ld/pcln.go:/pclntab)
// and with package debug/gosym and with symtab.go in package runtime.
type _func struct {
	entryOff uint32 // start pc, as offset from moduledata.text/pcHeader.textStart
	nameOff  int32  // function name, as index into moduledata.funcnametab.

	args        int32  // in/out args size
	deferreturn uint32 // offset of start of a deferreturn call instruction from entry, if any.

	pcsp      uint32
	pcfile    uint32
	pcln      uint32
	npcdata   uint32
	cuOffset  uint32 // runtime.cutab offset of this function's CU
	startLine int32  // line number of start of function (func keyword/TEXT directive)
	funcID    funcID // set for certain special runtime functions
	flag      funcFlag
	_         [1]byte // pad
	nfuncdata uint8   // must be last, must end on a uint32-aligned boundary

	// The end of the struct is followed immediately by two variable-length
	// arrays that reference the pcdata and funcdata locations for this
	// function.

	// pcdata contains the offset into moduledata.pctab for the start of
	// that index's table. e.g.,
	// &moduledata.pctab[_func.pcdata[_PCDATA_UnsafePoint]] is the start of
	// the unsafe point table.
	//
	// An offset of 0 indicates that there is no table.
	//
	// pcdata [npcdata]uint32

	// funcdata contains the offset past moduledata.gofunc which contains a
	// pointer to that index's funcdata. e.g.,
	// *(moduledata.gofunc +  _func.funcdata[_FUNCDATA_ArgsPointerMaps]) is
	// the argument pointer map.
	//
	// An offset of ^uint32(0) indicates that there is no entry.
	//
	// funcdata [nfuncdata]uint32
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
