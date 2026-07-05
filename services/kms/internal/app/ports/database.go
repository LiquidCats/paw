package ports

import (
	"context"
	"time"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
)

type KeyManagerRepository interface {
	CreateKey(ctx context.Context, entry entities.KeyEntry) error
	GetAllKeys(ctx context.Context) ([]entities.KeyEntry, error)
	GetKey(ctx context.Context, keyID entities.KeyID) (*entities.KeyEntry, error)
	SetExpiration(ctx context.Context, keyID entities.KeyID, expiresAt *time.Time) error
	SetStatus(ctx context.Context, keyID entities.KeyID, status entities.KeyStatus) error
}
