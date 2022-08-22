package cow

import (
	"github.com/moontrade/wormhole/pkg/spinlock"
)

type Slice[E any] struct {
	data []E
	null E
	mu   spinlock.Mutex
}

func NewSlice[E any]() *Slice[E] {
	return &Slice[E]{}
}

func NewSliceOf[E any](initial []E) *Slice[E] {
	return &Slice[E]{data: initial}
}

func (s *Slice[E]) Len() int {
	return len(s.data)
}

func (s *Slice[E]) Snapshot() []E {
	return s.data
}

func (s *Slice[E]) Get(index int) E {
	return s.data[index]
}

func (s *Slice[E]) ReplaceWith(data []E) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = data
}

func (s *Slice[E]) Take() []E {
	s.mu.Lock()
	defer s.mu.Unlock()
	data := s.data
	s.data = nil
	return data
}

func (s *Slice[E]) PeekOr(or E) E {
	data := s.data
	if len(data) == 0 {
		return or
	}
	return data[0]
}

func (s *Slice[E]) PopOr(or E) E {
	data := s.data
	if len(data) == 0 {
		return or
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	data = s.data
	if len(data) == 0 {
		return or
	}
	if len(data) == 1 {
		s.data = nil
		return data[0]
	}
	next := make([]E, len(data)-1)
	copy(next, data[1:])
	return data[0]
}

func (s *Slice[E]) Clone() []E {
	data := s.data
	clone := make([]E, len(data))
	clone = append(clone, data...)
	return clone
}

func (s *Slice[E]) Append(element E) *Slice[E] {
	s.mu.Lock()
	defer s.mu.Unlock()
	data := s.data
	next := make([]E, 0, len(data)+1)
	if len(data) > 0 {
		next = append(next, data...)
	}
	next = append(next, element)
	s.data = next
	return s
}

func (s *Slice[E]) AppendIndex(element E) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	data := s.data
	next := make([]E, 0, len(s.data)+1)
	if len(data) > 0 {
		next = append(next, data...)
	}
	next = append(next, element)
	s.data = next
	return len(next) - 1
}

func (s *Slice[E]) Remove(fn func(elem E) bool) (count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data := s.data
	if len(data) == 0 {
		return
	}

	var next []E = nil
	if len(data) > 1 {
		next = make([]E, 0, len(data)-1)
	}

	for _, elem := range data {
		if !fn(elem) {
			next = append(next, elem)
		} else {
			count++
		}
	}
	s.data = next
	return
}

func (s *Slice[E]) RemoveAt(index int) (elem E, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data := s.data
	if len(data) <= index {
		return
	}
	elem = data[index]
	ok = true

	if len(data) == 1 {
		s.data = nil
		return
	}

	next := make([]E, len(data)-1)
	if index == 0 {
		copy(next, data[1:])
	} else if index == len(next) {
		copy(next, data)
	} else {
		copy(next, data[0:index])
		copy(next[index:], data[index+1:])
	}
	return
}

func (s *Slice[E]) Iterate(fn func(element E) bool) {
	data := s.data
	for _, el := range data {
		if !fn(el) {
			return
		}
	}
}
