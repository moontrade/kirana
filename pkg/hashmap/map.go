// Copyright 2019 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

// This has been slightly modified to use wyhashF3 instead of xxHash64.
// wyhashF3 is faster and has high quality similar to xxHash64.

package hashmap

const (
	loadFactor  = 0.85                      // must be above 50%
	dibBitSize  = 16                        // 0xFFFF
	hashBitSize = 64 - dibBitSize           // 0xFFFFFFFFFFFF
	maxHash     = ^uint64(0) >> dibBitSize  // max 28,147,497,671,0655
	maxDIB      = ^uint64(0) >> hashBitSize // max 65,535
)

type entry[K comparable, V any] struct {
	hdib  uint64 // bitfield { hash:48 dib:16 }
	value V      // user value
	key   K      // user key
}

func (e *entry[K, V]) dib() int {
	return int(e.hdib & maxDIB)
}
func (e *entry[K, V]) hash() uint64 {
	return e.hdib >> dibBitSize
}
func (e *entry[K, V]) setDIB(dib int) {
	e.hdib = e.hdib>>dibBitSize<<dibBitSize | uint64(dib)&maxDIB
}
func (e *entry[K, V]) setHash(hash uint64) {
	e.hdib = hash<<dibBitSize | e.hdib&maxDIB
}
func makeHDIB(hash, dib uint64) uint64 {
	return hash<<dibBitSize | dib&maxDIB
}

// hash returns a 48-bit hash for 64-bit environments, or 32-bit hash for
// 32-bit environments.
func (m *Map[K, V]) hash(key K) uint64 {
	return m.hasher(key) >> dibBitSize
}

// Map is a hashmap. Like map[string]interface{}
type Map[K comparable, V any] struct {
	cap      int
	length   int
	mask     uint64
	growAt   int
	shrinkAt int
	buckets  []entry[K, V]
	cow      [2]*cow
	hasher   HasherFunc[K]
}

type cow struct {
	_ int // cannot be an empty struct
}

// New returns a new Map. Like map[string]interface{}
func New[K comparable, V any](cap int, hasher HasherFunc[K]) *Map[K, V] {
	m := new(Map[K, V])
	m.cap = cap
	sz := 8
	for sz < m.cap {
		sz *= 2
	}
	m.hasher = hasher
	m.buckets = make([]entry[K, V], sz)
	m.mask = uint64(len(m.buckets) - 1)
	m.growAt = int(float64(len(m.buckets)) * loadFactor)
	m.shrinkAt = int(float64(len(m.buckets)) * (1 - loadFactor))
	return m
}

func (m *Map[K, V]) resize(newCap int) {
	nmap := New[K, V](newCap, m.hasher)
	for i := 0; i < len(m.buckets); i++ {
		if m.buckets[i].dib() > 0 {
			nmap.set(m.buckets[i].hash(), m.buckets[i].key, m.buckets[i].value)
		}
	}
	c := m.cap
	*m = *nmap
	m.cap = c
}

// Set assigns a value to a key.
// Returns the previous value, or false when no value was assigned.
func (m *Map[K, V]) Set(key K, value V) (V, bool) {
	m.loadCow()
	if len(m.buckets) == 0 {
		*m = *New[K, V](0, m.hasher)
	}
	if m.length >= m.growAt {
		m.resize(len(m.buckets) * 2)
	}
	return m.set(m.hash(key), key, value)
}

func (m *Map[K, V]) set(hash uint64, key K, value V) (prev V, ok bool) {
	e := entry[K, V]{makeHDIB(hash, 1), value, key}
	i := e.hash() & m.mask
	for {
		if m.buckets[i].dib() == 0 {
			m.buckets[i] = e
			m.length++
			return prev, false
		}
		if e.hash() == m.buckets[i].hash() && e.key == m.buckets[i].key {
			prev = m.buckets[i].value
			m.buckets[i].value = e.value
			return prev, true
		}
		if m.buckets[i].dib() < e.dib() {
			e, m.buckets[i] = m.buckets[i], e
		}
		i = (i + 1) & m.mask
		e.setDIB(e.dib() + 1)
	}
}

// Get returns a value for a key.
// Returns false when no value has been assign for key.
func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	buckets := m.buckets
	if len(buckets) == 0 {
		return value, false
	}
	mask := uint64(len(buckets) - 1)
	h := m.hash(key)
	i := h & mask
	for {
		if buckets[i].dib() == 0 {
			return value, false
		}
		if buckets[i].hash() == h && buckets[i].key == key {
			return buckets[i].value, true
		}
		i = (i + 1) & mask
	}
}

// Len returns the number of values in map.
func (m *Map[K, V]) Len() int {
	return m.length
}

func (m *Map[K, V]) loadCow() {
	if m.cow[1] != m.cow[0] {
		// copy-on-write
		buckets := make([]entry[K, V], len(m.buckets))
		copy(buckets, m.buckets)
		m.buckets = buckets
		m.cow[1] = m.cow[0]
	}
}

// Delete deletes a value for a key.
// Returns the deleted value, or false when no value was assigned.
func (m *Map[K, V]) Delete(key K) (prev V, deleted bool) {
	m.loadCow()
	if len(m.buckets) == 0 {
		return prev, false
	}
	h := m.hash(key)
	i := h & m.mask
	for {
		if m.buckets[i].dib() == 0 {
			return prev, false
		}
		if m.buckets[i].hash() == h && m.buckets[i].key == key {
			prev = m.buckets[i].value
			m.remove(i)
			return prev, true
		}
		i = (i + 1) & m.mask
	}
}

func (m *Map[K, V]) remove(i uint64) {
	m.buckets[i].setDIB(0)
	for {
		pi := i
		i = (i + 1) & m.mask
		if m.buckets[i].dib() <= 1 {
			m.buckets[pi] = entry[K, V]{}
			break
		}
		m.buckets[pi] = m.buckets[i]
		m.buckets[pi].setDIB(m.buckets[pi].dib() - 1)
	}
	m.length--
	if len(m.buckets) > m.cap && m.length <= m.shrinkAt {
		m.resize(m.length)
	}
}

// Scan iterates over all key/values.
// It's not safe to call or Set or Delete while scanning.
func (m *Map[K, V]) Scan(iter func(key K, value V) bool) {
	buckets := m.buckets
	for i := 0; i < len(buckets); i++ {
		if buckets[i].dib() > 0 {
			if !iter(buckets[i].key, buckets[i].value) {
				return
			}
		}
	}
}

// Keys returns all keys as a slice
func (m *Map[K, V]) Keys() []K {
	keys := make([]K, 0, m.length)
	for i := 0; i < len(m.buckets); i++ {
		if m.buckets[i].dib() > 0 {
			keys = append(keys, m.buckets[i].key)
		}
	}
	return keys
}

// Values returns all values as a slice
func (m *Map[K, V]) Values() []V {
	values := make([]V, 0, m.length)
	for i := 0; i < len(m.buckets); i++ {
		if m.buckets[i].dib() > 0 {
			values = append(values, m.buckets[i].value)
		}
	}
	return values
}

// Copy the smapet. This is a copy-on-write operation and is very fast because
// it only performs a shadow copy.
func (m *Map[K, V]) Copy() *Map[K, V] {
	m2 := new(Map[K, V])
	*m2 = *m
	m2.cow[0] = new(cow)
	return m2
}
