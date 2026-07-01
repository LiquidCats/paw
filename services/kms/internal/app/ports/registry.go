package ports

type RegistryGetter[K comparable, V any] interface {
	Get(key K) (V, error)
	All() []V
}

type RegistryRegister[K comparable, V any] interface {
	Register(key K, value V)
}
