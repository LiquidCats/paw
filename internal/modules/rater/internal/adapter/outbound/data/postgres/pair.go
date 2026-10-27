package postgres

import (
	"context"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	"github.com/rotisserie/eris"
)

func (r *Repository) GetAllPairs(ctx context.Context) ([]entities.Pair, error) {
	items, err := r.db.GetAllPairs(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get all pairs")
	}

	pairs := make([]entities.Pair, 0, len(items))
	for _, item := range items {
		pairs = append(pairs, entities.Pair{
			From:   values.CurrencyISO(item.BaseAsset),
			To:     values.CurrencyISO(item.QuoteAsset),
			Symbol: values.Symbol(item.Symbol),
		})
	}

	return pairs, nil
}

func (r *Repository) GetPair(ctx context.Context, symbol values.Symbol) (*entities.Pair, error) {
	item, err := r.db.GetPair(ctx, symbol.String())
	if err != nil {
		return nil, eris.Wrap(err, "failed to get pair")
	}

	return &entities.Pair{
		From:   values.CurrencyISO(item.BaseAsset),
		To:     values.CurrencyISO(item.QuoteAsset),
		Symbol: values.Symbol(item.Symbol),
	}, nil
}

func (r *Repository) CountPairs(ctx context.Context) (int64, error) {
	count, err := r.db.CountPairs(ctx)
	if err != nil {
		return 0, eris.Wrap(err, "failed to count pairs")
	}

	return count, nil
}

func (r *Repository) HasPair(ctx context.Context, symbol values.Symbol) (bool, error) {
	has, err := r.db.HasPair(ctx, string(symbol))
	if err != nil {
		return false, eris.Wrap(err, "failed to check pair")
	}

	return has, nil
}
