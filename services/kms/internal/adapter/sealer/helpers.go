package sealer

import (
	"crypto/rand"
	"io"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	"github.com/awnumar/memguard"
	"golang.org/x/crypto/argon2"
)

func deriveKey(
	passphrase *memguard.LockedBuffer,
	salt *memguard.LockedBuffer,
	params entities.KDFParams,
) *memguard.LockedBuffer {
	return memguard.NewBufferFromBytes(
		argon2.IDKey(
			passphrase.Bytes(),
			salt.Bytes(),
			params.Iterations,
			params.MemoryKiB,
			params.Parallelism,
			entities.DEKSize,
		),
	)
}

func randomBytes(size int) (*memguard.LockedBuffer, error) {
	b := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, err
	}

	return memguard.NewBufferFromBytes(b), nil
}
