package hsm

import (
	"context"
	"errors"
	"fmt"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	domainErr "github.com/LiquidCats/paw/services/litehsm/internal/app/domain/errors"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/ports"
	"github.com/jackc/pgx/v5"
)

type KeyManagerSetStatus struct {
	db ports.KeyManagerRepository
}

func NewKeyManagerSetStatus(db ports.KeyManagerRepository) *KeyManagerSetStatus {
	return &KeyManagerSetStatus{db: db}
}

func (uc *KeyManagerSetStatus) Handle(ctx context.Context, keyID entities.KeyID, status entities.KeyStatus) error {
	if err := uc.validate(keyID, status); err != nil {
		return fmt.Errorf("struct=KeyManagerSetStatus, method=Handle, call=validate: %w", err)
	}

	if err := uc.db.SetStatus(ctx, keyID, status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainErr.ErrKeyNotFound
		}
		return fmt.Errorf("struct=KeyManagerSetStatus, method=Handle, call=db.SetStatus: %w", err)
	}

	return nil
}

func (uc *KeyManagerSetStatus) validate(keyID entities.KeyID, status entities.KeyStatus) error {
	if keyID == (entities.KeyID{}) {
		return domainErr.ErrKeyIsRequired
	}

	if status == "" {
		return domainErr.ErrStatusIsRequired
	}

	return nil
}
