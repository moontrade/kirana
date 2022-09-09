package wasmtimex

import (
	"testing"
	"time"
)

func TestEpochThread(t *testing.T) {
	engine := NewEngine()
	engine1 := NewEngine()
	defer engine.Delete()
	go func() {
		StartEpochThread(time.Microsecond*1000, engine)
		engine.Delete()
		StopEpochThread()
		engine = NewEngine()
		StartEpochThreadMultiple(time.Microsecond*1000, engine, engine1)
	}()
	time.Sleep(time.Hour)
}
