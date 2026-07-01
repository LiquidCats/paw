package state

type SliceState[T any] interface {
	Set(value T)
	Get() []T
}

type MapState[K comparable, V any] interface {
	Set(key K, value V)
	Get(key K) (V, bool)
	Del(key K)
	Has(key K) bool
}
