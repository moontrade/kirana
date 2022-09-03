package reactor

import (
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/cow"
	"github.com/moontrade/kirana/pkg/runtimex"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var ()

var (
	idCounter counter.Counter
	ticker    *Ticker
	blocking  *BlockingPool
	reactors  cow.Slice[*Reactor]
	loops     cow.Slice[*Reactor]
	mu        sync.Mutex
)

func init() {
	//runtime_registerPoolCleanup(gc)
}

func initTicker(tickDur time.Duration) *Ticker {
	mu.Lock()
	defer mu.Unlock()
	if ticker != nil {
		return ticker
	}
	ticker = StartTicker(tickDur)
	return ticker
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
			LockOSThread: false,
		})
		if err != nil {
			panic(err)
		}
		l[i] = loop
		loop.Start()
	}
	loops.ReplaceWith(l)
}
