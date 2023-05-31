// Copyright 2017 Joshua J Baker. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build darwin || netbsd || freebsd || openbsd || dragonfly

package netpoll

import "C"
import (
	"github.com/panjf2000/gnet/v2/pkg/errors"
	"golang.org/x/sys/unix"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

// IOEvent is the integer type of I/O events on BSD's.
type IOEvent = int16

const (
	// InitPollEventsCap represents the initial capacity of poller event-list.
	InitPollEventsCap = 64
	// MaxPollEventsCap is the maximum limitation of events that the poller can process.
	MaxPollEventsCap = 512
	// MinPollEventsCap is the minimum limitation of events that the poller can process.
	MinPollEventsCap = 16
	// MaxAsyncTasksAtOneTime is the maximum amount of asynchronous tasks that the event-loop will process at one time.
	MaxAsyncTasksAtOneTime = 128
	// EVFilterWrite represents writeable events from sockets.
	EVFilterWrite = unix.EVFILT_WRITE
	// EVFilterRead represents readable events from sockets.
	EVFilterRead = unix.EVFILT_READ
	// EVFilterSock represents exceptional events that are not read/write, like socket being closed,
	// reading/writing from/to a closed socket, etc.
	EVFilterSock = -0xd
)

var wakeEvents = []syscall.Kevent_t{{
	Ident:  0,
	Filter: syscall.EVFILT_USER,
	Fflags: syscall.NOTE_TRIGGER,
}}

type Poll[T any] struct {
	fd   int
	wait int32
}

func OpenPoll[T any]() *Poll[T] {
	l := new(Poll[T])
	p, err := syscall.Kqueue()
	if err != nil {
		panic(err)
	}
	l.fd = p
	_, err = syscall.Kevent(l.fd, []syscall.Kevent_t{{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}}, nil, nil)
	if err != nil {
		panic(err)
	}
	return l
}

// Close ...
func (p *Poll[T]) Close() error {
	return syscall.Close(p.fd)
}

// Wake ...
func (p *Poll[T]) Wake() error {
	//if atomicx.Casint32(&p.wait, 0, 1) {
	//if atomicx.Xchgint32(&p.wait, 1) == 0 {
	if atomic.CompareAndSwapInt32(&p.wait, 0, 1) {
		_, err := syscall.Kevent(p.fd, wakeEvents, nil, nil)
		return err
	}
	return nil
	//
	//_, err := syscall.Kevent(p.fd, wakeEvents, nil, nil)
	//return err
}

//goland:noinspection ALL
func (p *Poll[T]) Wait(
	timeout time.Duration,
	onEvent func(index, count, fd int, filter int16, attachment *T) error,
	onNextWait func(count int) (time.Duration, error),
) error {
	var (
		events   = make([]syscall.Kevent_t, 128)
		timespec = syscall.Timespec{
			Sec:  0,
			Nsec: int64(time.Millisecond * 250),
		}
		n   int
		err error
	)
	for {
		n, err = syscall.Kevent(p.fd, nil, events, &timespec)
		if n == 0 || (n < 0 && err == syscall.EINTR) {
			runtime.Gosched()
			continue
		} else if err != nil {
			return err
		}

		// Timeout?
		if n == 0 {
			if timeout, err = onNextWait(0); err != nil {
				return err
			}
		}

		var evFilter int16
		for i := 0; i < n; i++ {
			event := &events[i]
			if event.Ident != 0 {
				evFilter = event.Filter
				if (event.Flags&unix.EV_EOF != 0) || (event.Flags&unix.EV_ERROR != 0) {
					evFilter = EVFilterSock
				}
				attachement := (*T)(unsafe.Pointer(event.Udata))
				switch err = onEvent(i, n, int(event.Ident), evFilter, attachement); err {
				case nil:
				case errors.ErrAcceptSocket, errors.ErrEngineShutdown:
					return err
				default:
				}
			} else {
				if err := onEvent(i, n, int(event.Ident), 0, (*T)(unsafe.Pointer(uintptr(event.Data)))); err != nil {
					return err
				}
			}
		}

		atomic.StoreInt32(&p.wait, 0)
		if timeout, err = onNextWait(n); err != nil {
			return err
		}
		timespec.Nsec = int64(timeout)
	}
}

// AddRead ...
func (p *Poll[T]) AddRead(fd int, data *T) error {
	var evs [1]syscall.Kevent_t
	evs[0] = syscall.Kevent_t{
		Ident:  uint64(fd),
		Flags:  syscall.EV_ADD,
		Filter: syscall.EVFILT_READ,
		Data:   int64(uintptr(unsafe.Pointer(data))),
	}
	_, err := syscall.Kevent(p.fd, evs[:], nil, nil)
	return err
}

// AddReadWrite ...
func (p *Poll[T]) AddReadWrite(fd int, data *T) error {
	var evs [2]syscall.Kevent_t
	evs[0] = syscall.Kevent_t{
		Ident:  uint64(fd),
		Flags:  syscall.EV_ADD,
		Filter: syscall.EVFILT_READ,
		Data:   int64(uintptr(unsafe.Pointer(data))),
	}
	evs[1] = syscall.Kevent_t{
		Ident:  uint64(fd),
		Flags:  syscall.EV_ADD,
		Filter: syscall.EVFILT_WRITE,
		Data:   int64(uintptr(unsafe.Pointer(data))),
	}
	_, err := syscall.Kevent(p.fd, evs[:], nil, nil)
	return err
}

// ModRead ...
func (p *Poll[T]) ModRead(fd int, data *T) error {
	var evs [1]syscall.Kevent_t
	evs[0] = syscall.Kevent_t{
		Ident:  uint64(fd),
		Flags:  syscall.EV_DELETE,
		Filter: syscall.EVFILT_WRITE,
		Data:   int64(uintptr(unsafe.Pointer(data))),
	}
	_, err := syscall.Kevent(p.fd, evs[:], nil, nil)
	return err
}

// ModReadWrite ...
func (p *Poll[T]) ModReadWrite(fd int, data *T) error {
	var evs [1]syscall.Kevent_t
	evs[0] = syscall.Kevent_t{
		Ident:  uint64(fd),
		Flags:  syscall.EV_ADD,
		Filter: syscall.EVFILT_WRITE,
		Data:   int64(uintptr(unsafe.Pointer(data))),
	}
	_, err := syscall.Kevent(p.fd, evs[:], nil, nil)
	return err
}

// ModDetach ...
func (p *Poll[T]) ModDetach(fd int, data *T) error {
	var evs [2]syscall.Kevent_t
	evs[0] = syscall.Kevent_t{
		Ident:  uint64(fd),
		Flags:  syscall.EV_DELETE,
		Filter: syscall.EVFILT_READ,
		Data:   int64(uintptr(unsafe.Pointer(data))),
	}
	evs[1] = syscall.Kevent_t{
		Ident:  uint64(fd),
		Flags:  syscall.EV_DELETE,
		Filter: syscall.EVFILT_WRITE,
		Data:   int64(uintptr(unsafe.Pointer(data))),
	}
	_, err := syscall.Kevent(p.fd, evs[:], nil, nil)
	return err
}

//// Dequeue ...
//type Dequeue struct {
//	fd      int
//	changes []syscall.Kevent_t
//	wait    int32
//	//notes   noteQueue
//}
//
//// OpenPoll ...
//func OpenPoll() *Dequeue {
//	l := new(Dequeue)
//	p, err := syscall.Kqueue()
//	if err != nil {
//		panic(err)
//	}
//	l.fd = p
//	_, err = syscall.Kevent(l.fd, []syscall.Kevent_t{{
//		Ident:  0,
//		Filter: syscall.EVFILT_USER,
//		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
//	}}, nil, nil)
//	if err != nil {
//		panic(err)
//	}
//
//	return l
//}
//
//// Close ...
//func (p *Dequeue) Close() error {
//	return syscall.Close(p.fd)
//}
//
//// Wake ...
//func (p *Dequeue) Wake() error {
//	if atomic.CompareAndSwapInt32(&p.wait, 0, 1) {
//		_, err := syscall.Kevent(p.fd, wakeEvents, nil, nil)
//		return err
//	}
//	return nil
//}
//
//// Wait ...
//func (p *Dequeue) Wait(iter func(fd int, note bool) error) error {
//	events := make([]syscall.Kevent_t, 128)
//	timespec := syscall.Timespec{
//		Sec:  0,
//		Nsec: int64(time.Millisecond * 500),
//	}
//	for {
//		n, err := syscall.Kevent(p.fd, p.changes, events, &timespec)
//		if err != nil && err != syscall.EINTR {
//			return err
//		}
//		p.changes = p.changes[:0]
//
//		if n == 0 {
//			//println("timeout")
//		}
//
//		for i := 0; i < n; i++ {
//			fd := events[i].Ident
//			if fd == 0 {
//				if err := iter(0, true); err != nil {
//					return err
//				}
//				atomic.StoreInt32(&p.wait, 0)
//				if err := iter(0, true); err != nil {
//					return err
//				}
//			} else {
//				if err := iter(int(fd), false); err != nil {
//					return err
//				}
//			}
//		}
//	}
//}
//
//// AddRead ...
//func (p *Dequeue) AddRead(fd int) {
//	p.changes = append(p.changes,
//		syscall.Kevent_t{
//			Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ,
//		},
//	)
//}
//
//// AddReadWrite ...
//func (p *Dequeue) AddReadWrite(fd int) {
//	p.changes = append(p.changes,
//		syscall.Kevent_t{
//			Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ,
//		},
//		syscall.Kevent_t{
//			Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_WRITE,
//		},
//	)
//}
//
//// ModRead ...
//func (p *Dequeue) ModRead(fd int) {
//	p.changes = append(p.changes, syscall.Kevent_t{
//		Ident: uint64(fd), Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_WRITE,
//	})
//}
//
//// ModReadWrite ...
//func (p *Dequeue) ModReadWrite(fd int) {
//	p.changes = append(p.changes, syscall.Kevent_t{
//		Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_WRITE,
//	})
//}
//
//// ModDetach ...
//func (p *Dequeue) ModDetach(fd int) {
//	p.changes = append(p.changes,
//		syscall.Kevent_t{
//			Ident: uint64(fd), Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_READ,
//		},
//		syscall.Kevent_t{
//			Ident: uint64(fd), Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_WRITE,
//		},
//	)
//}
