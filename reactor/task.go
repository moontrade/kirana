package reactor

import (
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/spinlock"
	"github.com/moontrade/kirana/pkg/util"
	logger "github.com/moontrade/log"
	"time"
	"unsafe"
)

type Waker interface {
	Wake() error
}

type TaskProvider struct {
	task *Task
}

func (tp *TaskProvider) Reactor() *Reactor {
	t := tp.task
	if t == nil {
		return nil
	}
	return t.reactor
}
func (tp *TaskProvider) Wake() error {
	t := tp.task
	if t == nil {
		return nil
	}
	return t.Wake()
}
func (tp *TaskProvider) Task() *Task        { return tp.task }
func (tp *TaskProvider) SetTask(task *Task) { tp.task = task }

type Task struct {
	id        int64
	reactor   *Reactor
	future    Future
	started   int64
	lastPoll  int64
	interval  time.Duration
	intervals int64
	wakes     int64
	wakeAfter time.Duration
	polls     int64
	pid       int64
	head      *TaskSlot
	mu        spinlock.Mutex
	stop      bool
}

func (t *Task) CheckGID() bool {
	r := t.reactor
	if r == nil {
		return true
	}
	return r.CheckGID()
}

func (t *Task) init(id int64, reactor *Reactor, future Future) {
	*t = Task{
		id:      id,
		reactor: reactor,
		future:  future,
	}
}

func (t *Task) linkSlot(slot *TaskSlot) {
	if slot.task != nil && slot.task != t {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	head := t.head
	if head == nil {
		t.head = slot
	} else {
		head.prev = slot
		slot.next = head
	}
}

func (t *Task) unlinkSlot(slot *TaskSlot) {
	if slot.task != t {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	slot.task = nil
	if t.head == slot {
		t.head = slot.next
	} else {
		var (
			prev = slot.prev
			next = slot.next
		)
		if prev != nil {
			prev.next = next
		}
		if next != nil {
			next.prev = prev
		}
	}
}

func (t *Task) clearSlots() {
	t.mu.Lock()
	defer t.mu.Unlock()
	next := t.head
	for next != nil {
		next.task = nil
		next.Remove()
		next = next.next
	}
}

func (t *Task) remove() {
	defer func() {
		if e := recover(); e != nil {
			err := util.PanicToError(e)
			logger.Error(err, "Task.remove panic")
		}
	}()
	t.clearSlots()
	*t = Task{}
	taskPool.Put(t)
}

func (t *Task) ID() int64               { return t.id }
func (t *Task) Reactor() *Reactor       { return t.reactor }
func (t *Task) Poll() Future            { return t.future }
func (t *Task) Started() int64          { return t.started }
func (t *Task) LastPoll() int64         { return t.lastPoll }
func (t *Task) Interval() time.Duration { return t.interval }
func (t *Task) Wakes() int64            { return t.wakes }
func (t *Task) Polls() int64            { return t.polls }
func (t *Task) Stop() bool              { return t.stop }
func (t *Task) SetStop(stop bool) {
	t.stop = stop
}
func (t *Task) Wake() error {
	reactor := t.reactor
	if reactor != nil {
		return reactor.Wake(t)
	} else {
		return nil
	}
}
func (t *Task) WakeAfter(duration time.Duration) error {
	t.wakeAfter = duration
	reactor := t.reactor
	if reactor != nil {
		return reactor.WakeAfter(t, duration)
	} else {
		return nil
	}
}

type Conn struct {
	fd int
	rd []byte
	wr []byte
}

const taskSlotSize = int(unsafe.Sizeof(taskSwapSlot{}))

type taskSwapList struct {
	slots []taskSwapSlot
	ptr   unsafe.Pointer
	size  int
	dur   time.Duration
	ticks int64
}

func newTaskList(capacity int) *taskSwapList {
	capacity = pmath.CeilToPowerOf2(capacity)
	if capacity < 4 {
		capacity = 4
	}
	r := make([]taskSwapSlot, capacity, capacity)
	return &taskSwapList{
		slots: r,
		ptr:   unsafe.Pointer(&r[0]),
		size:  0,
	}
}

func (tq *taskSwapList) alloc(task *Task, wake bool) *taskSwapSlot {
	if tq.size == cap(tq.slots) {
		if cap(tq.slots) == 0 {
			tq.slots = make([]taskSwapSlot, 16)
			tq.ptr = unsafe.Pointer(&tq.slots[0])
		} else {
			n := make([]taskSwapSlot, cap(tq.slots)*2)
			copy(n, tq.slots)
			tq.slots = n
			tq.ptr = unsafe.Pointer(&tq.slots[0])
		}
	}
	idx := tq.size
	tq.size++
	return tq.get(idx).set(task, wake)
}

func (tq *taskSwapList) get(idx int) *taskSwapSlot {
	return (*taskSwapSlot)(unsafe.Add(tq.ptr, idx*taskSlotSize))
}

func (tq *taskSwapList) clear(idx int) {
	if tq.size == 0 || idx >= tq.size {
		return
	}
	tq.size--
	if idx < tq.size {
		tq.get(idx).task = tq.get(tq.size).clear()
	} else {
		tq.get(idx).clear()
	}
}

func (tq *taskSwapList) iterate(
	now int64,
	fn func(
		now int64,
		list *taskSwapList,
		slot *taskSwapSlot,
		task *Task,
	) bool) {
	idx := 0
	for idx < tq.size {
		slot := tq.get(idx)
		if slot.wake {
			fn(now, tq, slot, slot.task)
			tq.size--
			if idx < tq.size {
				last := tq.get(tq.size)
				slot.task = last.task
				slot.wake = last.wake
			} else {
				slot.task = nil
				slot.wake = false
			}
		} else if !fn(now, tq, slot, slot.task) {
			tq.size--
		} else {
			idx++
		}
	}
}

type taskSwapSlot struct {
	task *Task
	fn   func()
	wake bool
}

func (t *taskSwapSlot) set(task *Task, wake bool) *taskSwapSlot {
	t.task = task
	t.wake = wake
	return t
}

func (t *taskSwapSlot) clear() *Task {
	r := t.task
	t.wake = false
	t.task = nil
	return r
}
