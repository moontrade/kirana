package reactor

import (
	"errors"
	"fmt"
	"time"
)

type Cadence struct {
	Durations []time.Duration
}

func (c *Cadence) Tick() time.Duration {
	if len(c.Durations) == 0 {
		return 0
	}
	return c.Durations[0]
}

func clone[T any](s []T) []T {
	c := make([]T, len(s))
	copy(c, s)
	return c
}

type Wheel struct {
	tickDur   time.Duration
	maxDur    time.Duration
	durations []time.Duration
	wheel     [][]*taskSwapList
	current   int64
	size      int64
	lists     int
}

var (
	Micros50 = Cadence{Durations: []time.Duration{
		time.Microsecond * 50,
		time.Microsecond * 100,
		time.Microsecond * 150,
		time.Microsecond * 200,
		time.Microsecond * 250,
		time.Microsecond * 350,
		time.Microsecond * 500,
		time.Microsecond * 750,
		time.Microsecond * 900,
		time.Microsecond * 1000,
	}}
	Micros250 = Cadence{Durations: []time.Duration{
		time.Microsecond * 250,
		time.Microsecond * 500,
		time.Microsecond * 750,
		time.Microsecond * 1000,
	}}
	Millis5 = Cadence{Durations: []time.Duration{
		time.Millisecond * 5,
		time.Millisecond * 10,
		time.Millisecond * 20,
		time.Millisecond * 30,
		time.Millisecond * 50,
		time.Millisecond * 100,
		time.Millisecond * 200,
		time.Millisecond * 250,
		time.Millisecond * 500,
	}}
	Millis10 = Cadence{Durations: []time.Duration{
		time.Millisecond * 10,
		time.Millisecond * 20,
		time.Millisecond * 30,
		time.Millisecond * 50,
		time.Millisecond * 100,
		time.Millisecond * 200,
		time.Millisecond * 250,
		time.Millisecond * 500,
	}}
	Millis20 = Cadence{Durations: []time.Duration{
		time.Millisecond * 20,
		time.Millisecond * 40,
		time.Millisecond * 60,
		time.Millisecond * 100,
		time.Millisecond * 200,
		time.Millisecond * 500,
	}}
	Millis25 = Cadence{Durations: []time.Duration{
		time.Millisecond * 25,
		time.Millisecond * 50,
		time.Millisecond * 75,
		time.Millisecond * 100,
		time.Millisecond * 150,
		time.Millisecond * 250,
		time.Millisecond * 500,
		time.Millisecond * 750,
	}}
	Millis50 = Cadence{Durations: []time.Duration{
		time.Millisecond * 50,
		time.Millisecond * 100,
		time.Millisecond * 200,
		time.Millisecond * 250,
		time.Millisecond * 500,
		time.Millisecond * 750,
	}}
	Millis100 = Cadence{Durations: []time.Duration{
		time.Millisecond * 100,
		time.Millisecond * 200,
		time.Millisecond * 500,
	}}
	Millis250 = Cadence{Durations: []time.Duration{
		time.Millisecond * 250,
		time.Millisecond * 500,
		time.Millisecond * 750,
	}}
	Seconds = Cadence{Durations: []time.Duration{
		time.Second,
		time.Second * 2,
		time.Second * 3,
		time.Second * 5,
		time.Second * 10,
		time.Second * 15,
		time.Second * 20,
		time.Second * 30,
		time.Second * 45,
	}}
	Minutes = Cadence{Durations: []time.Duration{
		time.Second * 5,
		time.Second * 60,
		time.Minute * 2,
		time.Minute * 3,
		time.Minute * 5,
		time.Minute * 10,
		time.Minute * 15,
		time.Minute * 20,
		time.Minute * 30,
	}}
	Hours = Cadence{Durations: []time.Duration{
		time.Minute * 5,
		time.Hour,
		time.Hour * 2,
		time.Hour * 4,
		time.Hour * 6,
		time.Hour * 12,
		time.Hour * 24,
	}}
)

func NewWheel(cadence Cadence) Wheel {
	durations := cadence.Durations
	w, _ := newWheel(durations)
	return w
}

func newWheel(durations []time.Duration) (Wheel, error) {
	if len(durations) == 0 {
		return Wheel{}, errors.New("durations empty")
	}
	tick := durations[0]
	totalSlots := 0
	wheels := make([][]*taskSwapList, len(durations))
	for i := 0; i < len(wheels); i++ {
		dur := durations[i]
		if i > 0 && dur < durations[i-1] {
			return Wheel{}, fmt.Errorf("durations must get progressively longer: %s < %s", dur, durations[i-1])
		}
		if dur%tick != 0 {
			return Wheel{}, fmt.Errorf("durations must be divisible by tickDur: %s mod %s = %s", dur, tick, dur%tick)
		}
		size := int(dur / tick)
		subWheel := make([]*taskSwapList, size)
		totalSlots += size
		for x := 0; x < len(subWheel); x++ {
			list := &taskSwapList{}
			list.dur = dur
			list.ticks = int64(size)
			subWheel[x] = list
		}
		wheels[i] = subWheel
	}
	return Wheel{
		tickDur:   tick,
		durations: durations,
		maxDur:    durations[len(durations)-1],
		wheel:     wheels,
		lists:     totalSlots,
	}, nil
}

func (w *Wheel) tick(now int64, fn func(now int64, list *taskSwapList, slot *taskSwapSlot, task *Task) bool) {
	t := w.current
	w.current++

	for i := 0; i < len(w.wheel); i++ {
		list := w.wheel[i]
		slot := list[int(t)%len(list)]
		if slot.size == 0 {
			continue
		}
		slot.iterate(now, fn)
	}
}

func (w *Wheel) schedule(task *Task, duration time.Duration, wake bool) bool {
	if duration > w.maxDur {
		return false
	}
	var (
		current   = uint64(w.current)
		currentM1 = uint64(current) - 1
	)
	for i := 0; i < len(w.durations); i++ {
		if duration <= w.durations[i] {
			list := w.wheel[i]
			if len(list) == 1 {
				list[0].alloc(task, wake)
				w.size++
				return true
			}
			slot := list[currentM1%uint64(len(list))]
			slot.alloc(task, wake)
			w.size++
			return true
		}
	}
	return false
}
