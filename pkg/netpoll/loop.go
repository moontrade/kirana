package netpoll

import (
	"errors"
	"fmt"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/hashmap"
	"github.com/moontrade/kirana/pkg/socket"
	"github.com/moontrade/kirana/pkg/timex"
	"golang.org/x/sys/unix"
	"os"
	"runtime"
	"syscall"
	"time"
)

type Loop[T any] struct {
	idx      int
	poll     *Poll[Conn[T]]
	ln       *Listener
	packet   []byte
	active   counter.Counter
	conns    hashmap.SyncMap[int, *Conn[T]]
	udpConns hashmap.SyncMap[int, *Conn[T]]
	//conns        map[int]*Conn[T]
	//udpConns     map[int]*Conn[T]
	count        int32
	lockThread   bool
	tick         time.Duration
	onOpened     func(conn *Conn[T]) Action
	onTraffic    func(c *Conn[T], read bool) Action
	doRead       func(data *Conn[T]) ([]byte, error)
	doWrite      func(data *Conn[T]) ([]byte, error)
	onShortWrite func(data *Conn[T], wrote int, remaining []byte) error
}

func NewLoop[T any](ln *Listener) (*Loop[T], error) {
	if ln == nil {
		return nil, errors.New("nil listener")
	}
	l := &Loop[T]{
		poll:     OpenPoll[Conn[T]](),
		ln:       ln,
		conns:    *hashmap.NewSyncMap[int, *Conn[T]](0, 128, hashmap.HashInt),
		udpConns: *hashmap.NewSyncMap[int, *Conn[T]](0, 128, hashmap.HashInt),
	}
	return l, nil
}

func (l *Loop[T]) run() {
	if l.lockThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}
	var (
		poller   = l.poll
		timeout  = time.Second
		err      error
		begin    = int64(0)
		tick     = int64(l.tick)
		end      = begin
		nextTick = begin + tick
		acceptFD = l.ln.fd
	)

	_ = end
	_ = nextTick

	onEvent := func(index, count, fd int, filter int16, conn *Conn[T]) error {
		if fd == acceptFD {
			return l.accept(fd)
		}
		return nil
	}

	onLoop := func(count int) (time.Duration, error) {
		end = timex.NanoTime()
		return time.Second, nil
	}

	for {
		begin = timex.NanoTime()
		err = poller.Wait(
			timeout,
			onEvent,
			onLoop,
		)
		end = timex.NanoTime()
		if err != nil {
			return
		}
	}
}

func (l *Loop[T]) accept(fd int) error {
	if l.ln.network == "udp" {
		return l.readUDP(l.ln, fd)
	}
	nfd, sa, err := unix.Accept(fd)
	if err != nil {
		if err == unix.EAGAIN {
			return nil
		}
	}
	err = unix.SetNonblock(nfd, true)
	if err != nil {
		return err
	}

	removeAddr := socket.SockaddrToTCPOrUnixAddr(sa)
	err = socket.SetKeepAlivePeriod(nfd, int(time.Second*3600))
	if err != nil {
		return err
	}
	c := newTCPConn[T](nfd, l, sa, l.ln.addr, removeAddr)

	_ = l.poll.AddRead(c.fd, c)
	l.conns.Store(c.fd, c)
	l.active.Incr()
	c.opened = true

	return nil
}

func (l *Loop[T]) open(c *Conn[T]) error {
	action := l.onOpened(c)

	if len(c.wr) > 0 {
		if err := l.poll.AddReadWrite(c.fd, c); err != nil {
			return err
		}
	}

	return l.handle(c, action)
}

func (l *Loop[T]) handle(c *Conn[T], action Action) error {
	switch action {
	case 0:
		return nil
	case 1:
		return l.closeConn(c, true)
	}
	return nil
}

func (l *Loop[T]) readUDP(ln *Listener, fd int) error {
	n, sa, err := unix.Recvfrom(fd, l.packet, 0)
	if err != nil {
		if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
			return nil
		}
		return fmt.Errorf("failed to read UDP packet from fd=%d in event-loop(%d), %v",
			fd, l.idx, os.NewSyscallError("recvfrom", err))
	}
	_ = n
	_ = sa
	if fd == ln.fd {

	}
	return nil
}

func (l *Loop[T]) read(c *Conn[T]) error {
	if cap(c.rd) == 0 {
		// TODO: extract into callback
		c.rd = make([]byte, 4096)
	} else {
		c.rd = c.rd[:]
	}
	n, err := unix.Read(c.fd, c.rd)
	if err != nil {
		if err == unix.EAGAIN {
			return nil
		}
		return l.closeConn(c, false)
	}
	if n == 0 {
		err = unix.ECONNRESET
		return l.closeConn(c, false)
	}
	c.rd = c.rd[:n]
	return nil
}

func (l *Loop[T]) write(c *Conn[T]) error {
	data, err := l.doWrite(c)
	if err != nil {
		return l.closeConn(c, false)
	}
	if c.fd == 0 {
		return nil
	}
	if len(data) == 0 {
		// All data have been drained, it's no need to monitor the writable events,
		// remove the writable event from poller to help the future event-loops.
		_ = l.poll.ModRead(c.fd, c)
		return nil
	}

	n, err := syscall.Write(c.fd, data)
	if err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		return l.closeConn(c, false)
	}
	if n == len(data) {

	} else {
		data = data[n:]
		err = l.onShortWrite(c, n, data)
		if err != nil {
			return l.closeConn(c, false)
		}
	}
	return nil
}

func (l *Loop[T]) closeConn(c *Conn[T], flush bool) error {
	if flush && len(c.wr) > 0 {
		n, err := unix.Write(c.fd, c.wr)
		if n < len(c.wr) {
			// TODO: Warn
		}
		if err != nil {
			// TODO: Warn
		}
		if n > -1 {
			c.wr = c.wr[n:]
		}
	}
	_ = l.poll.ModDetach(c.fd, c)
	closeErr := unix.Close(c.fd)
	var err error
	if closeErr != nil {
		closeErr = fmt.Errorf("failed to close fd=%d in event-loop(%d): %v", c.fd, l.idx, os.NewSyscallError("close", closeErr))
		if err != nil {
			err = errors.New(err.Error() + " & " + closeErr.Error())
		} else {
			err = closeErr
		}
	}

	l.conns.Delete(c.fd)
	l.active.Decr()
	return nil
}
