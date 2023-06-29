package logger

import (
	"errors"
	"github.com/moontrade/kirana/pkg/timex"
	"reflect"
	"time"
	"unsafe"
)

var (
	DurationAsString   = true
	RawFieldName       = "raw"
	DataFieldName      = "data"
	DurationFieldName  = "dur"
	DurationsFieldName = "durs"
	ErrorsFieldName    = "errors"

	EmptyMessage = ""
)

func Log(level Level, msg string) {
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
		this.writebuf = nil
		var (
			sp0 uintptr
			sp  uintptr
		)
		if gpp.syscallsp != 0 {
			sp0 = gpp.syscallsp
		} else {
			sp0 = gpp.sched.sp
		}
		if usesLR {
			sp = sp0 + callerSPOffset
		} else {
			sp = sp0 + callerSPOffset
		}
		(*reflect.SliceHeader)(unsafe.Pointer(&gpp.writebuf)).Data = sp
	})
	sp := *(*uintptr)(unsafe.Pointer(&gp.writebuf))
	gp.writebuf = prevWriteBuf
	gp.m.g0.writebuf = prevG0WriteBuf
	pc := *(*uintptr)(unsafe.Pointer(sp))
	procUnpinGp(gp)
	f := funcInfoMap.GetForPC(pc)
	_ = f
	b := gp.writebuf
	if b == nil {
		b = make([]byte, 0, 512)
	}

	//p := memory.Alloc(512)
	//b := p.Bytes(0, 512, 512)
	//b = b[:0-]
	//defer memory.Free(p)
	//b := pool.Alloc(512)[:0]
	//defer pool.Free(b[:])

	//b = AppendRecord(b, timex.Now(), 1, f, args...)
	b = AppendRecord(b, timex.Fastnow(), level, f, msg)
	//b = SerializeJSON(b, timex.Fastnow(), 1, f, args...)
	//logger.LogAttrs(context.Background(), slog.LevelInfo, msg, args...)
	//b = doLog(b, timex.NanoTime(), 0, f, args...)
	//b = doLog(b, 0, 0, nil, args...)
	//fmt.Println(string(b))
	gp.writebuf = b[:0]
}

const TestMessage = "Test logging, but use a somewhat realistic message length."

var (
	TestTime     = time.Date(2022, time.May, 1, 0, 0, 0, 0, time.UTC)
	TestString   = "7e3b3b2aaeff56a7108fe11e154200dd/7819479873059528190"
	TestInt      = 32768
	TestDuration = 23 * time.Second
	TestError    = errors.New("fail")
)
