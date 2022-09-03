package reactor

import (
	"time"
)

type Future interface {
	Poll(ctx Context) error
}

type PollClose interface {
	PollClose(ev CloseEvent) error
}

type PollReason uint8

const (
	ReasonStart          PollReason = 0 // ReasonStart first time Poll is invoked after creating Task
	ReasonWake           PollReason = 1 // ReasonWake the Task is awoken. Namaste
	ReasonInterval       PollReason = 2 // ReasonInterval the Task's interval has elapsed.
	ReasonIntervalBehind PollReason = 3 // ReasonIntervalBehind the Task has missed interval wakes due to overloaded Reactor
	ReasonPing           PollReason = 4 // ReasonPing
	ReasonClose          PollReason = 5 // ReasonClose the Task will immediately close on return
)

type CloseEvent struct {
	Task   *Task
	Time   int64
	Reason any
}

// Context provides the low-level management of a Task with its Reactor.
type Context struct {
	// Task is the associated Task object.
	Task *Task
	// Time is a possible low frequency time from the reactor.
	// Each Reactor loop pass captures a high frequency nano-time which
	// is what this field is set to. High frequency time is relatively
	// expensive 10-30ns and a sudden burst of task polls or function invokes
	// can add 10-30ns per call which adds up.
	//
	// Generally, this has a precision in the nanosecond to microsecond level
	// and possibly single-digit millisecond to tick duration. Values beyond this
	// means the Reactor is overloaded.
	Time int64
	// Interval is the current interval of the Task or 0 if no interval is set.
	// If the Interval value changes, then the Reactor will subscribe or resubscribe
	// or unsubscribe as necessary.
	Interval time.Duration
	// After represents the amount of time to wait before waking. This is a one-time
	// wake and generally used for cooperative scheduling (sharing CPU between all Tasks).
	After time.Duration
	// Reason why Poll was called
	Reason PollReason
}

// SetInterval sets the interval for the task
func (p *Context) SetInterval(duration time.Duration) {
	p.Task.interval = duration
}

// WakeOnNextTick this task again on the next tick
func (p *Context) WakeOnNextTick() {
	p.Task.wakeAfter = time.Nanosecond
}

// WakeAfter wakes this task again after the specified time.Duration
func (p *Context) WakeAfter(duration time.Duration) {
	p.Task.wakeAfter = duration
}

// Stop marks the task to be stopped and deleted
func (p *Context) Stop() {
	p.Task.stop = true
}

// Reactor the Task belongs to.
func (p *Context) Reactor() *Reactor {
	return p.Task.reactor
}
