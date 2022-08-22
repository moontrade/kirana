// Copyright 2017 Joshua J Baker. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build darwin || netbsd || freebsd || openbsd || dragonfly

package netpoll

import "C"
import (
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

var wakeEvents = []syscall.Kevent_t{{
	Ident:  0,
	Filter: syscall.EVFILT_USER,
	Fflags: syscall.NOTE_TRIGGER,
}}

type Poll[T any] struct {
	fd      int
	changes []syscall.Kevent_t
	wait    int32
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
	if atomic.CompareAndSwapInt32(&p.wait, 0, 1) {
		_, err := syscall.Kevent(p.fd, wakeEvents, nil, nil)
		return err
	}
	return nil
}

//goland:noinspection ALL
func (p *Poll[T]) Wait(
	timeout time.Duration,
	onEvent func(index, count, fd int, attachment *T) error,
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
		n, err = syscall.Kevent(p.fd, p.changes, events, &timespec)
		if err != nil && err != syscall.EINTR {
			return err
		}
		p.changes = p.changes[:0]

		// Timeout?
		if n == 0 {
			if timeout, err = onNextWait(0); err != nil {
				return err
			}
		}

		for i := 0; i < n; i++ {
			event := &events[i]
			if event.Ident == 0 {
				if err := onEvent(i, n, 0, nil); err != nil {
					return err
				}

				//if err := onEvent(0, n, 0, nil); err != nil {
				//	return err
				//}
			} else {
				if err := onEvent(i, n, int(event.Ident), (*T)(unsafe.Pointer(uintptr(event.Data)))); err != nil {
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
func (p *Poll[T]) AddRead(fd int, data *T) {
	p.changes = append(p.changes,
		syscall.Kevent_t{
			Ident:  uint64(fd),
			Flags:  syscall.EV_ADD,
			Filter: syscall.EVFILT_READ,
			Data:   int64(uintptr(unsafe.Pointer(data))),
		},
	)
}

// AddReadWrite ...
func (p *Poll[T]) AddReadWrite(fd int, data *T) {
	p.changes = append(p.changes,
		syscall.Kevent_t{
			Ident:  uint64(fd),
			Flags:  syscall.EV_ADD,
			Filter: syscall.EVFILT_READ,
			Data:   int64(uintptr(unsafe.Pointer(data))),
		},
		syscall.Kevent_t{
			Ident:  uint64(fd),
			Flags:  syscall.EV_ADD,
			Filter: syscall.EVFILT_WRITE,
			Data:   int64(uintptr(unsafe.Pointer(data))),
		},
	)
}

// ModRead ...
func (p *Poll[T]) ModRead(fd int, data *T) {
	p.changes = append(p.changes, syscall.Kevent_t{
		Ident:  uint64(fd),
		Flags:  syscall.EV_DELETE,
		Filter: syscall.EVFILT_WRITE,
		Data:   int64(uintptr(unsafe.Pointer(data))),
	})
}

// ModReadWrite ...
func (p *Poll[T]) ModReadWrite(fd int, data *T) {
	p.changes = append(p.changes, syscall.Kevent_t{
		Ident:  uint64(fd),
		Flags:  syscall.EV_ADD,
		Filter: syscall.EVFILT_WRITE,
		Data:   int64(uintptr(unsafe.Pointer(data))),
	})
}

// ModDetach ...
func (p *Poll[T]) ModDetach(fd int, data *T) {
	p.changes = append(p.changes,
		syscall.Kevent_t{
			Ident:  uint64(fd),
			Flags:  syscall.EV_DELETE,
			Filter: syscall.EVFILT_READ,
			Data:   int64(uintptr(unsafe.Pointer(data))),
		},
		syscall.Kevent_t{
			Ident:  uint64(fd),
			Flags:  syscall.EV_DELETE,
			Filter: syscall.EVFILT_WRITE,
			Data:   int64(uintptr(unsafe.Pointer(data))),
		},
	)
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
