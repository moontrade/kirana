// Copyright 2017 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package netpoll

import (
	"golang.org/x/sys/unix"
	"sync/atomic"
	"syscall"
	"time"
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
	l.AddRead(l.wfd, nil)
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

// Do the interface allocations only once for common
// Errno values.
var (
	errEAGAIN error = unix.EAGAIN
	errEINVAL error = unix.EINVAL
	errENOENT error = unix.ENOENT
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e unix.Errno) error {
	switch e {
	case unix.EAGAIN:
		return errEAGAIN
	case unix.EINVAL:
		return errEINVAL
	case unix.ENOENT:
		return errENOENT
	}
	return e
}

// Single-word zero for use when we need a valid pointer to 0 bytes.
// See mksyscall.pl.
var _zero uintptr

func epollWait(epfd int, events []epollevent, msec int) (int, error) {
	var ep unsafe.Pointer
	if len(events) > 0 {
		ep = unsafe.Pointer(&events[0])
	} else {
		ep = unsafe.Pointer(&_zero)
	}
	var (
		np    uintptr
		errno unix.Errno
	)
	if msec == 0 { // non-block system call, use RawSyscall6 to avoid getting preempted by runtime
		np, _, errno = unix.RawSyscall6(unix.SYS_EPOLL_WAIT, uintptr(epfd), uintptr(ep), uintptr(len(events)), 0, 0, 0)
	} else {
		np, _, errno = unix.Syscall6(unix.SYS_EPOLL_WAIT, uintptr(epfd), uintptr(ep), uintptr(len(events)), uintptr(msec), 0, 0)
	}
	if errno != 0 {
		return int(np), errnoErr(errno)
	}
	return int(np), nil
}

// Wait ...
func (p *Poll[T]) Wait(
	timeout time.Duration,
	onEvent func(index, count, fd int, filter int16, attachment *T) error,
	onNextWait func(count int) (time.Duration, error),
) error {
	events := make([]epollevent, 128)
	timeoutMsec := int(timeout / time.Millisecond)
	for {
		n, err := epollWait(p.fd, events, timeoutMsec)
		if err != nil && err != syscall.EINTR {
			return err
		}
		if n == 0 {
			timeout, err = onNextWait(0)
			continue
		}
		for i := 0; i < n; i++ {
			if fd := int(events[i].events); fd != p.wfd {
				if err := onEvent(i, n, fd, 0, *(**T)(unsafe.Pointer(&events[i]))); err != nil {
					return err
				}
			} else if fd == p.wfd {
				var data [8]byte
				syscall.Read(p.wfd, data[:])
				timeout, err = onNextWait(n)
				//iter(p.wfd, true)
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
	_, _, e1 := syscall.RawSyscall6(syscall.SYS_EPOLL_CTL, uintptr(p.fd), uintptr(syscall.EPOLL_CTL_ADD), uintptr(fd), uintptr(unsafe.Pointer(&ev)), 0, 0)
	//if e1 != 0 {
	//	err = errnoErr(e1)
	//}
	return e1
	//return syscall.EpollCtl(
	//	p.fd, syscall.EPOLL_CTL_ADD, fd,
	//	&ev,
	//	//&syscall.EpollEvent{Fd: int32(fd),
	//	//	Events: syscall.EPOLLIN | syscall.EPOLLOUT,
	//	//},
	//)
}

// AddRead ...
func (p *Poll[T]) AddRead(fd int, data *T) error {
	var ev epollevent
	ev.events = readEvents
	*(**T)(unsafe.Pointer(&ev.data)) = data
	_, _, e1 := syscall.RawSyscall6(syscall.SYS_EPOLL_CTL, uintptr(p.fd), uintptr(syscall.EPOLL_CTL_ADD), uintptr(fd), uintptr(unsafe.Pointer(&ev)), 0, 0)
	//if e1 != 0 {
	//	err = errnoErr(e1)
	//}
	return e1
	//return syscall.EpollCtl(
	//	p.fd, syscall.EPOLL_CTL_ADD, fd,
	//	&ev,
	//	//&syscall.EpollEvent{Fd: int32(fd),
	//	//	Events: syscall.EPOLLIN,
	//	//},
	//)
}

// ModRead ...
func (p *Poll[T]) ModRead(fd int, data *T) error {
	var ev epollevent
	ev.events = readEvents
	*(**T)(unsafe.Pointer(&ev.data)) = data
	_, _, e1 := syscall.RawSyscall6(syscall.SYS_EPOLL_CTL, uintptr(p.fd), uintptr(syscall.EPOLL_CTL_MOD), uintptr(fd), uintptr(unsafe.Pointer(&ev)), 0, 0)
	//if e1 != 0 {
	//	err = errnoErr(e1)
	//}
	return e1
	//return syscall.EpollCtl(
	//	p.fd, syscall.EPOLL_CTL_MOD, fd,
	//	&ev,
	//	//&syscall.EpollEvent{Fd: int32(fd),
	//	//	Events: syscall.EPOLLIN,
	//	//},
	//)
}

// ModReadWrite ...
func (p *Poll[T]) ModReadWrite(fd int, data *T) error {
	var ev epollevent
	ev.events = readWriteEvents
	*(**T)(unsafe.Pointer(&ev.data)) = data
	_, _, e1 := syscall.RawSyscall6(syscall.SYS_EPOLL_CTL, uintptr(p.fd), uintptr(syscall.EPOLL_CTL_MOD), uintptr(fd), uintptr(unsafe.Pointer(&ev)), 0, 0)
	//if e1 != 0 {
	//	err = errnoErr(e1)
	//}
	return e1
	//return syscall.EpollCtl(
	//	p.fd, syscall.EPOLL_CTL_MOD, fd,
	//	&ev,
	//	//&syscall.EpollEvent{Fd: int32(fd),
	//	//	Events: syscall.EPOLLIN | syscall.EPOLLOUT,
	//	//},
	//)
}

// ModDetach ...
func (p *Poll[T]) ModDetach(fd int, data *T) error {
	var ev epollevent
	ev.events = syscall.EPOLLIN | syscall.EPOLLOUT
	*(**T)(unsafe.Pointer(&ev.data)) = data
	_, _, e1 := syscall.RawSyscall6(syscall.SYS_EPOLL_CTL, uintptr(p.fd), uintptr(syscall.EPOLL_CTL_DEL), uintptr(fd), uintptr(unsafe.Pointer(&ev)), 0, 0)
	//if e1 != 0 {
	//	err = errnoErr(e1)
	//}
	return e1
	//return syscall.EpollCtl(
	//	p.fd, syscall.EPOLL_CTL_DEL, fd,
	//	&ev,
	//	//&syscall.EpollEvent{Fd: int32(fd),
	//	//	Events: syscall.EPOLLIN | syscall.EPOLLOUT,
	//	//},
	//)
}
