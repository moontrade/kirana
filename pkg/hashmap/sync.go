package hashmap

import (
	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/spinlock"
	"runtime"
)

// Sync is a thread-safe version of Map. It achieves this by sharding into "x" number of shards each
// with a spinlock and an instance of Map. Shards are determined by key hash. The key is only ever hashed
// once per operation. Map code is embedded only to reduce the hash operation count.
type Sync[K comparable, V any] struct {
	shards []shard[K, V]
	mask   uint64
	hasher HasherFunc[K]
}

func NewSync[K comparable, V any](numShards, initialCapacity int, hasher HasherFunc[K]) *Sync[K, V] {
	if hasher == nil {
		return nil
	}
	if numShards < 1 {
		numShards = runtime.GOMAXPROCS(0) * 4
	}
	numShards = pmath.CeilToPowerOf2(numShards)
	shards := make([]shard[K, V], numShards)
	for i := 0; i < len(shards); i++ {
		shards[i] = shard[K, V]{
			m:      New[K, V](initialCapacity, hasher),
			hasher: hasher,
		}
	}
	return &Sync[K, V]{
		shards: shards,
		mask:   uint64(numShards - 1),
		hasher: hasher,
	}
}

func (m *Sync[K, V]) shard(key K) *shard[K, V] {
	return &m.shards[m.hasher(key)&m.mask]
}

// Get is volatile and extremely fast. It's possible to have a small window where it misses.
// When it does miss calling Load
func (m *Sync[K, V]) Get(key K) (val V, ok bool) {
	h := m.hasher(key)
	return m.shards[h&m.mask].Get(h>>dibBitSize, key)
}

func (m *Sync[K, V]) GetOrCreate(key K, supplier func(K) V) (value V, created bool) {
	return m.GetOrLoadCreate(key, supplier)
}

func (m *Sync[K, V]) Load(key K) (val V, ok bool) {
	h := m.hasher(key)
	return m.shards[h&m.mask].Load(h>>dibBitSize, key)
}

func (m *Sync[K, V]) LoadOrCreate(key K, supplier func(K) V) (value V, created bool) {
	h := m.hasher(key)
	return m.shards[h&m.mask].GetOrCreate(h>>dibBitSize, key, supplier)
}

func (mu *Sync[K, V]) GetOrLoad(key K) (val V, ok bool) {
	h := mu.hasher(key)
	shard := &mu.shards[h&mu.mask]
	val, ok = shard.Get(h>>dibBitSize, key)
	if ok {
		return
	}
	return shard.Load(h>>dibBitSize, key)
}

func (mu *Sync[K, V]) GetOrLoadCreate(key K, supplier func(K) V) (val V, created bool) {
	h := mu.hasher(key)
	shard := mu.shard(key)
	var ok bool
	val, ok = shard.Get(h>>dibBitSize, key)
	if ok {
		return
	}
	return shard.GetOrCreate(h>>dibBitSize, key, supplier)
}

func (m *Sync[K, V]) Scan(iter func(key K, value V) bool) {
	for _, s := range m.shards {
		s.Scan(iter)
	}
}

func (m *Sync[K, V]) ScanUnsafe(iter func(key K, value V) bool) {
	for _, s := range m.shards {
		s.ScanUnsafe(iter)
	}
}

func (m *Sync[K, V]) Range(iter func(key K, value V) bool) {
	for _, s := range m.shards {
		s.Scan(iter)
	}
}

func (m *Sync[K, V]) Put(key K, value V) (prev V, ok bool) {
	h := m.hasher(key)
	return m.shards[h&m.mask].Put(h>>dibBitSize, key, value)
}

func (m *Sync[K, V]) Store(key K, value V) (prev V, ok bool) {
	h := m.hasher(key)
	return m.shards[h&m.mask].Put(h>>dibBitSize, key, value)
}

func (m *Sync[K, V]) PutIf(
	key K,
	value V,
	condition func(prev V, prevExists bool) bool,
) (prev V, prevExists, ok bool) {
	h := m.hasher(key)
	return m.shards[h&m.mask].PutIf(h>>dibBitSize, key, value, condition)
}

func (m *Sync[K, V]) PutIfAbsent(key K, value V) (prev V, ok bool) {
	h := m.hasher(key)
	return m.shards[h&m.mask].PutIfAbsent(h>>dibBitSize, key, value)
}

func (m *Sync[K, V]) Delete(key K) (prev V, ok bool) {
	h := m.hasher(key)
	return m.shards[h&m.mask].Delete(h>>dibBitSize, key)
}

func (m *Sync[K, V]) DeleteIf(key K, condition func(existing V) bool) (prev V, ok bool) {
	h := m.hasher(key)
	return m.shards[h&m.mask].DeleteIf(h>>dibBitSize, key, condition)
}

type shard[K comparable, V any] struct {
	p      *Sync[K, V]
	m      *Map[K, V]
	hasher HasherFunc[K]
	mu     spinlock.Mutex
}

func (s *shard[K, V]) Load(hash uint64, key K) (val V, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	//return s.m.Get(key)
	m := s.m
	buckets := m.buckets
	if len(buckets) == 0 {
		return val, false
	}
	mask := uint64(len(buckets) - 1)
	i := hash & mask
	for {
		if buckets[i].dib() == 0 {
			return val, false
		}
		if buckets[i].hash() == hash && buckets[i].key == key {
			return buckets[i].value, true
		}
		i = (i + 1) & mask
	}
}

func (s *shard[K, V]) get(hash uint64, key K) (val V, ok bool) {
	var (
		m       = s.m
		buckets = m.buckets
	)
	if len(buckets) == 0 {
		return val, false
	}
	mask := uint64(len(buckets) - 1)
	i := hash & mask
	for {
		if buckets[i].dib() == 0 {
			return val, false
		}
		if buckets[i].hash() == hash && buckets[i].key == key {
			return buckets[i].value, true
		}
		i = (i + 1) & mask
	}
}

func (s *shard[K, V]) Get(hash uint64, key K) (val V, ok bool) {
	return s.get(hash, key)
}

func (s *shard[K, V]) GetOrCreate(hash uint64, key K, supplier func(K) V) (value V, created bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.m
	//m.loadCow()
	if len(m.buckets) == 0 {
		*m = *New[K, V](0, s.hasher)
	}
	if m.length >= m.growAt {
		m.resize(len(m.buckets) * 2)
	}

	e := entry[K, V]{}
	e.hdib = makeHDIB(hash, 1)
	e.key = key
	i := e.hash() & m.mask
	for {
		if m.buckets[i].dib() == 0 {
			value = supplier(key)
			e.value = value
			m.buckets[i] = e
			m.length++
			return value, true
		}
		if e.hash() == m.buckets[i].hash() && e.key == m.buckets[i].key {
			value = m.buckets[i].value
			return value, false
		}
		if m.buckets[i].dib() < e.dib() {
			e, m.buckets[i] = m.buckets[i], e
		}
		i = (i + 1) & m.mask
		e.setDIB(e.dib() + 1)
	}
}

func (s *shard[K, V]) Put(hash uint64, key K, value V) (prev V, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.m
	//m.loadCow()
	if len(m.buckets) == 0 {
		*m = *New[K, V](0, s.hasher)
	}
	if m.length >= m.growAt {
		m.resize(len(m.buckets) * 2)
	}
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

func (s *shard[K, V]) PutIfAbsent(hash uint64, key K, value V) (prev V, exists bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.m
	//m.loadCow()
	if len(m.buckets) == 0 {
		*m = *New[K, V](0, s.hasher)
	}
	if m.length >= m.growAt {
		m.resize(len(m.buckets) * 2)
	}
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
			return prev, true
		}
		if m.buckets[i].dib() < e.dib() {
			e, m.buckets[i] = m.buckets[i], e
		}
		i = (i + 1) & m.mask
		e.setDIB(e.dib() + 1)
	}
}

func (s *shard[K, V]) PutIf(
	hash uint64,
	key K, value V,
	condition func(existing V, existingExists bool) bool,
) (prev V, prevExists, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.m
	//m.loadCow()
	if len(m.buckets) == 0 {
		*m = *New[K, V](0, s.hasher)
	}
	if m.length >= m.growAt {
		m.resize(len(m.buckets) * 2)
	}
	e := entry[K, V]{makeHDIB(hash, 1), value, key}
	i := e.hash() & m.mask
	for {
		if m.buckets[i].dib() == 0 {
			if !condition(prev, false) {
				return
			}
			m.buckets[i] = e
			m.length++
			return prev, false, true
		}
		if e.hash() == m.buckets[i].hash() && e.key == m.buckets[i].key {
			prev = m.buckets[i].value
			if !condition(prev, true) {
				return prev, true, false
			}
			m.buckets[i].value = e.value
			return prev, true, true
		}
		if m.buckets[i].dib() < e.dib() {
			e, m.buckets[i] = m.buckets[i], e
		}
		i = (i + 1) & m.mask
		e.setDIB(e.dib() + 1)
	}
}

func (s *shard[K, V]) Delete(hash uint64, key K) (prev V, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.m
	//m.loadCow()
	if len(m.buckets) == 0 {
		return prev, false
	}
	i := hash & m.mask
	for {
		if m.buckets[i].dib() == 0 {
			return prev, false
		}
		if m.buckets[i].hash() == hash && m.buckets[i].key == key {
			prev = m.buckets[i].value
			m.remove(i)
			return prev, true
		}
		i = (i + 1) & m.mask
	}
}

func (s *shard[K, V]) DeleteIf(hash uint64, key K, condition func(existing V) bool) (prev V, deleted bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.m
	//m.loadCow()
	if len(m.buckets) == 0 {
		return
	}
	i := hash & m.mask
	for {
		if m.buckets[i].dib() == 0 {
			return prev, false
		}
		if m.buckets[i].hash() == hash && m.buckets[i].key == key {
			prev = m.buckets[i].value
			if !condition(prev) {
				return prev, false
			}
			m.remove(i)
			return prev, true
		}
		i = (i + 1) & m.mask
	}
}

// Scan iterate through all entries under the lock. Only do this when necessary.
func (s *shard[K, V]) Scan(iter func(key K, value V) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m.Scan(iter)
}

// ScanUnsafe iterates through all entries with no lock. It's possible to skip some entries
// and/or see deleted entries, etc. If a resize/rehash happens during, the scan will be on
// the old table
func (s *shard[K, V]) ScanUnsafe(iter func(key K, value V) bool) {
	s.m.Scan(iter)
}
