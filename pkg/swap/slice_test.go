package swap

import "testing"

func BenchmarkSwap(b *testing.B) {
	b.Run("Generic", func(b *testing.B) {
		s := New[*Item](itemSwapIndex, itemSetSwapIndex)
		item := &Item{}
		s.Add(item)
		item2 := &Item{}
		s.Add(item2)
		s.Remove(item2)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.Add(item2)
			s.Remove(item2)
		}
	})

	b.Run("Not Generic", func(b *testing.B) {
		s := new(SwapItem)
		item := &Item{}
		s.Add(item)
		item2 := &Item{}
		s.Add(item2)
		s.Remove(item2)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.Add(item2)
			s.Remove(item2)
		}
	})
}

type Item struct {
	index int
}

func itemSwapIndex(item *Item) int           { return item.index }
func itemSetSwapIndex(item *Item, index int) { item.index = index }

func (i *Item) SwapIndex() int {
	return i.index
}

func (i *Item) SetSwapIndex(index int) {
	i.index = index
}

type SwapItem struct {
	slots []*Item
	size  int
}

func (s *SwapItem) Add(value *Item) {
	value.index = len(s.slots)
	s.slots = append(s.slots, value)
}

func (s *SwapItem) Remove(value *Item) bool {
	if value == nil {
		return false
	}
	index := value.index
	if index < 0 || index >= len(s.slots) {
		return false
	}

	if len(s.slots) == 1 {
		s.slots[0] = nil
		s.slots = s.slots[:0]
		return true
	}

	if s.slots[index] != value {
		return false
	}

	tailIndex := len(s.slots) - 1
	tail := s.slots[tailIndex]
	s.slots[tailIndex] = nil
	s.slots = s.slots[0:tailIndex]
	tail.index = index
	return true
}

func (s *SwapItem) Get(index int) (value *Item, ok bool) {
	if index < 0 || index >= len(s.slots) {
		return
	}
	value = s.slots[index]
	ok = true
	return
}

func (s *SwapItem) Iterate(fn func(*Item) bool) {
	if len(s.slots) == 0 || fn == nil {
		return
	}
	for _, s := range s.slots {
		if !fn(s) {
			return
		}
	}
}
