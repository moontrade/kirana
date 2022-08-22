//go:build poll_no_opt
// +build poll_no_opt

package netpoll

import "golang.org/x/sys/unix"

type epollevent = unix.EpollEvent
