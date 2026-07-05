package ports

import (
	"github.com/LiquidCats/paw/services/kms/internal/app/domain/entities"
)

type Keychain[Pub, Priv any] interface {
	Derive(index entities.Index) (Keychain[Pub, Priv], error)
	PrivateKey() *Priv
	PublicKey() *Pub
	Destroy()
}
