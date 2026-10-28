package postgres

import (
	"context"

	"github.com/rotisserie/eris"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

func (r *Repository) GetAllProviders(ctx context.Context) ([]entities.Provider, error) {
	items, err := r.db.GetAllProviders(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get all providers")
	}

	var providers []entities.Provider
	for _, item := range items {
		providers = append(providers, entities.Provider{
			Name: values.ProviderName(item.Name),
		})
	}

	return providers, nil
}
