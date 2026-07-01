package ports

import (
	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
)

type Keychain[Pub, Priv any] interface {
	Derive(index entities.Index) (Keychain[Pub, Priv], error)
	PrivateKey() *Priv
	PublicKey() *Pub
	Destroy()
}
