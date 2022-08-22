package swap

import (
	"github.com/moontrade/wormhole/pkg/spinlock"
)

type SyncSlice[T comparable] struct {
	slots  []T
	null   T
	getter func(T) int
	setter func(T, int)
	mu     spinlock.Mutex
}

func NewSync[T comparable](getter func(T) int, setter func(T, int)) *SyncSlice[T] {
	if getter == nil {
		panic("getter is nil")
	}
	if setter == nil {
		panic("setter is nil")
	}
	return &SyncSlice[T]{
		getter: getter,
		setter: setter,
	}
}

func (s *SyncSlice[T]) Len() int {
	return len(s.slots)
}

func (s *SyncSlice[T]) Add(value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setter(value, len(s.slots))
	s.slots = append(s.slots, value)
}

func (s *SyncSlice[T]) Remove(value T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if value == s.null {
		return false
	}
	index := s.getter(value)
	if index < 0 || index >= len(s.slots) {
		return false
	}

	if len(s.slots) == 1 {
		s.slots[0] = s.null
		s.slots = s.slots[:0]
		return true
	}

	if s.slots[index] != value {
		return false
	}

	tailIndex := len(s.slots) - 1
	tail := s.slots[tailIndex]
	s.slots[tailIndex] = s.null
	s.slots = s.slots[0:tailIndex]
	s.setter(tail, index)
	return true
}

func (s *SyncSlice[T]) Get(index int) (value T, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 || index >= len(s.slots) {
		return
	}
	value = s.slots[index]
	ok = true
	return
}

func (s *SyncSlice[T]) Iterate(fn func(T) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.slots) == 0 || fn == nil {
		return
	}
	for _, s := range s.slots {
		if !fn(s) {
			return
		}
	}
}

func (s *SyncSlice[T]) CopyTo(to []T) []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.slots) > len(to) {
		to = make([]T, len(s.slots))
	}
	copy(to, s.slots)
	return to
}

func (s *SyncSlice[T]) Unsafe() []T {
	return s.slots
}
