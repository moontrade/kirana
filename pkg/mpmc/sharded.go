package mpmc

import (
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/runtimex"
	"unsafe"
)

type Sharded[T any] struct {
	shards []Bounded[T]
	mask   int
}

func NewSharded[T any](numShards int, queueSize int64) *Sharded[T] {
	if numShards < 2 {
		numShards = 2
	}
	numShards = pmath.CeilToPowerOf2(numShards)
	shards := make([]Bounded[T], numShards)
	for i := 0; i < len(shards); i++ {
		shards[i] = *NewBounded[T](queueSize)
	}
	return &Sharded[T]{
		shards: shards,
		mask:   len(shards) - 1,
	}
}

func (nr *Sharded[T]) Shard() *Bounded[T] {
	hash := runtimex.Pid()
	return &nr.shards[hash&nr.mask]
}

func (nr *Sharded[T]) Push(data *T) bool {
	return nr.Shard().Enqueue(data)
}

func (nr *Sharded[T]) PushUnsafe(data unsafe.Pointer) bool {
	return nr.Shard().EnqueueUnsafe(data)
}

func (nr *Sharded[T]) Pop() *T {
	return nr.Shard().Dequeue()
}

func (nr *Sharded[T]) PopDeref() T {
	return nr.Shard().DequeueDeref()
}

func (nr *Sharded[T]) PopMany(max int, fn func(*T)) int {
	return nr.Shard().DequeueMany(max, fn)
}

func (nr *Sharded[T]) PopManyDeref(max int, fn func(T)) int {
	return nr.Shard().DequeueManyDeref(max, fn)
}
