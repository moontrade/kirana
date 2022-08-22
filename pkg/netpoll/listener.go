package netpoll

import (
	reuseport "github.com/kavu/go_reuseport"
	"github.com/moontrade/wormhole/pkg/socket"
	"net"
	"os"
	"sync"
	"syscall"
)

type Listener struct {
	ln       net.Listener
	lnaddr   net.Addr
	pconn    net.PacketConn
	opts     addrOpts
	f        *os.File
	fd       int
	network  string
	address  string
	once     sync.Once
	addr     net.Addr
	sockOpts []socket.Option
}

func OpenTCPListener(reusePort bool, addr string) (*Listener, error) {
	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	l, err := net.ListenTCP("tcp", a)
	if err != nil {
		return nil, err
	}
	f, err := l.File()
	if err != nil {
		return nil, err
	}
	ln := &Listener{
		ln:      l,
		lnaddr:  l.Addr(),
		fd:      int(f.Fd()),
		f:       f,
		network: "tcp",
		addr:    a,
	}
	return ln, nil
}

func (ln *Listener) close() {
	if ln.fd != 0 {
		_ = syscall.Close(ln.fd)
	}
	if ln.f != nil {
		_ = ln.f.Close()
	}
	if ln.ln != nil {
		_ = ln.ln.Close()
	}
	if ln.pconn != nil {
		_ = ln.pconn.Close()
	}
	if ln.network == "unix" {
		_ = os.RemoveAll(ln.address)
	}
}

// system takes the net listener and detaches it from its parent
// event loop, grabs the file descriptor, and makes it non-blocking.
func (ln *Listener) system() error {
	var err error
	switch netln := ln.ln.(type) {
	case nil:
		switch pconn := ln.pconn.(type) {
		case *net.UDPConn:
			ln.f, err = pconn.File()
		}
	case *net.TCPListener:
		ln.f, err = netln.File()
	case *net.UnixListener:
		ln.f, err = netln.File()
	}
	if err != nil {
		ln.close()
		return err
	}
	ln.fd = int(ln.f.Fd())
	return syscall.SetNonblock(ln.fd, true)
}

func reuseportListenPacket(proto, addr string) (l net.PacketConn, err error) {
	return reuseport.ListenPacket(proto, addr)
}

func reuseportListen(proto, addr string) (l net.Listener, err error) {
	return reuseport.Listen(proto, addr)
}

type addrOpts struct {
	reusePort bool
}
