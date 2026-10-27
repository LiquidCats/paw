package outbound

import (
	"context"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/shopspring/decimal"
)

type RateAPIRepository interface {
	GetRate(ctx context.Context, pair entities.Pair) (decimal.Decimal, error)
}
