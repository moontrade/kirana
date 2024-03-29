package reactor

import (
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/cow"
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/runtimex"
)

var (
	idCounter    counter.Counter
	ticker       *Ticker
	blocking     *BlockingPool
	reactors     cow.Slice[*Reactor]
	reactorsMask = uint32(0)
	mu           sync.Mutex
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

func NextReactor() *Reactor {
	loops := reactors.Snapshot()
	return loops[runtimex.Fastrand()&reactorsMask]
}

func Init(
	numLoops int,
	tick Cadence,
	queueSize int,
	blockingQueueSize int,
) {
	if reactors.Len() > 0 {
		return
	}
	if numLoops == 0 {
		numLoops = runtime.GOMAXPROCS(0)
	}
	numLoops = pmath.CeilToPowerOf2(numLoops)
	reactorsMask = uint32(numLoops - 1)
	if queueSize < 1024 {
		queueSize = 1024
	}
	if blockingQueueSize < 64 {
		blockingQueueSize = 64
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
	reactors.ReplaceWith(l)
}
