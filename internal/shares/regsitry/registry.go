package regsitry

import "github.com/LiquidCats/paw/internal/kernel/ports"

var _ ports.Registry[any, any] = (*Registry[any, any])(nil)

type Registry[K comparable, V any] struct {
	items map[K]V
}

func NewRegistry[K comparable, V any]() *Registry[K, V] {
	return &Registry[K, V]{
		items: make(map[K]V),
	}
}

// Add implements [ports.Registry].
func (r *Registry[K, V]) Add(key K, value V) {
	r.items[key] = value
}

// Clear implements [ports.Registry].
func (r *Registry[K, V]) Clear() {
	r.items = make(map[K]V)
}

// Delete implements [ports.Registry].
func (r *Registry[K, V]) Delete(key K) {
	delete(r.items, key)
}

// Get implements [ports.Registry].
func (r *Registry[K, V]) Get(key K) (V, bool) {
	value, ok := r.items[key]
	return value, ok
}

// Iterate implements [ports.Registry].
func (r *Registry[K, V]) Iterate(f func(key K, value V)) {
	for key, value := range r.items {
		f(key, value)
	}
}

// Len implements [ports.Registry].
func (r *Registry[K, V]) Len() int {
	return len(r.items)
}
