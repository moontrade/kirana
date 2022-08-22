package swap

type Slice[T comparable] struct {
	slots  []T
	null   T
	getter func(T) int
	setter func(T, int)
}

func New[T comparable](getter func(T) int, setter func(T, int)) *Slice[T] {
	if getter == nil {
		panic("getter is nil")
	}
	if setter == nil {
		panic("setter is nil")
	}
	return &Slice[T]{
		getter: getter,
		setter: setter,
	}
}

func (s *Slice[T]) Len() int {
	return len(s.slots)
}

func (s *Slice[T]) Add(value T) {
	s.setter(value, len(s.slots))
	s.slots = append(s.slots, value)
}

func (s *Slice[T]) Remove(value T) bool {
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

func (s *Slice[T]) Get(index int) (value T, ok bool) {
	if index < 0 || index >= len(s.slots) {
		return
	}
	value = s.slots[index]
	ok = true
	return
}

func (s *Slice[T]) Iterate(fn func(T) bool) {
	if len(s.slots) == 0 || fn == nil {
		return
	}
	for _, s := range s.slots {
		if !fn(s) {
			return
		}
	}
}
