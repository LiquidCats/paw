package sealer

import (
	"errors"
	"fmt"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	"github.com/awnumar/memguard"
	"golang.org/x/crypto/chacha20poly1305"
)

var (
	ErrEmptyPassphrase = errors.New("empty passphrase")
	ErrEmptyData       = errors.New("empty data")

	ErrUnexpectedWrappedDelSize = errors.New("unexpected wrapped dek length")

	ErrFailedUnseal = errors.New("unseal failed")

	ErrOpenPassphrase   = errors.New("open passphrase")
	ErrInvalidKDFParams = errors.New("invalid kdf params")
)

type Sealing struct {
	magic   string
	version entities.Version
	kdf     entities.KDF
	aead    entities.AEAD
	params  entities.KDFParams
}

func NewDefault(magic string) *Sealing {
	return &Sealing{
		magic:   magic,
		version: entities.DefaultVersion,
		kdf:     entities.KDFArgon2id,
		aead:    entities.AEADxchacha20poly1305,
		params:  entities.DefaultKDFParams,
	}
}

func (s *Sealing) Seal(passphrase *memguard.Enclave, data *memguard.LockedBuffer) (*entities.Envelope, error) {
	var err error

	if passphrase == nil || passphrase.Size() == 0 {
		return nil, ErrEmptyPassphrase
	}

	if data == nil || data.Size() == 0 {
		return nil, ErrEmptyData
	}

	passphraseBuf, err := passphrase.Open()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpenPassphrase, err)
	}
	defer passphraseBuf.Destroy()

	if err = s.params.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidKDFParams, err)
	}

	envelope := &entities.Envelope{
		Version:   s.version,
		KDF:       s.kdf,
		AEAD:      s.aead,
		KDFParams: s.params,
	}

	copy(envelope.Magic[:], s.magic)

	// Random salt + two nonces (KEK wrap, DEK seal)

	envelope.Salt, err = randomBytes(entities.SaltSize)
	if err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	envelope.NonceKEK, err = randomBytes(entities.NonceSize)
	if err != nil {
		return nil, fmt.Errorf("generate KEK nonce: %w", err)
	}

	envelope.NonceDEK, err = randomBytes(entities.NonceSize)
	if err != nil {
		return nil, fmt.Errorf("generate DEL nonce: %w", err)
	}

	kek := deriveKey(passphraseBuf, envelope.Salt, s.params)
	defer kek.Destroy()

	dek, err := randomBytes(entities.DEKSize)
	if err != nil {
		return nil, fmt.Errorf("generate DEK: %w", err)
	}
	defer dek.Destroy()

	kekAEAD, err := chacha20poly1305.NewX(kek.Bytes())
	if err != nil {
		return nil, fmt.Errorf("create KEK AEAD: %w", err)
	}

	wrapped := kekAEAD.Seal(nil, envelope.NonceKEK.Bytes(), dek.Bytes(), envelope.AAD())
	if len(wrapped) != entities.WrappedDEKSize {
		return nil, ErrUnexpectedWrappedDelSize
	}

	envelope.WrappedDEK = memguard.NewBufferFromBytes(wrapped)

	// Seal the plaintext under the DEK
	dekAEAD, err := chacha20poly1305.NewX(dek.Bytes())
	if err != nil {
		return nil, fmt.Errorf("dek aead: %w", err)
	}

	ciphertext := dekAEAD.Seal(nil, envelope.NonceDEK.Bytes(), data.Bytes(), envelope.AAD())

	envelope.Ciphertext = memguard.NewBufferFromBytes(ciphertext)

	return envelope, nil
}

func (s *Sealing) Unseal(env *entities.Envelope, passphrase *memguard.Enclave) (*memguard.Enclave, error) {
	passphraseBuf, err := passphrase.Open()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpenPassphrase, err)
	}

	defer passphraseBuf.Destroy()

	if err = env.KDFParams.Validate(); err != nil {
		return nil, fmt.Errorf("invalid kdf params: %w", err)
	}

	kek := deriveKey(passphraseBuf, env.Salt, env.KDFParams)
	defer kek.Destroy()

	kekAEAD, err := chacha20poly1305.NewX(kek.Bytes())
	if err != nil {
		return nil, fmt.Errorf("kek aead: %w", err)
	}

	dekRaw, err := kekAEAD.Open(nil, env.NonceKEK.Bytes(), env.WrappedDEK.Bytes(), env.AAD())
	if err != nil {
		// Don't leak whether the passphrase was wrong vs corrupted file
		return nil, ErrFailedUnseal
	}

	dek := memguard.NewBufferFromBytes(dekRaw)
	defer dek.Destroy()

	dekAEAD, err := chacha20poly1305.NewX(dek.Bytes())
	if err != nil {
		return nil, fmt.Errorf("dek aead: %w", err)
	}

	plain, err := dekAEAD.Open(nil, env.NonceDEK.Bytes(), env.Ciphertext.Bytes(), env.AAD())
	if err != nil {
		return nil, ErrFailedUnseal
	}

	return memguard.NewEnclave(plain), nil
}
