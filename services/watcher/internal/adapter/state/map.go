package state

import "sync"

type MapState[K comparable, V any] struct {
	internal map[K]V

	mu sync.RWMutex
}

func NewMapState[K comparable, V any](capacity int) *MapState[K, V] {
	return &MapState[K, V]{
		internal: make(map[K]V, capacity),
		mu:       sync.RWMutex{},
	}
}

func (s *MapState[K, V]) Set(key K, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.internal[key] = value
}

func (s *MapState[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.internal[key]
	return val, ok
}

func (s *MapState[K, V]) Del(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.internal, key)
}

func (s *MapState[K, V]) Has(key K) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.internal[key]
	return ok
}
