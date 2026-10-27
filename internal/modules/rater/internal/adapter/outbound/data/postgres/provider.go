package postgres

import (
	"context"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	"github.com/rotisserie/eris"
)

func (r *Repository) GetAllProviders(ctx context.Context) ([]entities.Provider, error) {
	items, err := r.db.GetAllProviders(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get all providers")
	}

	var providers []entities.Provider
	for _, item := range items {
		var name values.ProviderName

		err = name.UnmarshalText([]byte(item.Name))
		if err != nil {
			return nil, eris.Wrap(err, "failed to unmarshal provider name")
		}

		providers = append(providers, entities.Provider{
			Name: name,
		})
	}

	return providers, nil
}
