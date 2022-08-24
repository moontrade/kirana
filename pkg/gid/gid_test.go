package gid

import (
	"testing"
	"time"
)

func TestGID(t *testing.T) {
	for i := 0; i < 10; i++ {
		go func() {
			t.Log("Gid", GID(), "Pid", PID())
		}()
	}
	t.Log("Gid", GID(), "Pid", PID())
	time.Sleep(time.Second * 2)
}

func BenchmarkGID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GID()
	}
}

func BenchmarkPID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PID()
	}
}

func BenchmarkGIDPID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GIDPID()
	}
}
