package netpoll

type Server[T any] struct {
	listeners []*Listener
	loops     []*Loop[T]
}

//func NewServer[T any](numLoops int, addrs ...string) (*Server[T], error) {
//	listeners, err := OpenListeners(true, addrs...)
//	if err != nil {
//		return nil, err
//	}
//	if numLoops <= 0 {
//		numLoops = runtime.GOMAXPROCS(0)
//	}
//	loops := make([]*Loop[T], numLoops, numLoops)
//	for i := 0; i < len(loops); i++ {
//		loop := &Loop[T]{
//			idx:    i,
//			poll:   OpenPoll[T](),
//			packet: make([]byte, 0xFFFF),
//			conns:  make(map[int]*T),
//		}
//		loops[i] = loop
//		for _, ln := range listeners {
//			loop.poll.AddRead(ln.fd, nil)
//		}
//	}
//	server := &Server[T]{
//		listeners: listeners,
//		loops:     loops,
//	}
//	return server, nil
//}
