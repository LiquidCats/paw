package ports

import (
	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	"github.com/awnumar/memguard"
)

type Sealer interface {
	Seal(passphrase *memguard.Enclave, data *memguard.LockedBuffer) (*entities.Envelope, error)
}

type Unsealer interface {
	Unseal(envelope *entities.Envelope, passphrase *memguard.Enclave) (*memguard.Enclave, error)
}
