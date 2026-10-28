package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rotisserie/eris"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres/db"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

func (r *Repository) GetRate(ctx context.Context, symbol values.Symbol, ts time.Time) (*entities.Rate, error) {
	item, err := r.db.GetRate(ctx, db.GetRateParams{
		Pair: string(symbol),
		Ts: pgtype.Timestamp{
			Time:             ts,
			InfinityModifier: 0,
			Valid:            true,
		},
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to get rate")
	}

	rate := &entities.Rate{
		Pair:     entities.NewPairFromSymbol(values.Symbol(item.Pair)),
		Price:    item.Price,
		Provider: values.ProviderName(item.Provider),
	}

	return rate, nil
}

func (r *Repository) HasRate(ctx context.Context, symbol values.Symbol, ts time.Time) (bool, error) {
	has, err := r.db.HasRate(ctx, db.HasRateParams{
		Pair: string(symbol),
		Ts:   pgtype.Timestamp{Time: ts, InfinityModifier: 0, Valid: true},
	})
	if err != nil {
		return false, eris.Wrap(err, "failed to check rate")
	}

	return has, nil
}

func (r *Repository) SaveRate(ctx context.Context, rate entities.Rate, ts time.Time) (*entities.Rate, error) {
	item, err := r.db.SaveRate(ctx, db.SaveRateParams{
		Pair:     string(rate.Pair.Symbol),
		Price:    rate.Price,
		Provider: rate.Provider.String(),
		Ts:       pgtype.Timestamp{Time: ts, InfinityModifier: 0, Valid: true},
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to save rate")
	}

	rate = entities.Rate{
		Pair:     entities.NewPairFromSymbol(values.Symbol(item.Pair)),
		Price:    item.Price,
		Provider: values.ProviderName(item.Provider),
	}

	return &rate, nil
}
