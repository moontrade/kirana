package logger

import (
	"errors"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/spinlock"
	"github.com/moontrade/kirana/pkg/timex"
	"golang.org/x/sys/cpu"
	"io"
	"math"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

func init() {

}

var (
	ErrFull      = errors.New("full")
	ErrTimeout   = errors.New("timeout")
	ErrEmptyData = errors.New("empty data")
	ErrTooBig    = errors.New("too big")
)

const CacheLinePad = unsafe.Sizeof(cpu.CacheLinePad{})

type writer struct {
	shards []*writerShard
	mask   int
	queue  *writerQueue
}

func newWriter(size int, shardSize int64) *writer {
	if size <= 0 {
		size = runtime.GOMAXPROCS(0)
	}
	size = pmath.CeilToPowerOf2(size)
	shards := make([]*writerShard, size)
	queue := newWriterQueue(int64(size * 4))
	for i := 0; i < len(shards); i++ {
		shards[i] = newWriterShard(i, shardSize, queue)
	}
	w := &writer{
		shards: shards,
		mask:   size - 1,
		queue:  queue,
	}
	go w.read()
	return w
}

func (w *writer) shard(gp *g) *writerShard {
	return w.shards[int(gp.m.id)&w.mask]
}

func (w *writer) read() {
	var (
		tm    = time.NewTimer(time.Second)
		n     int
		err   error
		wr    = io.Discard
		buf   = make([]byte, len(w.shards[0].buf))
		queue = w.queue
		mask  = w.mask
	)
	_ = mask

	read := func(index int64) {
		if index < 0 {
			return
		}
		n, err = w.shards[index].read(buf)
		if n > 0 {
			_, _ = wr.Write(buf[:n])
		}
		if err != nil {
		}
	}

	readAll := func() bool {
		size := 0
		for i := range w.shards {
			n, err = w.shards[i].read(buf)
			if n > 0 {
				_, _ = wr.Write(buf[:n])
				size += n
			}
			if err != nil {
			}
		}
		return size > 0
	}

	count := 0
	_ = count
	for {
		if queue.DequeueMany(math.MaxUint32, read) == 0 {
			if readAll() {
				continue
			}
			runtime.Gosched()
			tm.Reset(time.Second)
			select {
			case _, ok := <-queue.waker:
				if !ok {
					return
				}
			case <-tm.C:
				readAll()
			}
		} else {
			count++
			if count%100 == 0 {
				readAll()
			}
		}
	}

	//for {
	//	size = 0
	//	for _, shard := range w.shards {
	//		n, err = shard.read(buf)
	//		size += n
	//		if n > 0 {
	//			_, _ = wr.Write(buf[:n])
	//		}
	//		if err != nil {
	//		}
	//	}
	//
	//	if size > 0 {
	//		continue
	//	}
	//
	//Loop:
	//	for {
	//		select {
	//		case index, ok := <-w.waker:
	//			if !ok {
	//				return
	//			}
	//			n, err = w.shards[index].read(buf)
	//			size += n
	//			if n > 0 {
	//				_, _ = wr.Write(buf[:n])
	//			}
	//			if err != nil {
	//			}
	//		default:
	//			break Loop
	//		}
	//	}
	//
	//	tm.Reset(time.Second * 5)
	//	select {
	//	case index, ok := <-w.waker:
	//		if !ok {
	//			return
	//		}
	//		n, err = w.shards[index].read(buf)
	//		size += n
	//		if n > 0 {
	//			_, _ = wr.Write(buf[:n])
	//		}
	//		if err != nil {
	//		}
	//	case <-tm.C:
	//	}
	//}
}

type writerShard struct {
	index int64
	ridx  int64
	_     [CacheLinePad - 16]byte
	widx  int64
	_     [CacheLinePad - 8]byte
	cidx  int64
	_     [CacheLinePad - 8]byte
	waked int64
	_     [CacheLinePad - 8]byte
	mask  int64
	buf   []byte
	_     [CacheLinePad - unsafe.Sizeof(reflect.SliceHeader{}) - 32]byte
	waker *writerQueue
	wakes int64
	spins int64
	mu    spinlock.Mutex
}

// New returns the RingBuffer object
func newWriterShard(index int, capacity int64, waker *writerQueue) *writerShard {
	if capacity < 64 {
		capacity = 64
	}
	capacity = int64(pmath.CeilToPowerOf2(int(capacity)))
	return &writerShard{
		index: int64(index),
		buf:   make([]byte, capacity),
		mask:  capacity - 1,
		waker: waker,
	}
}

func (b *writerShard) read(buf []byte) (int, error) {
	atomic.StoreInt64(&b.waked, 0)
	if len(buf) == 0 {
		return 0, nil
	}
	var (
		r = atomic.LoadInt64(&b.ridx)
		c = atomic.LoadInt64(&b.cidx)
	)
	if r == c {
		return 0, nil
	}
	n := int(c - r)
	if n > len(buf) {
		n = len(buf)
	}
	// do copy
	index := r & b.mask
	if int(index)+n > len(b.buf) {
		copy(buf, b.buf[index:])
		remainder := len(b.buf) - int(index)
		copy(buf[remainder:n], b.buf)
	} else {
		copy(buf[:n], b.buf[index:])
	}
	atomic.StoreInt64(&b.ridx, r+int64(n))
	return n, nil
}

//	func (b *writerShard) write(buf []byte) (int, error) {
//		return b.writeTimeout(buf, time.Second)
//	}
//
// func (b *writerShard) writeTimeout(buf []byte, timeout time.Duration) (int, error) {
func (b *writerShard) write0(buf []byte) (int, error) {
	// Is it empty?
	if len(buf) == 0 {
		return 0, ErrEmptyData
	}
	// Is it bigger than the entire ring buffer?
	if len(buf) > len(b.buf) {
		return 0, ErrTooBig
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	var (
		end = atomic.LoadInt64(&b.cidx) + int64(len(buf))
		i   = end - int64(len(buf))
		r   = atomic.LoadInt64(&b.ridx)
		//notify = r == i
	)

	// Wait until space is available
	spins := 1
	backoff := 1
	for int64(len(b.buf))-(i-r) < int64(len(buf)) {
		for i := 0; i < backoff; i++ {
			runtime.Gosched()
		}
		if backoff < maxBackoff {
			backoff <<= 1
		} else if spins%maxBackoff == 0 {
			time.Sleep(time.Microsecond * 50)
		}
		r = atomic.LoadInt64(&b.ridx)
		atomic.AddInt64(&b.spins, 1)
		spins++
	}

	// Copy
	index := i & b.mask
	to := b.buf[index:]
	if len(to) < len(buf) {
		copy(to, buf)
		remainder := len(buf) - len(to)
		copy(b.buf[0:remainder], buf[len(to):])
	} else {
		copy(to, buf)
	}
	atomic.StoreInt64(&b.cidx, end)

	if b.waker != nil &&
		atomic.LoadInt64(&b.ridx) == i &&
		atomic.CompareAndSwapInt64(&b.waked, 0, 1) {
		for !b.waker.Enqueue(b.index) {
			runtime.Gosched()
		}
		//select {
		//case b.waker <- b.index:
		//	b.wakes++
		//default:
		//}
	}

	return len(buf), nil
}

const maxBackoff = 16

func (b *writerShard) write(buf []byte) (int, error) {
	// Is it empty?
	if len(buf) == 0 {
		return 0, ErrEmptyData
	}
	// Is it bigger than the entire ring buffer?
	if len(buf) > len(b.buf) {
		return 0, ErrTooBig
	}
	var (
		end   = atomic.AddInt64(&b.widx, int64(len(buf)))
		begin = end - int64(len(buf))
		r     = atomic.LoadInt64(&b.ridx)
		//notify = r == begin
	)
	// Wait until space is available
	spins := 1
	backoff := 1
	for int64(len(b.buf))-(begin-r) < int64(len(buf)) {
		for i := 0; i < backoff; i++ {
			runtime.Gosched()
		}
		if backoff < maxBackoff {
			backoff <<= 1
		} else if spins%maxBackoff == 0 {
			time.Sleep(time.Microsecond * 50)
		}
		r = atomic.LoadInt64(&b.ridx)
		atomic.AddInt64(&b.spins, 1)
		spins++
	}
	// Copy
	index := begin & b.mask
	to := b.buf[index:]
	if len(to) < len(buf) {
		copy(to, buf)
		remainder := len(buf) - len(to)
		copy(b.buf[0:remainder], buf[len(to):])
	} else {
		copy(to, buf)
	}
	// Writes must be ordered by waiting (spinning) for previous write to complete
	c := atomic.LoadInt64(&b.cidx)
	for c != begin {
		runtime.Gosched()
		c = atomic.LoadInt64(&b.cidx)
	}
	atomic.StoreInt64(&b.cidx, end)

	if b.waker != nil &&
		atomic.LoadInt64(&b.ridx) == begin &&
		atomic.CompareAndSwapInt64(&b.waked, 0, 1) {
		atomic.AddInt64(&b.wakes, 1)
		b.waker.Enqueue(b.index)
		//for !b.waker.Enqueue(b.index) {
		//	runtime.Gosched()
		//}
		//b.waker.Enqueue(b.index)
		//select {
		//case b.waker <- b.index:
		//	atomic.AddInt64(&b.wakes, 1)
		//default:
		//}
	}

	return len(buf), nil
}

const (
	writerQueueKill = int64(-1)
	writerQueueNil  = int64(-2)
)

type writerQueueNode struct {
	seq  int64
	data int64
	//_    [CacheLinePad - 16]byte
}

// writerQueue implements a circular buffer.
type writerQueue struct {
	head          int64
	_             [CacheLinePad - 8]byte
	tail          int64
	_             [CacheLinePad - 8]byte
	nodes         []writerQueueNode
	mask          int64
	_             [CacheLinePad - unsafe.Sizeof(reflect.SliceHeader{}) - 32]byte
	wake          int64
	_             [CacheLinePad - 8]byte
	waker         chan int64
	wakeCount     counter.Counter
	wakeFull      counter.Counter
	overflowCount counter.Counter
}

// newWriterQueue returns the RingBuffer object
func newWriterQueue(capacity int64) *writerQueue {
	if capacity < 4 {
		capacity = 4
	}
	capacity = int64(pmath.CeilToPowerOf2(int(capacity)))
	nodes := make([]writerQueueNode, capacity)
	for i := 0; i < len(nodes); i++ {
		//n := &nodes[i]
		atomic.StoreInt64(&nodes[i].seq, int64(i))
	}
	return &writerQueue{
		head:  0,
		tail:  0,
		mask:  capacity - 1,
		nodes: nodes,
		waker: make(chan int64, 1),
	}
}

func (b *writerQueue) Len() int {
	return int(atomic.LoadInt64(&b.tail) - atomic.LoadInt64(&b.head))
}

func (b *writerQueue) Cap() int {
	return len(b.nodes)
}

func (b *writerQueue) IsFull() bool {
	return atomic.LoadInt64(&b.tail)-atomic.LoadInt64(&b.head) >= (b.mask)
}

func (b *writerQueue) IsEmpty() bool {
	return atomic.LoadInt64(&b.tail)-atomic.LoadInt64(&b.head) == 0
}

func (b *writerQueue) Processed() int {
	return int(atomic.LoadInt64(&b.tail))
}

func (b *writerQueue) EnqueueTimeout(data int64, timeout time.Duration) bool {
	if b.Enqueue(data) {
		return true
	}
	var (
		begin = timex.NanoTime()
		count = 0
	)
	runtime.Gosched()
	for {
		if b.Enqueue(data) {
			return true
		}
		if timex.NanoTime()-begin >= int64(timeout) {
			return false
		}
		count++
		if count%10 == 0 {
			time.Sleep(time.Microsecond * 50)
		} else {
			runtime.Gosched()
		}
	}
}

func (b *writerQueue) Enqueue(data int64) bool {
	if data == writerQueueNil {
		return false
	}
	var (
		cell *writerQueueNode
		pos  = atomic.LoadInt64(&b.tail)
	)
	for {
		cell = &b.nodes[pos&b.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - pos
		if diff == 0 {
			//if atomicx.Casint64(&b.tail, pos, pos+1) {
			if atomic.CompareAndSwapInt64(&b.tail, pos, pos+1) {
				break
			}
			pos = atomic.LoadInt64(&b.tail)
		} else if diff < 0 {
			if b.IsFull() {
				if b.waker != nil && pos-atomic.LoadInt64(&b.head) == 0 {
					//if b.waker != nil && atomic.LoadInt64(&b.wake) == 0 {
					if atomic.CompareAndSwapInt64(&b.wake, 0, 1) {
						b.wakeCount.Incr()
						select {
						case b.waker <- 0:
						default:
							b.wakeFull.Incr()
						}
					}
				}
				return false
			}
			runtime.Gosched()
		} else {
			pos = atomic.LoadInt64(&b.tail)
		}
	}

	atomic.StoreInt64(&cell.data, data)
	atomic.StoreInt64(&cell.seq, pos+1)

	if b.waker != nil && pos-atomic.LoadInt64(&b.head) == 0 {
		//if b.waker != nil && atomic.LoadInt64(&b.wake) == 0 {
		if atomic.CompareAndSwapInt64(&b.wake, 0, 1) {
			b.wakeCount.Incr()
			select {
			case b.waker <- 0:
			default:
				b.wakeFull.Incr()
			}
		}
	}

	return true
}

func (b *writerQueue) Dequeue() int64 {
	atomic.StoreInt64(&b.wake, 0)
	var (
		cell *writerQueueNode
		pos  = atomic.LoadInt64(&b.head)
	)
	for {
		cell = &b.nodes[pos&b.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - (pos + 1)
		if diff == 0 {
			if atomic.CompareAndSwapInt64(&b.head, pos, pos+1) {
				break
			}
		} else if diff < 0 {
			if b.IsEmpty() {
				return writerQueueNil
			}
			runtime.Gosched()
		} else {
			pos = atomic.LoadInt64(&b.head)
		}
	}

	data := atomic.SwapInt64(&cell.data, writerQueueNil)
	for data == writerQueueNil {
		runtime.Gosched()
		data = atomic.SwapInt64(&cell.data, writerQueueNil)
	}
	atomic.StoreInt64(&cell.seq, pos+b.mask+1)
	return data
}

func (b *writerQueue) DequeueMany(maxCount int, consumer func(int64)) (count int) {
	atomic.StoreInt64(&b.wake, 0)
	var (
		cell *writerQueueNode
		pos  = atomic.LoadInt64(&b.head)
	)
	for {
		cell = &b.nodes[pos&b.mask]
		seq := atomic.LoadInt64(&cell.seq)
		diff := seq - (pos + 1)
		if diff == 0 {
			if atomic.CompareAndSwapInt64(&b.head, pos, pos+1) {
				data := atomic.SwapInt64(&cell.data, writerQueueNil)
				for data == writerQueueNil {
					runtime.Gosched()
					data = atomic.SwapInt64(&cell.data, writerQueueNil)
				}
				atomic.StoreInt64(&cell.seq, pos+b.mask+1)
				consumer(data)
				count++
				if count >= maxCount {
					return
				}
			}
		} else if diff < 0 {
			if b.IsEmpty() {
				return
			}
			runtime.Gosched()
		} else {
			pos = atomic.LoadInt64(&b.head)
		}
	}
}
