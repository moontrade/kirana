package cow

import (
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/spinlock"
)

type Map[K comparable, V any] struct {
	data map[K]V
	mu   spinlock.Mutex
}

func NewMap[K comparable, V any](initialSize int) *Map[K, V] {
	if initialSize < 1 {
		initialSize = 2
	}
	initialSize = pmath.CeilToPowerOf2(initialSize)
	return &Map[K, V]{
		data: make(map[K]V, initialSize),
	}
}

func (m *Map[K, V]) Len() int {
	return len(m.data)
}

func (m *Map[K, V]) Iterate(fn func(key K, value V) bool) {
	if fn == nil {
		return
	}
	data := m.data
	if len(data) == 0 {
		return
	}
	for k, v := range data {
		if !fn(k, v) {
			break
		}
	}
}

func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	value, ok = m.data[key]
	return
}

func (m *Map[K, V]) Put(key K, value V) (prev V, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	old := m.data
	next := make(map[K]V, len(old)+1)

	if old != nil {
		for k, v := range old {
			next[k] = v
		}
		prev, ok = old[key]
	}

	next[key] = value
	m.data = next
	return
}

func (m *Map[K, V]) Delete(key K) (prev V, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	old := m.data
	if len(old) == 0 {
		return
	}

	next := make(map[K]V, len(old)-1)

	if old != nil {
		for k, v := range old {
			if k != key {
				prev = v
				ok = true
			} else {
				next[k] = v
			}
		}
	}

	m.data = next
	return
}
