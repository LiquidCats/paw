package hsm

import (
	"context"
	"fmt"
	"time"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	domainErr "github.com/LiquidCats/paw/services/litehsm/internal/app/domain/errors"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/ports"
)

type KeyManagerCreateKey struct {
	db ports.KeyManagerRepository
}

func NewKeyManagerCreateKey(db ports.KeyManagerRepository) *KeyManagerCreateKey {
	return &KeyManagerCreateKey{
		db: db,
	}
}

func (uc *KeyManagerCreateKey) Handle(ctx context.Context, entry entities.KeyEntry) (*entities.KeyEntry, error) {
	if err := uc.validate(&entry); err != nil {
		return nil, fmt.Errorf("struct=KeyManagerCreateKey, method=Handle, call=validate: %w", err)
	}

	idx1 := entities.RandomHardenedKeyIndex()
	idx2 := entities.RandomHardenedKeyIndex()
	idx3 := entities.NewIndex(0, false)

	dp := entities.DerivationPath{idx1, idx2, idx3}

	entry.DerivationPath = dp
	entry.Status = entities.KeyStatusDisabled

	err := uc.db.CreateKey(ctx, entry)
	if err != nil {
		return nil, fmt.Errorf("struct=KeyManagerCreateKey, method=Handle, call=db.CreateKey: %w", err)
	}

	return &entry, nil
}

func (uc *KeyManagerCreateKey) validate(entry *entities.KeyEntry) error {
	if entry.DerivationPath != nil {
		return domainErr.ErrDerivationPathCannotBeSet
	}

	if entry.KeyID != (entities.KeyID{}) {
		return domainErr.ErrKeyIDCannotBeSet
	}

	if len(entry.SeedFingerprint) != 0 {
		return domainErr.ErrSeedFingerprintCannotBeSet
	}

	if len(entry.Alias) == 0 {
		return domainErr.ErrAliasCannotBeEmpty
	}

	if len(entry.Alias) < 3 {
		return domainErr.ErrAliasCannotBeLessThan3Chars
	}

	if len(entry.Status) > 0 {
		return domainErr.ErrStatusCannotBeSet
	}

	if len(entry.Alias) > 250 {
		return domainErr.ErrAliasCannotBeLongerThan250Chars
	}

	if entry.ExpiresAt != nil {
		if !entry.ExpiresAt.After(time.Now()) {
			return domainErr.ErrExpirationDateCannotBeInThePast
		}
	}

	return nil
}
