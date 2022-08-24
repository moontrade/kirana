//go:build darwin || dragonfly || freebsd || linux || openbsd || solaris || netbsd
// +build darwin dragonfly freebsd linux openbsd solaris netbsd

package mmap

import (
	"golang.org/x/sys/unix"
)

func mmap(len int, inprot, inflags, fd uintptr, off int64) ([]byte, error) {
	//flags := unix.MAP_PRIVATE
	flags := unix.MAP_SHARED
	prot := unix.PROT_READ
	switch {
	case inprot&COPY != 0:
		prot |= unix.PROT_WRITE
		flags = unix.MAP_PRIVATE
	case inprot&RDWR != 0:
		prot |= unix.PROT_WRITE
	}
	if inprot&EXEC != 0 {
		prot |= unix.PROT_EXEC
	}
	if inflags&ANON != 0 {
		flags |= unix.MAP_ANON
	}

	b, err := unix.Mmap(int(fd), off, len, prot, flags)
	if err != nil {
		return nil, err
	}
	return b, nil
}

//func Remap(m MMap, len int, inprot, inflags, fd uintptr, off int64) (MMap, error) {
//	unix.Syscall6(
//		unix.SYS_MREMAP)
//}

func (m MMap) flush() error {
	return unix.Msync(m, unix.MS_SYNC)
}

func (m MMap) flushAsync() error {
	return unix.Msync(m, unix.MS_ASYNC)
}

func (m MMap) lock() error {
	return unix.Mlock(m)
}

func (m MMap) unlock() error {
	return unix.Munlock(m)
}

func (m MMap) unmap() error {
	return unix.Munmap(m)
}
