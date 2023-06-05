//go:generate easyjson -all $GOFILE
package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/moontrade/kirana/logger/slog"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

type testMessage struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (t *testMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(*t)
}

func TestInfo(t *testing.T) {
	//Info(
	//	"started_in", (time.Millisecond * 20).String(),
	//	"enabled", true,
	//	errors.New("first"),
	//	"started in: %s", time.Millisecond*25,
	//)
	//
	//Info(
	//	"started_in", time.Millisecond*20,
	//	"enabled", true,
	//	errors.New("first"),
	//	"started",
	//)
	//
	//Info(
	//	"started_in", time.Millisecond*20,
	//	"enabled", true,
	//	"started",
	//)
	//
	//jsonBytes, _ := (&testMessage{
	//	Id:   "150",
	//	Name: "MNO",
	//}).MarshalJSON()

	//Info(
	//	os.ErrClosed,
	//	&testMessage{
	//		Id:   "100",
	//		Name: "XYZ",
	//	},
	//	"started_in", time.Millisecond*20,
	//	"enabled", true,
	//	"payload", &testMessage{
	//		Id:   "101",
	//		Name: "ABC",
	//	},
	//	JSON(jsonBytes),
	//	jsonBytes,
	//	"event", JSON(jsonBytes),
	//	"started",
	//)
}

func stub() {
	getg()
}

func Hi() {
	g1 := getg()
	_ = g1
	//funk := getCallerPC()
	//printFileAndLine(funk)
	doLogger()

	func() {
		//doLogger()
		//funk := getCallerPC()
		//printFileAndLine(funk)
	}()
}

func doLogger() {
	gp := getg()
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
	printFileAndLine(pc)
}

func Hi2() {
	g1 := getg()
	_ = g1
	doLogger()
	func() {
		doLogger()

		func() {
			doLogger()

			func() {
				doLogger()
			}()
		}()
	}()
}

func BenchmarkInfo(b *testing.B) {
	b.Run("slog", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			Slog(TestMessage,
				String("string", TestString),
				Int("status", TestInt),
				Duration("duration", TestDuration),
				Time("time", TestTime),
				Any("error", TestError),
			)
		}
	})
	b.Run("kirana", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			Log(
				"string", TestString,
				"status", TestInt,
				"duration", TestDuration,
				//"time", TestTime,
				"error", TestError,
				TestMessage,
			)
		}
	})
}

func printFileAndLine(pc uintptr) {
	f := runtime.FuncForPC(pc)
	fi, line := f.FileLine(pc)
	println(f.Name(), fi, line)
}

func getPC() {
	pcbuf := make([]uintptr, 1)
	runtime.Callers(3, pcbuf)
	frames := runtime.CallersFrames(pcbuf)
	for {
		frame, ok := frames.Next()

		fmt.Println(frame.Function, frame.File, frame.Line, frame.Entry, frame.Func.Entry())

		if !ok {
			break
		}

		//break
	}
}

type disabledHandler struct{}

func (disabledHandler) Enabled(context.Context, slog.Level) bool { return true }
func (disabledHandler) Handle(ctx context.Context, r slog.Record) error {
	//panic("should not be called")
	return nil
}

func (disabledHandler) HandleRaw(r slog.Record, attr ...slog.Attr) error {
	return nil
}

func (disabledHandler) WithAttrs([]slog.Attr) slog.Handler {
	panic("disabledHandler: With unimplemented")
}

func (disabledHandler) WithGroup(string) slog.Handler {
	panic("disabledHandler: WithGroup unimplemented")
}
