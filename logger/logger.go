package logger

import (
	"reflect"
	"runtime"
	"strconv"
	"unsafe"
)

func Info0() {
	gp := procPin()
	prevWriteBuf := gp.writebuf
	prevG0WriteBuf := gp.m.g0.writebuf
	gp.writebuf = *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(gp)),
		Len:  0,
		Cap:  0,
	}))
	gp.m.g0.writebuf = gp.writebuf
	systemstack(func() {
		this := getg()
		gpp := (*g)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.writebuf)).Data))
		//this.writebuf = nil
		var sp0 uintptr
		if gpp.syscallsp != 0 {
			sp0 = gpp.syscallsp
		} else {
			sp0 = gpp.sched.sp
		}
		sp := sp0 + callerSPOffset
		(*reflect.SliceHeader)(unsafe.Pointer(&gpp.writebuf)).Data = sp
	})
	sp := *(*uintptr)(unsafe.Pointer(&gp.writebuf))
	gp.writebuf = prevWriteBuf
	gp.m.g0.writebuf = prevG0WriteBuf
	pc := *(*uintptr)(unsafe.Pointer(sp))
	procUnpinGp(gp)
	f := funcInfoMap.GetForPC(pc)

	println("INF", f.String())
}

func Info() {
	pc, file, line, ok := runtime.Caller(1)
	_ = file
	_ = line
	_ = ok
	f := funcInfoMap.GetForPC(pc)
	println("INF", f.String(), file+":"+strconv.FormatInt(int64(line), 10))
}

func Err() {
	pc, file, line, ok := runtime.Caller(1)
	_ = file
	_ = line
	_ = ok
	f := funcInfoMap.GetForPC(pc)
	println("ERR", f.String(), file+":"+strconv.FormatInt(int64(line), 10))
}
