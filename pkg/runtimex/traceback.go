//go:build (amd64 || arm64) && go1.20

package runtimex

import (
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/goarch"
	"reflect"
	"runtime"
	"unsafe"
)

// The code in this file implements stack trace walking for all architectures.
// The most important fact about a given architecture is whether it uses a link register.
// On systems with link registers, the prologue for a non-leaf function stores the
// incoming value of LR at the bottom of the newly allocated stack frame.
// On systems without link registers (x86), the architecture pushes a return PC during
// the call instruction, so the return PC ends up above the stack frame.
// In this file, the return PC is always called LR, no matter how it was found.

const usesLR = goarch.MinFrameSize > 0

const framepointer_enabled = goarch.GOARCH == "amd64" || goarch.GOARCH == "arm64"

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

// A stkframe holds information about a single physical stack frame.
type stkframe struct {
	// fn is the function being run in this frame. If there is
	// inlining, this is the outermost function.
	fn funcInfo

	// pc is the program counter within fn.
	//
	// The meaning of this is subtle:
	//
	// - Typically, this frame performed a regular function call
	//   and this is the return PC (just after the CALL
	//   instruction). In this case, pc-1 reflects the CALL
	//   instruction itself and is the correct source of symbolic
	//   information.
	//
	// - If this frame "called" sigpanic, then pc is the
	//   instruction that panicked, and pc is the correct address
	//   to use for symbolic information.
	//
	// - If this is the innermost frame, then PC is where
	//   execution will continue, but it may not be the
	//   instruction following a CALL. This may be from
	//   cooperative preemption, in which case this is the
	//   instruction after the call to morestack. Or this may be
	//   from a signal or an un-started goroutine, in which case
	//   PC could be any instruction, including the first
	//   instruction in a function. Conventionally, we use pc-1
	//   for symbolic information, unless pc == fn.entry(), in
	//   which case we use pc.
	pc      uintptr
	startSP uintptr
	prevSP  uintptr
	delta   uintptr

	// continpc is the PC where execution will continue in fn, or
	// 0 if execution will not continue in this frame.
	//
	// This is usually the same as pc, unless this frame "called"
	// sigpanic, in which case it's either the address of
	// deferreturn or 0 if this frame will never execute again.
	//
	// This is the PC to use to look up GC liveness for this frame.
	continpc uintptr

	lr   uintptr // program counter at caller aka link register
	sp   uintptr // stack pointer at pc
	fp   uintptr // stack pointer at caller aka frame pointer
	varp uintptr // top of local variables
	argp uintptr // pointer to function arguments
}

type funcInfo struct {
	//_func uintptr
	*_func
	datap uintptr
}

func (f funcInfo) valid() bool {
	return f._func != nil
}

func (f funcInfo) _Func() *runtime.Func {
	return (*runtime.Func)(unsafe.Pointer(f._func))
}

// isInlined reports whether f should be re-interpreted as a *funcinl.
func (f *_func) isInlined() bool {
	return f.entryOff == ^uint32(0) // see comment for funcinl.ones
}

// entry returns the entry PC for f.
func (f funcInfo) entry() uintptr {
	//return f.datap.textAddr(f.entryOff)
	return f._Func().Entry()
}

// alignUp rounds n up to a multiple of a. a must be a power of 2.
func alignUp(n, a uintptr) uintptr {
	return (n + a - 1) &^ (a - 1)
}

// alignDown rounds n down to a multiple of a. a must be a power of 2.
func alignDown(n, a uintptr) uintptr {
	return n &^ (a - 1)
}

// divRoundUp returns ceil(n / a).
func divRoundUp(n, a uintptr) uintptr {
	// a is generally a power of two. This will get inlined and
	// the compiler will optimize the division.
	return (n + a - 1) / a
}

//go:linkname gcallers runtime.gcallers
func gcallers(gp *g, skip int, pcbuf []uintptr) int

//go:linkname Systemstack runtime.systemstack
func Systemstack(fn func())

//go:linkname findfunc runtime.findfunc
func findfunc(pc uintptr) funcInfo

//go:linkname funcspdelta runtime.funcspdelta
func funcspdelta(f funcInfo, targetpc uintptr, cache *pcvalueCache) int32

//go:linkname pcdatavalue runtime.pcdatavalue
func pcdatavalue(f funcInfo, table uint32, targetpc uintptr, cache *pcvalueCache) int32

//go:linkname pcdatastart runtime.pcdatastart
func pcdatastart(f funcInfo, table uint32) uint32

//go:linkname pcvalue runtime.pcvalue
func pcvalue(f funcInfo, off uint32, targetpc uintptr, strict bool) int32

//go:linkname funcdata runtime.funcdata
func funcdata(f funcInfo, i uint8) unsafe.Pointer

//go:linkname step runtime.step
func step(p []byte, pc *uintptr, val *int32, first bool) (newp []byte, ok bool)

//go:linkname findnull runtime.findnull
func findnull(s *byte) int

//go:linkname funcname runtime.funcname
func funcname(f funcInfo) string

//go:linkname cfuncname runtime.cfuncname
func cfuncname(f funcInfo) *byte

//go:linkname gostringnocopy runtime.gostringnocopy
func gostringnocopy(str *byte) string

//go:linkname elideWrapperCalling runtime.elideWrapperCalling
func elideWrapperCalling(id funcID) bool

//go:linkname tracebackCgoContext runtime.tracebackCgoContext
func tracebackCgoContext(pcbuf *uintptr, printing bool, ctxt uintptr, n, max int) int

// PCDATA and FUNCDATA table indexes.
//
// See funcdata.h and ../cmd/internal/objabi/funcdata.go.
const (
	_PCDATA_UnsafePoint   = 0
	_PCDATA_StackMapIndex = 1
	_PCDATA_InlTreeIndex  = 2
	_PCDATA_ArgLiveIndex  = 3

	_FUNCDATA_ArgsPointerMaps    = 0
	_FUNCDATA_LocalsPointerMaps  = 1
	_FUNCDATA_StackObjects       = 2
	_FUNCDATA_InlTree            = 3
	_FUNCDATA_OpenCodedDeferInfo = 4
	_FUNCDATA_ArgInfo            = 5
	_FUNCDATA_ArgLiveInfo        = 6
	_FUNCDATA_WrapInfo           = 7

	_ArgsSizeUnknown = -0x80000000
)

const (
	// PCDATA_UnsafePoint values.
	_PCDATA_UnsafePointSafe   = -1 // Safe for async preemption
	_PCDATA_UnsafePointUnsafe = -2 // Unsafe for async preemption

	// _PCDATA_Restart1(2) apply on a sequence of instructions, within
	// which if an async preemption happens, we should back off the PC
	// to the start of the sequence when resume.
	// We need two so we can distinguish the start/end of the sequence
	// in case that two sequences are next to each other.
	_PCDATA_Restart1 = -3
	_PCDATA_Restart2 = -4

	// Like _PCDATA_RestartAtEntry, but back to function entry if async
	// preempted.
	_PCDATA_RestartAtEntry = -5
)

// inlinedCall is the encoding of entries in the FUNCDATA_InlTree table.
type inlinedCall struct {
	funcID    funcID // type of the called function
	_         [3]byte
	nameOff   int32 // offset into pclntab for name of called function
	parentPc  int32 // position of an instruction whose source position is the call site (offset from entry)
	startLine int32 // line number of start of function (func keyword/TEXT directive)
}

type pcvalueCache struct {
	entries [2][8]pcvalueCacheEnt
}

type pcvalueCacheEnt struct {
	// targetpc and off together are the key of this cache entry.
	targetpc uintptr
	off      uint32
	// val is the value of this cached pcvalue entry.
	val int32
}

func traceback(pc0, sp0, lr0 uintptr, gp *g, skip int, pcbuf *uintptr, max int) stkframe {
	if gp.syscallsp != 0 {
		pc0 = gp.syscallpc
		sp0 = gp.syscallsp
		if usesLR {
			lr0 = 0
		}
	} else {
		pc0 = gp.sched.pc
		sp0 = gp.sched.sp
		if usesLR {
			lr0 = gp.sched.lr
		}
	}

	var frame stkframe
	frame.pc = pc0
	frame.sp = sp0
	frame.startSP = sp0
	if usesLR {
		frame.lr = lr0
	}

	// If the PC is zero, it's likely a nil function call.
	// Start in the caller's frame.
	if frame.pc == 0 {
		if usesLR {
			frame.pc = *(*uintptr)(unsafe.Pointer(frame.sp))
			frame.lr = 0
		} else {
			frame.pc = uintptr(*(*uintptr)(unsafe.Pointer(frame.sp)))
			frame.sp += goarch.PtrSize
		}
	}

	// runtime/internal/atomic functions call into kernel helpers on
	// arm < 7. See runtime/internal/atomic/sys_linux_arm.s.
	//
	// Start in the caller's frame.
	//if goarch.GOARCH == "arm" && goarm < 7 && runtime.GOOS == "linux" && frame.pc&0xffff0000 == 0xffff0000 {
	//	// Note that the calls are simple BL without pushing the return
	//	// address, so we use LR directly.
	//	//
	//	// The kernel helpers are frameless leaf functions, so SP and
	//	// LR are not touched.
	//	frame.pc = frame.lr
	//	frame.lr = 0
	//}

	f := findfunc(frame.pc)
	if !f.valid() {
		return frame
	}
	frame.fn = f

	var cache pcvalueCache
	deltaCount := uintptr(0)
	wasPanic := false
	cgoCtxt := gp.cgoCtxt

	n := 0
	lastFuncID := funcID_normal
	for n < max {
		f = frame.fn
		if f.pcsp == 0 {
			// No frame information, must be external function, like race support.
			// See golang.org/issue/13568.
			break
		}
		// Compute function info flags.
		flag := f.flag
		if f.funcID == funcID_cgocallback {
			// cgocallback does write SP to switch from the g0 to the curg stack,
			// but it carefully arranges that during the transition BOTH stacks
			// have cgocallback frame valid for unwinding through.
			// So we don't need to exclude it with the other SP-writing functions.
			flag &^= funcFlag_SPWRITE
		}
		if frame.pc == pc0 && frame.sp == sp0 && pc0 == gp.syscallpc && sp0 == gp.syscallsp {
			// Some Syscall functions write to SP, but they do so only after
			// saving the entry PC/SP using entersyscall.
			// Since we are using the entry PC/SP, the later SP write doesn't matter.
			flag &^= funcFlag_SPWRITE
		}
		// Found an actual function.
		// Derive frame pointer and link register.
		if frame.fp == 0 {
			delta := uintptr(funcspdelta(f, frame.pc, &cache))
			deltaCount += delta
			frame.delta = deltaCount
			frame.fp = frame.sp + delta
			//frame.fp = frame.sp + uintptr(funcspdelta(f, frame.pc))
			if !usesLR {
				// On x86, call instruction pushes return PC before entering new function.
				frame.fp += goarch.PtrSize
			}
		}

		var flr funcInfo
		if flag&funcFlag_TOPFRAME != 0 {
			// This function marks the top of the stack. Stop the traceback.
			frame.lr = 0
			flr = funcInfo{}
		} else if flag&funcFlag_SPWRITE != 0 && n > 0 {
			// The function we are in does a write to SP that we don't know
			// how to encode in the spdelta table. Examples include context
			// switch routines like runtime.gogo but also any code that switches
			// to the g0 stack to run host C code. Since we can't reliably unwind
			// the SP (we might not even be on the stack we think we are),
			// we stop the traceback here.
			// This only applies for profiling signals (callback == nil).
			//
			// For a GC stack traversal (callback != nil), we should only see
			// a function when it has voluntarily preempted itself on entry
			// during the stack growth check. In that case, the function has
			// not yet had a chance to do any writes to SP and is safe to unwind.
			// isAsyncSafePoint does not allow assembly functions to be async preempted,
			// and preemptPark double-checks that SPWRITE functions are not async preempted.
			// So for GC stack traversal we leave things alone (this if body does not execute for n == 0)
			// at the bottom frame of the stack. But farther up the stack we'd better not
			// find any.
			frame.lr = 0
			flr = funcInfo{}
		} else {
			var lrPtr uintptr
			if usesLR {
				if n == 0 && frame.sp < frame.fp || frame.lr == 0 {
					lrPtr = frame.sp
					frame.lr = *(*uintptr)(unsafe.Pointer(lrPtr))
				}
			} else {
				if frame.lr == 0 {
					lrPtr = frame.fp - goarch.PtrSize
					frame.lr = uintptr(*(*uintptr)(unsafe.Pointer(lrPtr)))
				}
			}
			flr = findfunc(frame.lr)
		}

		frame.varp = frame.fp
		if !usesLR {
			// On x86, call instruction pushes return PC before entering new function.
			frame.varp -= goarch.PtrSize
		}

		// For architectures with frame pointers, if there's
		// a frame, then there's a saved frame pointer here.
		//
		// NOTE: This code is not as general as it looks.
		// On x86, the ABI is to save the frame pointer word at the
		// top of the stack frame, so we have to back down over it.
		// On arm64, the frame pointer should be at the bottom of
		// the stack (with R29 (aka FP) = RSP), in which case we would
		// not want to do the subtraction here. But we started out without
		// any frame pointer, and when we wanted to add it, we didn't
		// want to break all the assembly doing direct writes to 8(RSP)
		// to set the first parameter to a called function.
		// So we decided to write the FP link *below* the stack pointer
		// (with R29 = RSP - 8 in Go functions).
		// This is technically ABI-compatible but not standard.
		// And it happens to end up mimicking the x86 layout.
		// Other architectures may make different decisions.
		if frame.varp > frame.sp && framepointer_enabled {
			frame.varp -= goarch.PtrSize
		}

		frame.argp = frame.fp + goarch.MinFrameSize

		// Determine frame's 'continuation PC', where it can continue.
		// Normally this is the return address on the stack, but if sigpanic
		// is immediately below this function on the stack, then the frame
		// stopped executing due to a trap, and frame.pc is probably not
		// a safe point for looking up liveness information. In this panicking case,
		// the function either doesn't return at all (if it has no defers or if the
		// defers do not recover) or it returns from one of the calls to
		// deferproc a second time (if the corresponding deferred func recovers).
		// In the latter case, use a deferreturn call site as the continuation pc.
		frame.continpc = frame.pc
		if wasPanic {
			if frame.fn.deferreturn != 0 {
				frame.continpc = frame.fn.entry() + uintptr(frame.fn.deferreturn) + 1
				// Note: this may perhaps keep return variables alive longer than
				// strictly necessary, as we are using "function has a defer statement"
				// as a proxy for "function actually deferred something". It seems
				// to be a minor drawback. (We used to actually look through the
				// gp._defer for a defer corresponding to this function, but that
				// is hard to do with defer records on the stack during a stack copy.)
				// Note: the +1 is to offset the -1 that
				// stack.go:getStackMap does to back up a return
				// address make sure the pc is in the CALL instruction.
			} else {
				frame.continpc = 0
			}
		}

		if pcbuf != nil {
			pc := frame.pc
			// backup to CALL instruction to read inlining info (same logic as below)
			tracepc := pc
			// Normally, pc is a return address. In that case, we want to look up
			// file/line information using pc-1, because that is the pc of the
			// call instruction (more precisely, the last byte of the call instruction).
			// Callers expect the pc buffer to contain return addresses and do the
			// same -1 themselves, so we keep pc unchanged.
			// When the pc is from a signal (e.g. profiler or segv) then we want
			// to look up file/line information using pc, and we store pc+1 in the
			// pc buffer so callers can unconditionally subtract 1 before looking up.
			// See issue 34123.
			// The pc can be at function entry when the frame is initialized without
			// actually running code, like runtime.mstart.
			//if (n == 0 && flags&_TraceTrap != 0) || waspanic || pc == f.entry() {
			//	pc++
			//} else {
			//	tracepc--
			//}
			tracepc--

			// If there is inlining info, record the inner frames.
			if inldata := funcdata(f, _FUNCDATA_InlTree); inldata != nil {
				inltree := (*[1 << 20]inlinedCall)(inldata)
				for {
					ix := pcdatavalue(f, _PCDATA_InlTreeIndex, tracepc, &cache)
					if ix < 0 {
						break
					}
					if inltree[ix].funcID == funcID_wrapper && elideWrapperCalling(lastFuncID) {
						// ignore wrappers
					} else if skip > 0 {
						skip--
					} else if n < max {
						(*[1 << 20]uintptr)(unsafe.Pointer(pcbuf))[n] = pc
						n++
					}
					lastFuncID = inltree[ix].funcID
					// Back up to an instruction in the "caller".
					tracepc = frame.fn.entry() + uintptr(inltree[ix].parentPc)
					pc = tracepc + 1
				}
			}
			// Record the main frame.
			if f.funcID == funcID_wrapper && elideWrapperCalling(lastFuncID) {
				// Ignore wrapper functions (except when they trigger panics).
			} else if skip > 0 {
				skip--
			} else if n < max {
				(*[1 << 20]uintptr)(unsafe.Pointer(pcbuf))[n] = pc
				n++
			}
			lastFuncID = f.funcID
			n-- // offset n++ below
		}

		n++

		if f.funcID == funcID_cgocallback && len(cgoCtxt) > 0 {
			ctxt := cgoCtxt[len(cgoCtxt)-1]
			cgoCtxt = cgoCtxt[:len(cgoCtxt)-1]

			// skip only applies to Go frames.
			// callback != nil only used when we only care
			// about Go frames.
			if skip == 0 {
				n = tracebackCgoContext(pcbuf, false, ctxt, n, max)
			}
		}

		wasPanic = f.funcID == funcID_sigpanic
		injectedCall := wasPanic || f.funcID == funcID_asyncPreempt || f.funcID == funcID_debugCallV2

		// Do not unwind past the bottom of the stack.
		if !flr.valid() {
			break
		}

		if frame.pc == frame.lr && frame.sp == frame.fp {
			// If the next frame is identical to the current frame, we cannot make progress.
			print("runtime: traceback stuck. pc=", frame.pc, " sp=", frame.sp, "\n")
			//tracebackHexdump(stack, &frame, frame.sp)
			panic("traceback stuck")
		}

		// Unwind to next frame.
		frame.fn = flr
		frame.pc = frame.lr
		frame.lr = 0
		frame.prevSP = frame.sp
		frame.sp = frame.fp
		frame.fp = 0

		// On link register architectures, sighandler saves the LR on stack
		// before faking a call.
		if usesLR && injectedCall {
			x := *(*uintptr)(unsafe.Pointer(frame.sp))
			frame.sp += alignUp(goarch.MinFrameSize, goarch.StackAlign)
			f = findfunc(frame.pc)
			frame.fn = f
			if !f.valid() {
				frame.pc = x
			} else if funcspdelta(f, frame.pc, &cache) == 0 {
				frame.lr = x
			}
		}
	}

	return frame
}

var callerSPOffset uintptr = 0

func init() {
	profileOffset()
}

func profileOffset() uintptr {
	//wg.Add(1)
	//go profileOffset0()
	profileOffset0()
	//wg.Wait()
	return callerSPOffset
}

func profileOffset0() {
	profileOffset1()
	//_ = getCallerFunc()
	//wg.Done()
}

func profileOffset1() {
	profileOffset2()
}

func profileOffset2() {
	gp := getg()
	Systemstack(func() {
		var pcbuf [3]uintptr
		//var f = traceback(^uintptr(0), ^uintptr(0), 0, gp, 0, &pcbuf[0], len(pcbuf))
		var f = traceback(^uintptr(0), ^uintptr(0), 0, gp, 0, nil, len(pcbuf))
		callerSPOffset = f.prevSP - f.startSP

		//for _, v := range pcbuf {
		//	if v > 0 {
		//		f := funcInfoMap.GetForPC(v)
		//		if f != nil {
		//			println(f.name, f.formatted)
		//		}
		//	}
		//}
	})
	retSP := uintptr(0)
	_ = retSP
	println("stack offset", callerSPOffset)
}

var count counter.Counter

func getCallerFunc() *FuncInfo[any] {
	gp := procPin()
	prevWriteBuf := gp.writebuf
	prevG0WriteBuf := gp.m.g0.writebuf
	gp.writebuf = *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(gp)),
		Len:  0,
		Cap:  0,
	}))
	gp.m.g0.writebuf = gp.writebuf

	Systemstack(func() {
		this := getg()
		gpp := (*g)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.writebuf)).Data))
		if gpp != gp {
			count.Incr()
		}
		gpp = gp
		this.writebuf = nil
		var (
			//pc0 uintptr
			sp0 uintptr
			//lr0 uintptr
		)

		if gpp.syscallsp != 0 {
			sp0 = gpp.syscallsp
			if usesLR {
				//lr0 = 0
			}
		} else {
			sp0 = gpp.sched.sp
			if usesLR {
				//lr0 = gpp.sched.lr
			}
		}

		var sp uintptr
		if usesLR {
			sp = sp0 + callerSPOffset
		} else {
			sp = sp0 + callerSPOffset
		}
		(*reflect.SliceHeader)(unsafe.Pointer(&gpp.writebuf)).Data = sp
	})

	retSP := *(*uintptr)(unsafe.Pointer(&gp.writebuf))
	gp.writebuf = prevWriteBuf
	gp.m.g0.writebuf = prevG0WriteBuf
	pc := *(*uintptr)(unsafe.Pointer(retSP))
	procUnpinGp(gp)
	info := funcInfoMap.GetSlow(pc)
	return info
}

func VisitCaller(fn func(*FuncInfo[any])) *FuncInfo[any] {
	//gp := getg()
	gp := procPin()
	prevWriteBuf := gp.writebuf
	prevG0WriteBuf := gp.m.g0.writebuf
	gp.writebuf = *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(gp)),
		Len:  0,
		Cap:  0,
	}))
	gp.m.g0.writebuf = gp.writebuf

	Systemstack(func() {
		this := getg()
		gpp := (*g)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.writebuf)).Data))
		this.writebuf = nil
		var (
			//pc0 uintptr
			sp0 uintptr
			//lr0 uintptr
		)

		if gpp == nil {
			count.Incr()
			return
		}

		if gpp.syscallsp != 0 {
			sp0 = gpp.syscallsp
			if usesLR {
				//lr0 = 0
			}
		} else {
			sp0 = gpp.sched.sp
			if usesLR {
				//lr0 = gpp.sched.lr
			}
		}

		var sp uintptr
		if usesLR {
			sp = sp0 + callerSPOffset
		} else {
			sp = sp0 + callerSPOffset
		}
		(*reflect.SliceHeader)(unsafe.Pointer(&gpp.writebuf)).Data = sp
	})

	retSP := *(*uintptr)(unsafe.Pointer(&gp.writebuf))
	gp.writebuf = prevWriteBuf
	gp.m.g0.writebuf = prevG0WriteBuf
	pc := *(*uintptr)(unsafe.Pointer(retSP))
	procUnpinGp(gp)
	info := funcInfoMap.GetSlow(pc)
	fn(info)
	return info
}

type VisitArgs struct {
	Args uintptr
	Func *FuncInfo[any]
}

func VisitCallerArgs(args *VisitArgs, fn func(visitArgs *VisitArgs)) {
	gp := procPin()
	prevWriteBuf := gp.writebuf
	prevG0WriteBuf := gp.m.g0.writebuf
	gp.writebuf = *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(gp)),
		Len:  0,
		Cap:  0,
	}))
	gp.m.g0.writebuf = gp.writebuf

	Systemstack(func() {
		this := getg()
		gpp := (*g)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.writebuf)).Data))
		this.writebuf = nil
		var (
			//pc0 uintptr
			sp0 uintptr
			//lr0 uintptr
		)

		if gpp == nil {
			count.Incr()
			return
		}

		if gpp.syscallsp != 0 {
			sp0 = gpp.syscallsp
			if usesLR {
				//lr0 = 0
			}
		} else {
			sp0 = gpp.sched.sp
			if usesLR {
				//lr0 = gpp.sched.lr
			}
		}

		var sp uintptr
		if usesLR {
			sp = sp0 + callerSPOffset
		} else {
			sp = sp0 + callerSPOffset
		}
		(*reflect.SliceHeader)(unsafe.Pointer(&gpp.writebuf)).Data = sp
	})

	retSP := *(*uintptr)(unsafe.Pointer(&gp.writebuf))
	gp.writebuf = prevWriteBuf
	gp.m.g0.writebuf = prevG0WriteBuf
	pc := *(*uintptr)(unsafe.Pointer(retSP))
	procUnpinGp(gp)
	args.Func = funcInfoMap.GetSlow(pc)
	fn(args)
}
