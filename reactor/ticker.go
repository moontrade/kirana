package reactor

/*
#include <time.h>
void kirana_sleep(size_t arg0, size_t arg1) {
	//struct timespec ts;
    //ts.tv_sec = 0;
    //ts.tv_nsec = arg0;
	//nanosleep(&ts, &ts);
	//std::this_thread::sleep_for((std::chrono::nanoseconds)arg0);
}
*/
import "C"
import (
	"errors"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/cow"
	"github.com/moontrade/kirana/pkg/timex"
	"github.com/moontrade/kirana/pkg/util"
	logger "github.com/moontrade/log"
	"github.com/moontrade/unsafe/cgo"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Tick struct {
	Time      int64
	Tick      int64
	Dur       time.Duration
	Precision time.Duration
}

type Ticker struct {
	tick       time.Duration
	ticks      counter.Counter
	ticksDur   counter.TimeCounter
	tickDurMin counter.Counter
	tickDurMax counter.Counter
	skews      counter.Counter
	skewMax    counter.Counter
	notifyList cow.Slice[*TickListener]
	stop       int32
	mu         sync.Mutex
	wg         sync.WaitGroup
}

func StartTicker(duration time.Duration) *Ticker {
	if duration < time.Microsecond {
		duration = time.Microsecond
	}
	t := &Ticker{
		tick:       duration,
		notifyList: *cow.NewSlice[*TickListener](),
	}
	t.wg.Add(1)
	go t.run()
	return t
}

func (t *Ticker) Close() error {
	if !atomic.CompareAndSwapInt32(&t.stop, 0, 1) {
		return os.ErrClosed
	}
	return nil
}

func (t *Ticker) Register(duration time.Duration, owner interface{}, ch chan int64) (*TickListener, error) {
	ln, err := newTickListener(t, duration, owner, ch)
	if err != nil {
		return nil, err
	}
	t.notifyList.Append(ln)
	return ln, nil
}

func (t *Ticker) remove(tl *TickListener) {
	t.notifyList.Remove(func(elem *TickListener) bool {
		return elem == tl
	})
}

func (t *Ticker) run() {
	defer t.wg.Done()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var (
		//Tick    int64
		started = timex.NanoTime()
		begin   = started
		end     int64
		elapsed int64
		sleep   time.Duration
		tickDur = int64(t.tick)
		next    = begin + tickDur
		list    = &t.notifyList
		msg     = Tick{Dur: t.tick, Precision: t.tick}
	)
	notify := func(ln *TickListener) bool {
		ln.tick(msg)
		return true
	}
	for atomic.LoadInt32(&t.stop) == 0 {
		msg.Tick = int64(t.ticks)
		msg.Time = begin
		list.Iterate(notify)
		end = timex.NanoTime()
		elapsed = end - begin
		if t.tickDurMin == 0 || int64(t.tickDurMin) > elapsed {
			t.tickDurMin.Store(elapsed)
		}
		if int64(t.tickDurMax) < elapsed {
			t.tickDurMax.Store(elapsed)
		}
		t.ticksDur.Add(elapsed)
		t.ticks++
		// Skew?
		if end > next {
			t.skews.Incr()
			ticksBehind := (end - next) / tickDur
			if (end-next)%tickDur > 0 {
				ticksBehind++
			}
			if int64(t.skewMax) < ticksBehind {
				t.skewMax.Store(ticksBehind)
			}
			//timeBehind := end - next
			t.ticks.Add(ticksBehind)
			next += tickDur * ticksBehind
			sleep = time.Duration(next - end)
			next += tickDur

			//logger.Warn("behind", ticksBehind, "ticker is behind %d ticks %s time sleeping for %s", ticksBehind, time.Duration(timeBehind), sleep)

			if sleep > 0 {
				park(sleep)
			}
			//endTick := int64(t.ticks) + ticksBehind
			//for ; int64(t.ticks) <= endTick; t.ticks.Incr() {
			//	if atomic.LoadInt32(&t.stop) == 1 {
			//		return
			//	}
			//	next += tickDur
			//	begin = timex.NanoTime()
			//	slots.Iterate(notify)
			//	end = timex.NanoTime()
			//	elapsed = end - begin
			//	if t.tickDurMin == 0 || int64(t.tickDurMin) > elapsed {
			//		t.tickDurMin.Store(elapsed)
			//	}
			//	if int64(t.tickDurMax) < elapsed {
			//		t.tickDurMax.Store(elapsed)
			//	}
			//	t.ticksDur.Add(elapsed)
			//}
		} else {
			sleep = time.Duration(next - end)
			next += tickDur
			if sleep > 0 {
				park(sleep)
			}
		}

		begin = timex.NanoTime()
	}
}

func park(duration time.Duration) {
	//time.Sleep(duration)
	cgo.NonBlockingSleep(duration)
}

type TickListener struct {
	ticker        *Ticker
	owner         interface{}
	ch            chan int64
	chOwned       bool
	dur           time.Duration
	last          int64
	total         int64
	next          Tick
	notifySuccess counter.Counter
	notifyFails   counter.Counter
	notifyPanics  counter.Counter
	mu            sync.Mutex
}

func (tl *TickListener) Chan() <-chan int64 {
	return tl.ch
}

func (tl *TickListener) Close() error {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	if tl.ticker == nil {
		return os.ErrClosed
	}

	tl.ticker.remove(tl)
	if tl.chOwned {
		close(tl.ch)
	}
	tl.ch = nil
	tl.ticker = nil
	return nil
}

func newTickListener(
	ticker *Ticker,
	duration time.Duration,
	owner interface{},
	ch chan int64,
) (*TickListener, error) {
	if duration <= 0 {
		return nil, errors.New("duration must be positive")
	}
	chOwned := false
	if ch == nil {
		ch = make(chan int64, 1)
		chOwned = true
	}
	return &TickListener{
		ticker:  ticker,
		owner:   owner,
		ch:      ch,
		chOwned: chOwned,
		dur:     duration,
		next:    Tick{Dur: duration},
	}, nil
}

func (tl *TickListener) tick(tick Tick) {
	if tick.Dur <= 0 {
		return
	}
	tl.total += int64(tick.Dur)

	elapsed := tl.total - tl.last
	if elapsed < int64(tl.dur) {
		return
	}

	tl.next.Precision = tick.Dur
	if elapsed == int64(tl.dur) {
		tl.last = tl.total
		tl.next.Time = tick.Time
		tl.doNotify()
		tl.next.Tick++
	} else if elapsed > int64(tl.dur) {
		tl.next.Time = tick.Time
		count := elapsed / int64(tl.dur)
		for i := int64(0); i < count; i++ {
			tl.doNotify()
			tl.last += int64(tl.dur)
			tl.next.Tick++
		}
	}
}

func (tl *TickListener) doNotify() {
	defer func() {
		if e := recover(); e != nil {
			tl.notifyPanics++
			logger.WarnErr(util.PanicToError(e), "panic")
		}
	}()
	select {
	case tl.ch <- tl.next.Tick:
		tl.notifySuccess++
	default:
		tl.notifyFails++
	}
}
