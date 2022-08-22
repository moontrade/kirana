package pool

import (
	"github.com/moontrade/wormhole/pkg/atomicx"
	"strconv"
	"unsafe"
)

type RC[T any] struct {
	count int32
	_     int32
	shard *Shard[RC[T]]
	value T
}

func (rc *RC[T]) Value() T {
	return rc.value
}

func (rc *RC[T]) Clone() *RC[T] {
	count := atomicx.Xaddint32(&rc.count, 1)
	if count < 2 {
		panic("rc clone count is less than 2: " + strconv.Itoa(int(count)))
	}
	return rc
}

func (rc *RC[T]) Release() {
	count := atomicx.Xaddint32(&rc.count, -1)
	if count == 0 {
		s := rc.shard
		if s != nil {
			s.PutUnsafe(unsafe.Pointer(rc))
		}
	} else if count < 0 {
		panic("count is less than 0: " + strconv.Itoa(int(count)))
	}
}
