package hsm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	domainErr "github.com/LiquidCats/paw/services/litehsm/internal/app/domain/errors"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/ports"
	"github.com/jackc/pgx/v5"
)

type KeyManagerSetExpiration struct {
	db ports.KeyManagerRepository
}

func NewKeyManagerSetExpiration(db ports.KeyManagerRepository) *KeyManagerSetExpiration {
	return &KeyManagerSetExpiration{db: db}
}

func (uc *KeyManagerSetExpiration) Handle(ctx context.Context, keyID entities.KeyID, expiration time.Time) error {
	err := uc.validate(keyID, expiration)
	if err != nil {
		return fmt.Errorf("struct=KeyManagerSetExpiration, method=Handle, call=validate: %w", err)
	}
	err = uc.db.SetExpiration(ctx, keyID, &expiration)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainErr.ErrKeyNotFound
		}
		return fmt.Errorf("struct=KeyManagerSetExpiration, method=Handle, call=db.SetExpiration: %w", err)
	}
	return nil
}

func (uc *KeyManagerSetExpiration) validate(keyID entities.KeyID, expiration time.Time) error {
	if keyID == (entities.KeyID{}) {
		return domainErr.NewValidationError("keyID is required")
	}
	if expiration.IsZero() {
		return domainErr.NewValidationError("expiration is required")
	}

	if !expiration.After(time.Now()) {
		return domainErr.NewValidationError("expiration date cannot be in the past")
	}

	return nil
}
