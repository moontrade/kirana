// Copyright 2017 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package netpoll

import (
	"sync/atomic"
	"syscall"
	"unsafe"
)

// Poll ...
type Poll[T any] struct {
	fd   int // epoll fd
	wfd  int // wake fd
	wait int32
}

// OpenPoll ...
func OpenPoll[T any]() *Poll[T] {
	l := new(Poll[T])
	p, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}
	l.fd = p
	r0, _, e0 := syscall.Syscall(syscall.SYS_EVENTFD2, 0, 0, 0)
	if e0 != 0 {
		syscall.Close(p)
		panic(err)
	}
	l.wfd = int(r0)
	l.AddRead(l.wfd)
	return l
}

// Close ...
func (p *Poll[T]) Close() error {
	if err := syscall.Close(p.wfd); err != nil {
		return err
	}
	return syscall.Close(p.fd)
}

// Wake ...
func (p *Poll[T]) Wake() error {
	if atomic.CompareAndSwapInt32(&p.wait, 0, 1) {
		var x uint64 = 1
		_, err := syscall.Write(p.wfd, (*(*[8]byte)(unsafe.Pointer(&x)))[:])
		return err
	}
	return nil
}

// Wait ...
func (p *Poll[T]) Wait(iter func(fd int, note bool) error) error {
	events := make([]syscall.EpollEvent, 64)
	for {
		n, err := syscall.EpollWait(p.fd, events, 100)
		if err != nil && err != syscall.EINTR {
			return err
		}
		for i := 0; i < n; i++ {
			if fd := int(events[i].Fd); fd != p.wfd {
				if err := iter(fd, false); err != nil {
					return err
				}
			} else if fd == p.wfd {
				var data [8]byte
				syscall.Read(p.wfd, data[:])
				iter(p.wfd, true)
			}
		}
	}
}

const (
	readEvents      = unix.EPOLLPRI | unix.EPOLLIN
	writeEvents     = unix.EPOLLOUT
	readWriteEvents = readEvents | writeEvents
)

// AddReadWrite ...
func (p *Poll[T]) AddReadWrite(fd int, data *T) error {
	var ev epollevent
	ev.events = readWriteEvents
	*(**T)(unsafe.Pointer(&ev.data)) = data
	return syscall.EpollCtl(
		p.fd, syscall.EPOLL_CTL_ADD, fd,
		&ev,
		//&syscall.EpollEvent{Fd: int32(fd),
		//	Events: syscall.EPOLLIN | syscall.EPOLLOUT,
		//},
	)
}

// AddRead ...
func (p *Poll[T]) AddRead(fd int, data *T) error {
	var ev epollevent
	ev.events = readEvents
	*(**T)(unsafe.Pointer(&ev.data)) = data
	return syscall.EpollCtl(
		p.fd, syscall.EPOLL_CTL_ADD, fd,
		&ev,
		//&syscall.EpollEvent{Fd: int32(fd),
		//	Events: syscall.EPOLLIN,
		//},
	)
}

// ModRead ...
func (p *Poll[T]) ModRead(fd int, data *T) error {
	var ev epollevent
	ev.events = readEvents
	*(**T)(unsafe.Pointer(&ev.data)) = data
	return syscall.EpollCtl(
		p.fd, syscall.EPOLL_CTL_MOD, fd,
		&ev,
		//&syscall.EpollEvent{Fd: int32(fd),
		//	Events: syscall.EPOLLIN,
		//},
	)
}

// ModReadWrite ...
func (p *Poll[T]) ModReadWrite(fd int, data *T) error {
	var ev epollevent
	ev.events = readWriteEvents
	*(**T)(unsafe.Pointer(&ev.data)) = data
	return syscall.EpollCtl(
		p.fd, syscall.EPOLL_CTL_MOD, fd,
		&ev,
		//&syscall.EpollEvent{Fd: int32(fd),
		//	Events: syscall.EPOLLIN | syscall.EPOLLOUT,
		//},
	)
}

// ModDetach ...
func (p *Poll[T]) ModDetach(fd int, data *T) error {
	var ev epollevent
	ev.events = syscall.EPOLLIN | syscall.EPOLLOUT
	*(**T)(unsafe.Pointer(&ev.data)) = data
	return syscall.EpollCtl(
		p.fd, syscall.EPOLL_CTL_DEL, fd,
		&ev,
		//&syscall.EpollEvent{Fd: int32(fd),
		//	Events: syscall.EPOLLIN | syscall.EPOLLOUT,
		//},
	)
}
