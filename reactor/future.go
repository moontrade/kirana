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
	ReasonStart    PollReason = 0
	ReasonWake     PollReason = 1
	ReasonInterval PollReason = 2
	ReasonPing     PollReason = 3
	ReasonClose    PollReason = 4
)

type CloseEvent struct {
	Task   *Task
	Time   int64
	Reason any
}

type Context struct {
	Task     *Task
	Time     int64
	Interval time.Duration
	After    time.Duration
	Reason   PollReason
}

// SetInterval sets the interval for the task
func (p *Context) SetInterval(duration time.Duration) {
	p.Task.interval = duration
}

// Wake this task again on the next tick
func (p *Context) Wake() {
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

// Reactor the task belongs to
func (p *Context) Reactor() *Reactor {
	return p.Task.reactor
}
