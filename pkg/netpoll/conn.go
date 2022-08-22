package netpoll

import (
	"github.com/moontrade/wormhole/pkg/socket"
	"golang.org/x/sys/unix"
	"net"
)

type Conn[T any] struct {
	loop       *Loop[T]
	ctx        T
	fd         int
	peer       unix.Sockaddr // remote socket address
	localAddr  net.Addr      // local addr
	remoteAddr net.Addr      // remote addr
	rd         []byte
	wr         []byte
	isDatagram bool // UDP protocol
	opened     bool // connection opened event fired
}

func newTCPConn[T any](
	fd int, el *Loop[T],
	sa unix.Sockaddr,
	localAddr, remoteAddr net.Addr) *Conn[T] {
	c := &Conn[T]{
		fd:         fd,
		peer:       sa,
		loop:       el,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}
	return c
}

func newUDPConn[T any](
	fd int, el *Loop[T],
	localAddr net.Addr, sa unix.Sockaddr,
	connected bool) *Conn[T] {
	c := &Conn[T]{
		fd:         fd,
		peer:       sa,
		loop:       el,
		localAddr:  localAddr,
		remoteAddr: socket.SockaddrToUDPAddr(sa),
		isDatagram: true,
	}
	if connected {
		c.peer = nil
	}
	return c
}

func (c *Conn[T]) releaseTCP() {
	c.opened = false
	c.peer = nil
	c.ctx = nil
	//c.buffer = nil
	if addr, ok := c.localAddr.(*net.TCPAddr); ok && c.localAddr != c.loop.ln.addr {
		_ = addr
		//bsPool.Put(addr.IP)
		//if len(addr.Zone) > 0 {
		//	bsPool.Put(toolkit.StringToBytes(addr.Zone))
		//}
	}
	if addr, ok := c.remoteAddr.(*net.TCPAddr); ok {
		_ = addr
		//bsPool.Put(addr.IP)
		//if len(addr.Zone) > 0 {
		//	bsPool.Put(toolkit.StringToBytes(addr.Zone))
		//}
	}
	c.localAddr = nil
	c.remoteAddr = nil
	//c.inboundBuffer.Done()
	//c.outboundBuffer.Free()
	//netpoll.PutPollAttachment(c.pollAttachment)
	//c.pollAttachment = nil
}
