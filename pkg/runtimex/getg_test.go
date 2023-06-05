package runtimex

import (
	"github.com/moontrade/unsafe/cgo"
	"testing"
	"time"
)

func TestGID(t *testing.T) {
	for i := 0; i < 10; i++ {
		go func() {
			t.Log("Gid", GoroutineID(), "Pid", ProcessorID())
		}()
	}
	t.Log("Gid", GoroutineID(), "Pid", ProcessorID())
	time.Sleep(time.Second * 2)
}

func BenchmarkGoroutineID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GoroutineID()
	}
}

func BenchmarkMachineID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MachineID()
	}
}

func BenchmarkProcessorID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProcessorID()
	}
}

func BenchmarkGIDPID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GIDPID()
	}
}

func TestTrampoline(t *testing.T) {
	cgo.NonBlocking((*byte)(funcToPointer(trampoline)), 0, 0)
}

func trampoline(a0, a1 uintptr) {
	println("trampoline")
}
