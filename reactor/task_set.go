package reactor

import (
	"errors"
	"github.com/moontrade/kirana/pkg/atomicx"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/spinlock"
	"os"
	"runtime"
	"time"
)

type FutureTask interface {
	Future

	Task() *Task

	SetTask(task *Task)
}

type TaskSet struct {
	slots               *WakeLists
	numReactors         counter.Counter
	numEntries          counter.Counter
	numTasks            counter.Counter
	numFuncs            counter.Counter
	onEmpty             func()
	now                 counter.Counter
	lastWakeLatency     counter.Counter
	lastSoftWakeLatency counter.Counter
	mu                  spinlock.Mutex
	closed              bool
}

func (tl *TaskSet) LastWakeLatency() int64 {
	return tl.lastWakeLatency.Load()
}

func (tl *TaskSet) LastSoftWakeLatency() int64 {
	return tl.lastSoftWakeLatency.Load()
}

func (tl *TaskSet) IsEmpty() bool { return tl.numEntries == 0 }

func (tl *TaskSet) NumReactors() int { return int(tl.numReactors) }

func (tl *TaskSet) NumEntries() int { return int(tl.numTasks) }

func (tl *TaskSet) Release() bool {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	if tl.numEntries.Load() > 0 {
		return false
	}
	//slots := tl.slots
	//tl.slots = nil
	//for _, slot := range slots.slots {
	//
	//}
	return true
}

func (tl *TaskSet) decrTasks() {
	tl.numTasks.Decr()
	if tl.numEntries.Decr() == 0 && tl.IsEmpty() && tl.onEmpty != nil {
		tl.onEmpty()
	}
}

func (tl *TaskSet) decrFuncs() {
	tl.numFuncs.Decr()
	if tl.numEntries.Decr() == 0 && tl.IsEmpty() && tl.onEmpty != nil {
		tl.onEmpty()
	}
}

func (tl *TaskSet) Close() error {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.closed = true
	return nil
}

func (tl *TaskSet) Stop() int64 {
	_ = tl.Close()
	return tl.numTasks.Load()
}

func (tl *TaskSet) Wake() error {
	slots := tl.slots.slots
	for i := 0; i < len(slots); i++ {
		for !slots[i].wake() {
			runtime.Gosched()
		}
	}
	return nil
}

func (tl *TaskSet) Spawn(future FutureTask) (*Task, error) {
	return tl.SpawnOn(NextEventLoop(), future)
}

func (tl *TaskSet) SpawnOn(reactor *Reactor, future FutureTask) (*Task, error) {
	if reactor == nil {
		return nil, errors.New("no event loops")
	}
	if future == nil {
		return nil, errors.New("nil future")
	}
	task := taskPool.Get()
	task.init(reactor.idCounter.Incr(), reactor, future)
	future.SetTask(task)
	var err error
	_, err = tl.Add(future)
	if err != nil {
		taskPool.Put(task)
		return nil, err
	}
	if !reactor.spawnQ.Push(task) {
		return nil, ErrQueueFull
	}
	return task, nil
}

func (tl *TaskSet) SpawnInterval(
	future FutureTask,
	interval time.Duration,
) (*Task, error) {
	return tl.SpawnIntervalOn(NextEventLoop(), future, interval)
}

func (tl *TaskSet) SpawnIntervalOn(
	reactor *Reactor,
	future FutureTask,
	interval time.Duration,
) (*Task, error) {
	if reactor == nil {
		return nil, errors.New("no event loops")
	}
	if future == nil {
		return nil, errors.New("nil future")
	}
	task := taskPool.Get()
	task.init(reactor.idCounter.Incr(), reactor, future)
	task.interval = interval
	future.SetTask(task)
	var err error
	_, err = tl.Add(future)
	if err != nil {
		taskPool.Put(task)
		return nil, err
	}
	if !reactor.spawnQ.Push(task) {
		return nil, ErrQueueFull
	}
	return task, nil
}

func (tl *TaskSet) Add(value FutureTask) (*TaskSlot, error) {
	task := value.Task()
	if task == nil {
		return nil, errors.New("nil task")
	}
	r := task.Reactor()
	if r == nil {
		return nil, errors.New("nil reactor")
	}
	reactorID := r.ID()
	tl.mu.Lock()
	if tl.closed {
		tl.mu.Unlock()
		return nil, os.ErrClosed
	}
	if tl.slots == nil {
		tl.slots = wakeListSlicePool.Get()
	}
	slot, ok := tl.slots.GetOrCreate(reactorID, tl)
	if !ok {
		return nil, errors.New("reactor id must be 0-65535")
	}
	if slot.reactor == nil {
		slot.reactor = r
		tl.numReactors.Incr()
	}
	tl.mu.Unlock()
	taskSlot, err := slot.addWake(value)
	if err != nil {
		return nil, err
	}
	task.linkSlot(taskSlot)
	return taskSlot, nil
}

func (tl *TaskSet) AddFunc(r *Reactor, value func()) (*FuncSlot, error) {
	if value == nil {
		return nil, errors.New("nil task")
	}
	if r == nil {
		r = NextEventLoop()
	}
	reactorID := r.ID()
	tl.mu.Lock()
	if tl.closed {
		tl.mu.Unlock()
		return nil, os.ErrClosed
	}
	if tl.slots == nil {
		tl.slots = wakeListSlicePool.Get()
	}
	slot, ok := tl.slots.GetOrCreate(reactorID, tl)
	if !ok {
		return nil, errors.New("reactor id must be 0-65535")
	}
	if slot.reactor == nil {
		tl.numReactors.Incr()
	}
	tl.mu.Unlock()
	funcSlot, err := slot.addFunc(value)
	if err != nil {
		return nil, err
	}
	return funcSlot, nil
}

type WakeList struct {
	owner     *TaskSet
	reactor   *Reactor
	wakes     TaskSwapSlice
	intervals TaskSwapSlice
	funcs     FuncSwapSlice
	size      int64
	version   int64
	closed    int64
	fn        func(slot *TaskSlot) bool
	running   int64
	isWaking  int64
	lastWake  int64
	mu        spinlock.Mutex
	runMu     spinlock.Mutex
}

func (w *WakeList) init(owner *TaskSet) {
	w.reactor = nil
	w.owner = owner
	w.closed = 0
	w.size = 0
	w.version = 0
	w.running = 0
	w.isWaking = 0
}

func (w *WakeList) onWake(now int64) int64 {
	more := atomicx.Loadint64(&w.isWaking)
	if more == 0 {
		return 0
	}
	atomicx.Xaddint64(&w.isWaking, -more)
	return more
	//atomic.StoreInt64(&w.isWaking, 0)
	//atomic.StoreInt64(&w.lastWake, now)
}

func (w *WakeList) wake() bool {
	wakes := atomicx.Xaddint64(&w.isWaking, 1)
	if wakes > 1 {
		return true
		//runtime.Gosched()
		//return true
		//if wakes > 2 {
		//	return true
		//}
	}
	r := w.reactor
	if r != nil {
		err := r.wakeList(w)
		if err != nil {
			//runtime.Gosched()
			err = r.wakeList(w)
			if err != nil {
				if wakes > 1 {
					runtime.Gosched()
					return true
				}
				return false
			}
		}
	}
	return true
}

func (w *WakeList) Reactor() *Reactor { return w.reactor }

func (w *WakeList) Len() int { return int(w.size) }

func (w *WakeList) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed > 0 {
		return nil
	}
	return nil
}

func (w *WakeList) addWake(item FutureTask) (*TaskSlot, error) {
	task := item.Task()
	if task == nil {
		return nil, errors.New("nil task")
	}
	slot := taskSlotPool.Get()
	slot.init(w, &w.wakes, item, task, true)
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed > 0 {
		return nil, os.ErrClosed
	}
	w.wakes.Add(slot)
	w.version++
	w.size++
	w.owner.numTasks.Incr()
	w.owner.numEntries.Incr()
	return slot, nil
}

func (w *WakeList) addInterval(item FutureTask) (*TaskSlot, error) {
	task := item.Task()
	if task == nil {
		return nil, errors.New("nil task")
	}
	slot := taskSlotPool.Get()
	slot.init(w, &w.intervals, item, task, false)
	slot.interval = task.interval
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed > 0 {
		return nil, os.ErrClosed
	}
	w.intervals.Add(slot)
	w.version++
	w.size++
	w.owner.numTasks.Incr()
	w.owner.numEntries.Incr()
	return slot, nil
}

func (w *WakeList) remove(slot *TaskSlot) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !slot.slice.Remove(slot) {
		return false
	}
	t := slot.task
	if t != nil {
		t.unlinkSlot(slot)
	}
	w.version++
	w.size--
	w.owner.decrTasks()
	taskSlotPool.Put(slot)
	return true
}

func (w *WakeList) addFunc(fn func()) (*FuncSlot, error) {
	if fn == nil {
		return nil, errors.New("nil func")
	}
	slot := funcSlotPool.Get()
	slot.init(w, fn)
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed > 0 {
		return nil, os.ErrClosed
	}
	w.funcs.Add(slot)
	w.version++
	w.size++
	w.owner.numFuncs.Incr()
	w.owner.numEntries.Incr()
	return slot, nil
}

func (w *WakeList) removeFunc(slot *FuncSlot) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.funcs.Remove(slot) {
		return false
	}
	w.version++
	w.size--
	w.owner.decrFuncs()
	funcSlotPool.Put(slot)
	return true
}

func (w *WakeList) invoke(fn func()) bool {
	r := w.reactor
	if r == nil {
		return true
	}
	return r.Invoke(fn)
}

func (w *WakeList) iter() {
	_ = w.reactor.Invoke(w.run)
}

// This is executed on reactor thread
func (w *WakeList) run() {
	r := w.reactor
	if r == nil {
		return
	}
	fn := w.fn
	if fn == nil {
		return
	}
	w.mu.Lock()
	w.wakes.Iterate(fn)
	w.mu.Unlock()
}

type TaskSlot struct {
	slot     *WakeList
	slice    *TaskSwapSlice
	idx      int
	future   FutureTask
	task     *Task
	lastWake int64
	interval time.Duration
	prev     *TaskSlot
	next     *TaskSlot
	wake     bool
}

func (ts *TaskSlot) init(wl *WakeList, slice *TaskSwapSlice, provider FutureTask, task *Task, wake bool) {
	ts.slot = wl
	ts.slice = slice
	ts.idx = -1
	ts.future = provider
	ts.task = task
	ts.lastWake = 0
	ts.interval = 0
	ts.wake = wake
}

func (ts *TaskSlot) swapIndex() int         { return ts.idx }
func (ts *TaskSlot) setSwapIndex(index int) { ts.idx = index }
func (ts *TaskSlot) LastWake() int64        { return ts.lastWake }

func (ts *TaskSlot) Remove() bool {
	if ts == nil {
		return false
	}
	slot := ts.slot
	if slot == nil {
		return false
	}
	ts.slot = nil
	return slot.remove(ts)
}

type FuncSlot struct {
	wl       *WakeList
	idx      int
	lastWake int64
	Value    func()
}

func (fs *FuncSlot) init(wl *WakeList, value func()) {
	fs.wl = wl
	fs.lastWake = 0
	fs.idx = -1
	fs.Value = value
}

func (fs *FuncSlot) swapIndex() int         { return fs.idx }
func (fs *FuncSlot) setSwapIndex(index int) { fs.idx = index }
func (fs *FuncSlot) LastWake() int64        { return fs.lastWake }
func (fs *FuncSlot) Remove() bool {
	if fs == nil {
		return false
	}
	slot := fs.wl
	if slot == nil {
		return false
	}
	fs.wl = nil
	return slot.removeFunc(fs)
}

type TaskSwapSlice struct {
	slots    []*TaskSlot
	lastWake int64
	mu       spinlock.Mutex
}

func (s *TaskSwapSlice) Len() int {
	return len(s.slots)
}

func (s *TaskSwapSlice) LastWake() int64 { return s.lastWake }

func (s *TaskSwapSlice) Add(value *TaskSlot) {
	value.setSwapIndex(len(s.slots))
	s.slots = append(s.slots, value)
}

func (s *TaskSwapSlice) Remove(value *TaskSlot) bool {
	if value == nil {
		return false
	}
	index := value.swapIndex()
	if index < 0 || index >= len(s.slots) {
		return false
	}

	if len(s.slots) == 1 {
		s.slots[0] = nil
		s.slots = s.slots[:0]
		return true
	}

	if s.slots[index] != value {
		return false
	}

	tailIndex := len(s.slots) - 1
	tail := s.slots[tailIndex]
	s.slots[tailIndex] = nil
	s.slots = s.slots[0:tailIndex]
	tail.setSwapIndex(index)
	return true
}

func (s *TaskSwapSlice) Get(index int) (value *TaskSlot, ok bool) {
	if index < 0 || index >= len(s.slots) {
		return
	}
	value = s.slots[index]
	ok = true
	return
}

func (s *TaskSwapSlice) Iterate(fn func(slot *TaskSlot) bool) int64 {
	if len(s.slots) == 0 || fn == nil {
		return 0
	}
	count := 0
	for ; count < len(s.slots); count++ {
		if !fn(s.slots[count]) {
			return int64(count)
		}
		count++
	}
	return int64(count)
}

func (s *TaskSwapSlice) wake(now int64, fn func(slot *TaskSlot)) int64 {
	if len(s.slots) == 0 || fn == nil {
		return 0
	}
	var (
		count = 0
		slot  *TaskSlot
	)
	for i := 0; i < len(s.slots); i++ {
		slot = s.slots[i]
		if slot == nil {
			continue
		}
		slot.lastWake = now
		fn(slot)
		count++
	}
	return int64(count)
}

func (s *TaskSwapSlice) Unsafe() []*TaskSlot {
	return s.slots
}

type FuncSwapSlice struct {
	slots    []*FuncSlot
	lastWake int64
	mu       spinlock.Mutex
}

func (s *FuncSwapSlice) Len() int {
	return len(s.slots)
}

func (s *FuncSwapSlice) LastWake() int64 { return s.lastWake }

func (s *FuncSwapSlice) Add(value *FuncSlot) {
	value.setSwapIndex(len(s.slots))
	s.slots = append(s.slots, value)
}

func (s *FuncSwapSlice) Remove(value *FuncSlot) bool {
	if value == nil {
		return false
	}
	index := value.swapIndex()
	if index < 0 || index >= len(s.slots) {
		return false
	}

	if len(s.slots) == 1 {
		s.slots[0] = nil
		s.slots = s.slots[:0]
		return true
	}

	if s.slots[index] != value {
		return false
	}

	tailIndex := len(s.slots) - 1
	tail := s.slots[tailIndex]
	s.slots[tailIndex] = nil
	s.slots = s.slots[0:tailIndex]
	tail.setSwapIndex(index)
	return true
}

func (s *FuncSwapSlice) Get(index int) (value *FuncSlot, ok bool) {
	if index < 0 || index >= len(s.slots) {
		return
	}
	value = s.slots[index]
	ok = true
	return
}

func (s *FuncSwapSlice) Iterate(fn func(slot *FuncSlot) bool) {
	if len(s.slots) == 0 || fn == nil {
		return
	}
	for _, s := range s.slots {
		if !fn(s) {
			return
		}
	}
}

func (s *FuncSwapSlice) wake(now int64, fn func(slot *FuncSlot)) int64 {
	s.lastWake = now
	if len(s.slots) == 0 {
		return 0
	}
	var (
		count = 0
		slot  *FuncSlot
	)
	for i := 0; i < len(s.slots); i++ {
		slot = s.slots[count]
		if slot == nil {
			continue
		}
		slot.lastWake = now
		fn(slot)
		count++
	}
	return int64(count)
}

func (s *FuncSwapSlice) Unsafe() []*FuncSlot {
	return s.slots
}
