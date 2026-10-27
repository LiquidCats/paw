package inbound

import (
	"context"
	"time"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

type RateService interface {
	Historical(ctx context.Context, pair entities.Pair, ts time.Time) (*entities.Rate, error)
	Current(ctx context.Context, pair entities.Pair) (*entities.Rate, values.ProviderName, error)
}
