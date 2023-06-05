package logger

import (
	"testing"
	"time"
)

func TestGID(t *testing.T) {
	for i := 0; i < 10; i++ {
		go func() {
			t.Log("Gid", getg().goid, "Pid", getg().m.p.id)
		}()
	}
	t.Log("Gid", getg().goid, "Pid", getg().m.p.id)
	time.Sleep(time.Second * 1)
}

func BenchmarkGetg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getg()
	}
}
