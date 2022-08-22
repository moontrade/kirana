package reactor

import (
	"github.com/moontrade/wormhole/pkg/counter"
	"github.com/moontrade/wormhole/pkg/cow"
	"github.com/moontrade/wormhole/pkg/runtimex"
	"runtime"
	"strconv"
)

var ()

var (
	idCounter counter.Counter
	ticker    *Ticker
	blocking  *BlockingPool
	reactors  cow.Slice[*Reactor]
	loops     cow.Slice[*Reactor]
)

func init() {
	//runtime_registerPoolCleanup(gc)
}

func NumReactors() int { return reactors.Len() }

func NextEventLoop() *Reactor {
	loops := loops.Snapshot()
	if len(loops) == 0 {
		return nil
	}
	if len(loops) == 1 {
		return loops[0]
	}
	return loops[runtimex.Pid()%len(loops)]
}

func Init(
	numLoops int,
	tick Cadence,
	queueSize int,
	blockingQueueSize int,
) {
	if loops.Len() > 0 {
		return
	}
	if numLoops == 0 {
		numLoops = runtime.GOMAXPROCS(0)
	}
	if queueSize < 1024 {
		queueSize = 1024
	}
	if blockingQueueSize < 1024 {
		blockingQueueSize = 1024
	}
	blocking = NewBlockingPool(0, blockingQueueSize)
	ticker = StartTicker(tick.Tick())

	l := make([]*Reactor, numLoops)
	for i := 0; i < numLoops; i++ {
		loop, err := NewReactor(Config{
			Name:         "ev-" + strconv.Itoa(i),
			Level1Wheel:  NewWheel(tick),
			Level2Wheel:  NewWheel(Seconds),
			Level3Wheel:  NewWheel(Minutes),
			InvokeQSize:  queueSize,
			WakeQSize:    queueSize,
			SpawnQSize:   queueSize,
			LockOSThread: true,
		})
		if err != nil {
			panic(err)
		}
		l[i] = loop
		loop.Start()
	}
	loops.ReplaceWith(l)
}
