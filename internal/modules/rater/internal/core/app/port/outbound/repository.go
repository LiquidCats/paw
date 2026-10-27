package outbound

import (
	"context"
	"time"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

type RateDatabaseRepository interface {
	GetRate(ctx context.Context, symbol values.Symbol, ts time.Time) (*entities.Rate, error)
	HasRate(ctx context.Context, symbol values.Symbol, ts time.Time) (bool, error)
	SaveRate(ctx context.Context, rate entities.Rate, ts time.Time) (*entities.Rate, error)
}

type PairDatabaseRepository interface {
	GetAllPairs(ctx context.Context) ([]entities.Pair, error)
	GetPair(ctx context.Context, symbol values.Symbol) (*entities.Pair, error)
	CountPairs(ctx context.Context) (int64, error)
	HasPair(ctx context.Context, symbol values.Symbol) (bool, error)
}

type ProviderDatabaseRepository interface {
	GetAllProviders(ctx context.Context) ([]entities.Provider, error)
}
