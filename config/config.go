package config

import (
	"runtime"
	"time"
)

var (
	NumLoops        = runtime.GOMAXPROCS(0)
	BlockingQSize   = 16384
	LoopInvokeQSize = 16384
	LoopWakeQSize   = 16384
	LoopSpawnQSize  = 16384
	TickCadence     = []time.Duration{
		time.Millisecond * 250,
		time.Millisecond * 500,
		time.Millisecond * 750,
	}
)
