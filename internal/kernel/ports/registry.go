package ports

type Registry[K comparable, V any] interface {
	Add(key K, value V)
	Get(key K) (V, bool)
	Delete(key K)
	Len() int
	Clear()
	Iterate(f func(key K, value V))
}
