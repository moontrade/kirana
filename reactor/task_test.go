package reactor

import (
	"fmt"
	"testing"
)

func TestTaskList(t *testing.T) {
	q := newTaskList(128)
	q.alloc(&Task{}, false)
	q.alloc(&Task{}, false)
	q.alloc(&Task{}, false)
	q.clear(1)
	q.clear(1)
	fmt.Println(q.size)
}

func BenchmarkTaskList(b *testing.B) {
	removeFromMiddle := func(size int) {
		b.Run(fmt.Sprintf("Remove From Middle %d", size), func(b *testing.B) {
			q := newTaskList(size)
			for i := 0; i < size; i++ {
				q.alloc(&Task{}, false)
			}
			task := &Task{}
			middle := size / 2
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				q.clear(middle)
				q.alloc(task, false)
			}
		})
	}
	removeFromEnd := func(size int) {
		b.Run(fmt.Sprintf("Remove From End %d", size), func(b *testing.B) {
			q := newTaskList(size)
			for i := 0; i < size; i++ {
				q.alloc(&Task{}, false)
			}
			task := &Task{}
			end := size - 1
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				q.clear(end)
				q.alloc(task, false)
			}
		})
	}

	removeFromMiddle(16)
	removeFromEnd(16)
	removeFromMiddle(64)
	removeFromEnd(64)
	removeFromMiddle(128)
	removeFromEnd(128)
	removeFromMiddle(256)
	removeFromEnd(256)
	removeFromMiddle(512)
	removeFromEnd(512)
	removeFromMiddle(1024)
	removeFromEnd(1024)
	removeFromMiddle(4096)
	removeFromEnd(4096)
	removeFromMiddle(32768)
	removeFromEnd(32768)
}
