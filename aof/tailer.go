package aof

import (
	"errors"
	"github.com/moontrade/kirana/pkg/atomicx"
	"github.com/moontrade/kirana/pkg/util"
	"github.com/moontrade/kirana/reactor"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type TailerState int32

func (t *TailerState) Load() TailerState {
	return TailerState(atomic.LoadInt32((*int32)(t)))
}
func (t *TailerState) store(state TailerState) {
	atomic.StoreInt32((*int32)(t), int32(state))
}
func (t *TailerState) xchg(state TailerState) TailerState {
	return TailerState(atomicx.Xchgint32((*int32)(t), int32(state)))
}
func (t *TailerState) swap(state TailerState) TailerState {
	return TailerState(atomic.SwapInt32((*int32)(t), int32(state)))
}
func (t *TailerState) cas(old, new TailerState) bool {
	return atomicx.Casint32((*int32)(t), int32(old), int32(new))
}
func (t *TailerState) compareAndSwap(old, new TailerState) bool {
	return atomic.CompareAndSwapInt32((*int32)(t), int32(old), int32(new))
}

const (
	TailerStart   TailerState = 1 // Tailer was recently spawned and waiting for first start Dequeue
	TailerReading TailerState = 2 // Tailer is currently reading and is not at the tail yet
	TailerTail    TailerState = 3 // Tailer has started and is now at the tail
	TailerEOF     TailerState = 4 // Tailer parent AOF is finished the tailer goes from tail to Checkpoint state
	TailerClosing TailerState = 5 // Tailer is waiting for next Dequeue to close
	TailerClosed  TailerState = 6 // Tailer is now safe to delete
)

type ReadEvent struct {
	Time       int64
	Tailer     *Tailer
	Begin, End int64
	Tail       []byte
	contents   []byte
	FileState  FileState
	EOF        bool
	Reason     reactor.PollReason
}

func (re *ReadEvent) CheckGID() bool {
	return re.Tailer.CheckGID()
}

// Contents of the entire AOF
func (re *ReadEvent) Contents() []byte {
	return re.contents
}

type ClosedEvent struct {
	Tailer *Tailer
	Reason any
}

type Consumer interface {
	PollRead(event ReadEvent) (int64, error)

	PollReadClosed(reason error)
}

type Tailer struct {
	reactor.TaskProvider
	a  *AOF
	i  int64
	s  TailerState
	c  Consumer
	mu sync.Mutex
}

func (t *Tailer) State() TailerState {
	return t.s.Load()
}

func (t *Tailer) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	switch t.s.xchg(TailerClosing) {
	case TailerClosing:
	default:
		_ = t.Task().Wake()
	}
	return nil
}

func (t *Tailer) pushClosed(err error) {
	defer func() {
		t.s.store(TailerClosed)
		e := recover()
		if e != nil {
			err = util.PanicToError(e)
			//logger.Error(err, "Consumer.PollReadClosed panic")
		}
	}()
	var (
		c = t.c
	)
	if c != nil {
		c.PollReadClosed(err)
	}
}

func (t *Tailer) pushRead(event ReadEvent) (n int64, err error) {
	defer func() {
		e := recover()
		if e != nil {
			err = util.PanicToError(e)
			//logger.Error(err, "Tailer.PollRead panic")
		}
	}()
	var (
		c = t.c
	)
	if c != nil {
		n, err = c.PollRead(event)
	}
	return
}

func (t *Tailer) Poll(ctx reactor.Context) error {
Begin:
	var (
		state = t.s.Load()
	)

	switch state {
	case TailerClosing:
		t.pushClosed(reactor.ErrStop)
		return reactor.ErrStop
	case TailerClosed:
		return reactor.ErrStop
	case TailerStart, TailerTail, TailerEOF:
	}

	var (
		fileState = t.a.state.load()
		size      = atomic.LoadInt64(&t.a.size)
		toState   = state
	)

	// File was forcefully closed
	if fileState >= FileStateClosing {
		t.s.store(TailerClosing)
		t.pushClosed(os.ErrClosed)
		return reactor.ErrStop
	}

	if len(t.a.data) == 0 {
		return nil
	}

	if t.i == size {
		if fileState != FileStateEOF {
			return nil
		}
	}

	n, err := t.pushRead(ReadEvent{
		Time:      ctx.Time,
		Tailer:    t,
		Begin:     t.i,
		End:       size,
		Tail:      t.a.data[t.i:size],
		contents:  t.a.data,
		FileState: fileState,
		EOF:       fileState == FileStateEOF,
		Reason:    ctx.Reason,
	})
	if err != nil {
		if err != os.ErrClosed && err != reactor.ErrStop {
			//logger.WarnErr(err)
			return err
		} else {
			t.s.store(TailerClosing)
			t.pushClosed(os.ErrClosed)
			return reactor.ErrStop
		}
	}

	if n < 0 {
		n = 0
	} else if n > size {
		n = size
	}

	if n == size {
		if fileState == FileStateEOF {
			toState = TailerEOF
		} else {
			toState = TailerTail
		}
	} else {
		toState = TailerReading
		ctx.WakeOnNextTick()
	}

	if state != toState {
		if !t.s.cas(state, toState) {
			goto Begin
		}
	}

	atomic.StoreInt64(&t.i, n)

	return nil
}

func (aof *AOF) Subscribe(
	c Consumer,
) (*Tailer, error) {
	return aof.SubscribeInterval(0, c)
}

func (aof *AOF) SubscribeOn(
	r *reactor.Reactor,
	c Consumer,
) (*Tailer, error) {
	return aof.SubscribeIntervalOn(r, 0, c)
}

func (aof *AOF) SubscribeInterval(
	interval time.Duration,
	c Consumer,
) (*Tailer, error) {
	if c == nil {
		return nil, errors.New("nil consumer")
	}
	var (
		tailer = &Tailer{
			a: aof,
			c: c,
		}
		err error
	)
	if interval <= 0 {
		_, err = aof.tailers.Spawn(tailer)
	} else {
		_, err = aof.tailers.SpawnInterval(tailer, interval)
	}
	if err != nil {
		return nil, err
	}
	return tailer, nil
}

func (aof *AOF) SubscribeIntervalOn(
	r *reactor.Reactor,
	interval time.Duration,
	c Consumer,
) (*Tailer, error) {
	if r == nil {
		return nil, errors.New("nil reactor")
	}
	if c == nil {
		return nil, errors.New("nil consumer")
	}
	var (
		tailer = &Tailer{
			a: aof,
			c: c,
		}
		err error
	)
	if interval <= 0 {
		_, err = aof.tailers.SpawnOn(r, tailer)
	} else {
		_, err = aof.tailers.SpawnIntervalOn(r, tailer, interval)
	}
	if err != nil {
		return nil, err
	}
	return tailer, nil
}
