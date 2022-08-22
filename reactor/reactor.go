package reactor

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytedance/gopkg/util/gopool"
	logger "github.com/moontrade/log"
	"github.com/moontrade/wormhole/pkg/counter"
	"github.com/moontrade/wormhole/pkg/hashmap"
	"github.com/moontrade/wormhole/pkg/mpsc"
	"github.com/moontrade/wormhole/pkg/pmath"
	"github.com/moontrade/wormhole/pkg/runtimex"
	"github.com/moontrade/wormhole/pkg/timex"
	"github.com/moontrade/wormhole/pkg/util"
	"github.com/panjf2000/ants"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Stats struct {
	ticks              counter.Counter
	ticksDur           counter.TimeCounter
	ticksDurMin        counter.Counter
	ticksDurMax        counter.Counter
	spawns             counter.Counter
	spawnsDur          counter.TimeCounter
	wakes              counter.Counter
	wakesDur           counter.TimeCounter
	wakeLists          counter.Counter
	wakesListsDur      counter.TimeCounter
	wakeListsWakes     counter.Counter
	wakesListsWakesDur counter.TimeCounter
	wakeListsInvokes   counter.Counter
	invokes            counter.Counter
	invokesDur         counter.TimeCounter
	flushes            counter.Counter
	flushesDur         counter.TimeCounter
	skew               counter.Counter
	skewDur            counter.TimeCounter
	droppedDur         counter.TimeCounter
	level1Ticks        counter.Counter
	level1TicksDur     counter.TimeCounter
	level1TicksDurMin  counter.Counter
	level1TicksDurMax  counter.Counter
	level2Ticks        counter.Counter
	level2TicksDur     counter.TimeCounter
	level2TicksDurMin  counter.Counter
	level2TicksDurMax  counter.Counter
	level3Ticks        counter.Counter
	level3TicksDur     counter.TimeCounter
	level3TicksDurMin  counter.Counter
	level3TicksDurMax  counter.Counter
	pidSwitches        counter.Counter
}

var (
	ErrQueueFull = errors.New("queue full")
	ErrStop      = errors.New("stop")
)

const (
	DefaultInvokeQueueSize = 1024 * 1
	DefaultWakeQueueSize   = 1024 * 1
	DefaultSpawnQueueSize  = 1024 * 1
)

type Tick struct {
	Time      int64
	Tick      int64
	Dur       time.Duration
	Precision time.Duration
}

type Config struct {
	Name         string
	Level1Wheel  Wheel
	Level2Wheel  Wheel
	Level3Wheel  Wheel
	InvokeQSize  int
	WakeQSize    int
	SpawnQSize   int
	LockOSThread bool
}

// Reactor runs all tasks on a single goroutine. It has an optimized timing mechanism
// with a fixed tickDur duration and a fixed interval duration. The interval is broken down
// into slots for better resource allocation. For example, a tickDur duration of 4ms with 5
// slots gives an interval duration of 20ms with each 4ms handling ~20% of the load. The
// timing is constantly adjusting to ensure the tickDur duration is accurate from the start
// adjusting for CPU Time.
// In addition, there is a lock-free MPSC queue that accepts invokes to run immediately
// without having to wait for a Tick.
type Reactor struct {
	Stats
	id             int
	now            int64
	size           counter.Counter
	currentTick    counter.Counter
	idCounter      counter.Counter
	state          int64
	wakeQ          *mpsc.Bounded[Task]
	wakeListQ      *mpsc.Bounded[WakeList]
	spawnQ         *mpsc.Bounded[Task]
	invokeQ        *mpsc.Bounded[func()]
	timer          chan Tick
	tickWheel      Wheel
	level2Wheel    Wheel
	level3Wheel    Wheel
	tickDur        time.Duration
	ticksPerLevel2 int64
	ticksPerLevel3 int64
	tasks          *hashmap.Sync[int64, *Task]
	workerPool     gopool.Pool
	ctx            context.Context
	cancel         context.CancelFunc
	tickCount      counter.Counter
	nextTick       counter.Counter
	pid            int
	lockOSThread   bool
	wg             sync.WaitGroup
}

func NewReactor(config Config) (*Reactor, error) {
	if config.Name == "" {
		config.Name = "loop"
	}
	if config.InvokeQSize <= 4 {
		config.InvokeQSize = DefaultInvokeQueueSize
	}
	if config.WakeQSize <= 4 {
		config.WakeQSize = DefaultWakeQueueSize
	}
	if config.SpawnQSize <= 4 {
		config.SpawnQSize = DefaultSpawnQueueSize
	}
	config.InvokeQSize = pmath.CeilToPowerOf2(config.InvokeQSize)
	config.WakeQSize = pmath.CeilToPowerOf2(config.WakeQSize)
	config.SpawnQSize = pmath.CeilToPowerOf2(config.SpawnQSize)
	if len(config.Level1Wheel.durations) == 0 {
		config.Level1Wheel = NewWheel(Millis250)
	}
	if len(config.Level2Wheel.durations) == 0 {
		config.Level2Wheel = NewWheel(Seconds)
	}
	if len(config.Level3Wheel.durations) == 0 {
		config.Level3Wheel = NewWheel(Minutes)
	}
	if config.Level2Wheel.tickDur%config.Level1Wheel.tickDur != 0 {
		return nil, fmt.Errorf("seconds Tick not evenly divisible by millisecond Tick: %s mod %s = %s",
			config.Level2Wheel.tickDur, config.Level1Wheel.tickDur, config.Level2Wheel.tickDur%config.Level1Wheel.tickDur)
	}
	if config.Level3Wheel.tickDur%config.Level1Wheel.tickDur != 0 {
		return nil, fmt.Errorf("minutes Tick not evenly divisible by millisecond Tick: %s mod %s = %s",
			config.Level3Wheel.tickDur, config.Level1Wheel.tickDur, config.Level3Wheel.tickDur%config.Level1Wheel.tickDur)
	}
	ctx, cancel := context.WithCancel(context.Background())
	w := &Reactor{
		tickDur:        config.Level1Wheel.tickDur,
		tickWheel:      config.Level1Wheel,
		level2Wheel:    config.Level2Wheel,
		ticksPerLevel2: int64(config.Level2Wheel.tickDur / config.Level1Wheel.tickDur),
		level3Wheel:    config.Level3Wheel,
		ticksPerLevel3: int64(config.Level3Wheel.tickDur / config.Level1Wheel.tickDur),
		tasks:          hashmap.NewSync[int64, *Task](8, 1024, hashmap.HashInt64),
		wakeQ:          mpsc.NewBounded[Task](int64(config.WakeQSize), nil),
		wakeListQ:      mpsc.NewBounded[WakeList](int64(config.WakeQSize), nil),
		spawnQ:         mpsc.NewBounded[Task](int64(config.SpawnQSize), nil),
		invokeQ:        mpsc.NewBounded[func()](int64(config.InvokeQSize), nil),
		timer:          make(chan Tick, 1),
		workerPool:     gopool.NewPool(config.Name, 10000, nil),
		ctx:            ctx,
		cancel:         cancel,
		lockOSThread:   config.LockOSThread,
	}
	w.id = reactors.AppendIndex(w)
	return w, nil
}

func (r *Reactor) ID() int { return r.id }

func (r *Reactor) Now() int64 { return r.now }

func (r *Reactor) SnapshotStats() Stats {
	return r.Stats
}

func (r *Reactor) Start() {
	if !atomic.CompareAndSwapInt64(&r.state, 0, 1) {
		return
	}
	go r.run()
}

func (r *Reactor) Duration(ticks int64) time.Duration {
	return r.tickDur * time.Duration(ticks)
}

func (r *Reactor) Ticks(duration time.Duration) int64 {
	if duration <= 0 {
		return 0
	}
	return int64(duration) / int64(r.tickDur)
}

func (r *Reactor) Wake(task *Task) error {
	if task == nil {
		return errors.New("task is nil")
	}
	reactor := task.reactor
	if reactor == nil {
		return errors.New("task is not scheduled")
	}
	if reactor != r {
		return reactor.Wake(task)
	}
	r.wakeQ.Push(task)
	return nil
}

func (r *Reactor) WakeAfter(task *Task, after time.Duration) error {
	if task == nil {
		return errors.New("task is nil")
	}
	if after <= 0 {
		return r.Wake(task)
	}
	reactor := task.reactor
	if reactor == nil {
		return errors.New("task is not scheduled")
	}
	task.wakeAfter = after
	if reactor != r {
		return reactor.Wake(task)
	}
	if !r.wakeQ.Push(task) {
		return ErrQueueFull
	} else {
		return nil
	}
}

func (r *Reactor) WakeList(list *WakeList) error {
	if list == nil {
		return errors.New("nil slots")
	}
	if list.Len() == 0 {
		return nil
	}
	if list.reactor != r {
		return list.reactor.WakeList(list)
	}
	if !r.wakeListQ.Push(list) {
		return ErrQueueFull
	} else {
		return nil
	}
}

func (r *Reactor) Invoke(fn func()) bool {
	if fn == nil {
		return false
	}
	return r.invokeQ.PushUnsafe(runtimex.FuncToPointer(fn))
}

func (r *Reactor) InvokeRef(fn *func()) bool {
	if fn == nil {
		return false
	}
	return r.invokeQ.PushUnsafe(runtimex.FuncToPointer(*fn))
}

func (r *Reactor) InvokeBlocking(fn func()) bool {
	if fn == nil {
		return false
	}
	return InvokeBlocking(fn)
}

func (r *Reactor) Spawn(future Future) (*Task, error) {
	if future == nil {
		return nil, errors.New("nil future")
	}
	task := taskPool.Get()
	task.init(r.idCounter.Incr(), r, future)
	if provider, ok := future.(FutureTask); ok {
		provider.SetTask(task)
	}
	if !r.spawnQ.Push(task) {
		return nil, ErrQueueFull
	}
	return task, nil
}

func (r *Reactor) SpawnInterval(future Future, interval time.Duration) (*Task, error) {
	if future == nil {
		return nil, errors.New("nil future")
	}
	if interval < 0 {
		interval = 0
	}
	task := taskPool.Get()
	task.init(r.idCounter.Incr(), r, future)
	task.interval = interval
	if provider, ok := future.(FutureTask); ok {
		provider.SetTask(task)
	}
	if !r.spawnQ.Push(task) {
		return nil, ErrQueueFull
	}
	return task, nil
}

func (r *Reactor) SpawnWorker(fn func()) error {
	return ants.Submit(fn)
}

func (r *Reactor) run() {
	defer func() {
		e := recover()
		if e != nil {
			logger.Error(util.PanicToError(e))
		}
	}()

	r.pid = runtimex.Pid()
	var (
		invokeQ       = r.invokeQ
		invokeQWake   = invokeQ.Wake()
		wakeQ         = r.wakeQ
		wakeQWake     = wakeQ.Wake()
		wakeListQ     = r.wakeListQ
		wakeListQWake = wakeListQ.Wake()
		spawnQ        = r.spawnQ
		spawnQWake    = spawnQ.Wake()
		//now         = timex.NanoTime()
		pid     = r.pid
		nextPid = pid
	)

	checkLoopPid := func() {
		nextPid = runtimex.Pid()
		if nextPid != pid {
			//logger.Debug("loop pid changed from %d to %d between iterations", pid, nextPid)
		}
		pid = nextPid
	}

	checkPid := func() {
		nextPid = runtimex.Pid()
		if nextPid != pid {
			//logger.Warn("loop pid changed from %d to %d during loop execution", pid, nextPid)
		}
		pid = nextPid
	}

	r.nextTick.Store(timex.NanoTime() + int64(r.tickDur))

	go func() {
		interval := int64(r.tickDur)
	Loop:
		for {
			begin := timex.NanoTime()
			time.Sleep(time.Duration(r.nextTick.Load() - begin))

			t := r.nextTick.Load()
			tick := r.currentTick.Incr()
			r.nextTick.Add(interval)

			select {
			case r.timer <- Tick{
				Time: t,
				Tick: tick,
			}:
			// caught up
			case <-r.ctx.Done():
				break Loop
			default:
				// behind
			}
		}
	}()

	onSpawn := func(task *Task) {
		r.pollStart(r.now, task)
	}

	onWake := func(task *Task) {
		r.pollWake(r.now, task)
	}

	onWakeList := func(list *WakeList) {
		r.pollWakeList(r.now, list)
	}

	onFn := func(task func()) {
		r.invoke(task)
	}

	flushSpawnQueue := func() {
		if spawnQ.IsEmpty() {
			return
		}
		r.now = timex.NanoTime()
		count := spawnQ.PopMany(math.MaxUint32, onSpawn)
		end := timex.NanoTime()
		r.spawns.Add(int64(count))
		r.spawnsDur.Add(end - r.now)
	}

	flushWakeQueue := func() {
		if wakeQ.IsEmpty() {
			return
		}
		r.now = timex.NanoTime()
		count := wakeQ.PopMany(math.MaxUint32, onWake)
		end := timex.NanoTime()
		r.wakes.Add(int64(count))
		r.wakesDur.Add(end - r.now)
	}

	flushWakeListsQueue := func() {
		if wakeListQ.IsEmpty() {
			return
		}
		r.now = timex.NanoTime()
		count := wakeListQ.PopMany(math.MaxUint32, onWakeList)
		end := timex.NanoTime()
		r.wakeLists.Add(int64(count))
		r.wakesListsDur.Add(end - r.now)
	}

	flushInvokeQueue := func() {
		if invokeQ.IsEmpty() {
			return
		}
		now := timex.NanoTime()
		count := invokeQ.PopManyDeref(math.MaxUint32, onFn)
		end := timex.NanoTime()
		r.invokes.Add(int64(count))
		r.invokesDur.Add(end - now)
	}

	flushQueues := func() {
		flushInvokeQueue()
		flushWakeQueue()
		flushWakeListsQueue()
		flushSpawnQueue()
	}

	onTick := func(msg Tick) {
		interval := int64(r.tickDur)
		start := timex.NanoTime()
		begin := start
		r.tick(msg, begin)
		end := timex.NanoTime()
		elapsed := end - begin

		checkPid()

		// Stats
		r.ticks.Incr()
		r.ticksDur.Add(elapsed)
		if r.ticksDurMin == 0 || r.ticksDurMin.Load() > elapsed {
			r.ticksDurMin.Store(elapsed)
		}
		if r.ticksDurMax.Load() < elapsed {
			r.ticksDurMax.Store(elapsed)
		}

		begin = end
		flushQueues()
		end = timex.NanoTime()
		r.flushesDur.Add(end - begin)
		elapsed = end - start
		r.flushes.Add(1)

		if elapsed > interval {
			r.skew.Incr()
			r.skewDur.Add(elapsed)

			r.rebalance()
		}
	}

	if r.lockOSThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	lastTick := int64(0)
Loop:
	for {
		select {
		case msg := <-r.timer:
			checkLoopPid()
			if lastTick < msg.Tick-1 {
				r.catchup(lastTick, msg.Tick)
			}
			lastTick = msg.Tick
			onTick(msg)
			checkPid()

		case <-wakeListQWake:
			checkLoopPid()
			flushWakeListsQueue()
			checkPid()

		case <-wakeQWake:
			checkLoopPid()
			flushWakeQueue()
			checkPid()

		case <-spawnQWake:
			checkLoopPid()
			flushSpawnQueue()
			checkPid()

		case <-invokeQWake:
			checkLoopPid()
			flushInvokeQueue()
			checkPid()

		case <-r.ctx.Done():
			break Loop
		}
	}
}

func (r *Reactor) catchup(lastTick, currentTick int64) {
	logger.Warn("skew detected of %d ticks", currentTick-1-lastTick)
	logger.Warn("catching up...")

	for nextTick := lastTick + 1; nextTick <= currentTick; nextTick++ {
		now := timex.NanoTime()
		r.tick(Tick{
			Time: now,
			Tick: nextTick,
		}, now)
	}
}

func (r *Reactor) rebalance() {
	//logger.Warn("rebalancing...")
	//logger.Warn("rebalanced")
}

func (r *Reactor) tick(msg Tick, now int64) {
	r.tickWheel.tick(now, r.onTick)
	if msg.Tick%r.ticksPerLevel2 == 0 {
		//logger.Debug("level 2 wheel Tick")
		r.level2Wheel.tick(now, r.onTick)
	}
	if msg.Tick%r.ticksPerLevel3 == 0 {
		//logger.Debug("level 3 wheel Tick")
		r.level3Wheel.tick(now, r.onTick)
	}
}

func (r *Reactor) onTick(now int64, list *taskSwapList, slot *taskSwapSlot, task *Task) bool {
	if slot.wake {
		r.pollWake(now, task)
		return false
	}
	return r.pollInterval(now, list, task)
}

func (r *Reactor) schedule(task *Task, delay time.Duration, wake bool) {
	if delay < r.tickWheel.maxDur && r.tickWheel.tickDur > 0 {
		r.tickWheel.schedule(task, delay, wake)
	} else if delay < r.level2Wheel.maxDur && r.level2Wheel.tickDur > 0 {
		r.level2Wheel.schedule(task, delay, wake)
	} else if delay < r.level3Wheel.maxDur && r.level3Wheel.tickDur > 0 {
		r.level3Wheel.schedule(task, delay, wake)
	}
}

func (r *Reactor) stopTask(time int64, task *Task) {
	defer func() {
		task.remove()
		if e := recover(); e != nil {
			err := util.PanicToError(e)
			logger.Error(err, "Reactor.invoke panic")
		}
	}()
	_, ok := r.tasks.Delete(task.id)
	if !ok {
		return
	}
	task.stop = true
	task.clearSlots()

	if pc, ok := task.future.(PollClose); ok {
		err := pc.PollClose(CloseEvent{
			Task:   task,
			Time:   time,
			Reason: nil,
		})
		if err != nil {
			logger.Warn(err)
		}
	}
}

func (r *Reactor) invoke(fn func()) {
	defer func() {
		if e := recover(); e != nil {
			err := util.PanicToError(e)
			logger.Error(err, "Reactor.invoke panic")
		}
	}()
	if fn != nil {
		fn()
	}
}

func (r *Reactor) pollStart(now int64, task *Task) {
	defer func() {
		if e := recover(); e != nil {
			err := util.PanicToError(e)
			logger.Error(err, "Reactor.pollInvoke panic")
		}
	}()
	task.started = now
	err := task.future.Poll(Context{
		Task:   task,
		Time:   now,
		Reason: ReasonStart,
	})

	if err != nil {
		if err == ErrStop {
			task.stop = true
		} else {
			logger.Warn(err)
		}
	}

	if task.stop {
		r.stopTask(now, task)
		return
	}

	r.tasks.Put(task.id, task)

	if task.wakeAfter > 0 {
		r.schedule(task, task.wakeAfter, true)
	}

	if task.interval > 0 {
		r.schedule(task, task.interval, false)
	}
}

func (r *Reactor) pollWakeList(now int64, list *WakeList) {
	count := list.wakes.wake(now, r.onTaskSlotWake)
	r.wakeListsWakes.Add(count)

	count = list.funcs.wake(now, r.onFuncSlotWake)
	r.wakeListsInvokes.Add(count)
}

func (r *Reactor) onTaskSlotWake(slot *TaskSlot) {
	t := slot.task
	if t == nil || !slot.wake {
		return
	}
	r.pollWake(r.now, t)
}

func (r *Reactor) onFuncSlotWake(slot *FuncSlot) {
	fn := slot.Value
	if fn == nil {
		return
	}
	r.invoke(fn)
}

func (r *Reactor) pollWake(now int64, task *Task) {
	if task.stop {
		return
	}
	defer func() {
		if e := recover(); e != nil {
			err := util.PanicToError(e)
			logger.Error(err, "Reactor.pollInvoke panic")
		}
	}()

	interval := task.interval
	wakeAfter := task.wakeAfter
	if wakeAfter > 0 {
		task.wakeAfter = 0
		r.schedule(task, wakeAfter, true)
		return
	}

	task.wakes++
	err := task.future.Poll(Context{
		Task:   task,
		Time:   now,
		Reason: ReasonWake,
	})

	if err != nil {
		if err == ErrStop {
			task.stop = true
		} else {
			logger.Warn(err)
		}
	}

	if task.stop {
		r.stopTask(now, task)
		return
	}

	newWakeAfter := task.wakeAfter
	if newWakeAfter != wakeAfter {
		if newWakeAfter > 0 {
			r.schedule(task, newWakeAfter, true)
		}
	}

	nextInterval := task.interval
	// Interval change requested?
	if nextInterval != interval {
		if nextInterval <= 0 {
			task.interval = 0
		}
	}
}

func (r *Reactor) pollInterval(now int64, list *taskSwapList, task *Task) bool {
	if task.stop || task.interval == 0 || list.dur != task.interval {
		// remove
		return false
	}

	defer func() {
		if e := recover(); e != nil {
			err := util.PanicToError(e)
			logger.Error(err, "Reactor.pollInterval panic")
		}
	}()

	interval := task.interval
	wakeAfter := task.wakeAfter

	task.intervals++
	err := task.future.Poll(Context{
		Task:     task,
		Time:     now,
		Interval: interval,
		Reason:   ReasonInterval,
	})

	if err != nil {
		if err == ErrStop {
			task.stop = true
		} else {
			logger.Warn(err)
		}
	}

	if task.stop {
		r.stopTask(now, task)
		// remove
		return false
	}

	newWakeAfter := task.wakeAfter
	if newWakeAfter != wakeAfter {
		if newWakeAfter > 0 {
			r.schedule(task, newWakeAfter, true)
		}
	}

	// Interval change requested?
	nextInterval := task.interval
	if nextInterval != interval {
		if nextInterval <= 0 {
			task.interval = 0
			return false
		}
		// Schedule new interval
		r.schedule(task, nextInterval, false)
		// Remove from this taskSwapList
		return false
	}
	return true
}

func (r *Reactor) Print() {
	avg := time.Duration(r.ticksDur.Load()) / time.Duration(r.currentTick.Load())
	fmt.Println("Size			", r.size.Load())
	fmt.Println("PID				", r.pid)
	fmt.Println("PID Switches	", r.pidSwitches.Load())
	//fmt.Println("Capacity		", r.cap)
	fmt.Println("Ticks			", r.currentTick.Load())
	//fmt.Println("Ticks Dur 		", Time.Duration(r.ticksDur.Load()))
	fmt.Println("Tick Avg Dur 	", time.Duration(r.ticksDur.Load())/time.Duration(r.currentTick.Load()))
	//fmt.Println("Skew			", r.skew.Load())
	//fmt.Println("Skew Dur		", Time.Duration(r.skewDur.Load()))
	//fmt.Println("Dropped Dur	", r.droppedDur.Load())
	////fmt.Println("Ticks Dur 		", Time.Duration(r.ticksDur.Load()))
	//fmt.Println("Jobs			", r.invokes.Load())
	////fmt.Println("Ticks Dur 		", Time.Duration(r.ticksDur.Load()))
	//fmt.Println("Jobs Avg Dur 	", Time.Duration(r.invokesDur.Load())/Time.Duration(r.invokes.Load()))
	//fmt.Println("Interval 		", r.tickDur)
	fmt.Println("Tick CPU 		", float64(avg)/float64(r.tickDur))
	fmt.Println("Min 	 		", time.Duration(r.ticksDurMin.Load()))
	fmt.Println("Max 	 		", time.Duration(r.ticksDurMax.Load()))
	//for i, slots := range r.tickWheel {
	//	fmt.Println("Ring: ", i, " Size: ", slots.activeSize)
	//}
}
