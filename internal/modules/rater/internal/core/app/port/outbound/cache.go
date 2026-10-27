package outbound

import (
	"context"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
)

type RateCacheRepository interface {
	GetRate(ctx context.Context, pair entities.Pair) (*entities.Rate, error)
	PutRate(ctx context.Context, rate entities.Rate) error
}
