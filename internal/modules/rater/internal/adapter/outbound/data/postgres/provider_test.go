package postgres_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres/db"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

func TestRepository_GetAllProviders(t *testing.T) {
	t.Parallel()

	t.Run("maps rows", func(t *testing.T) {
		t.Parallel()
		repo := postgres.NewRepository(db.New(fakeDBTX{
			queryFn: func(context.Context, string, ...any) (pgx.Rows, error) {
				return &fakeRows{rows: [][]any{
					{"cex", pgtypeTimestamp()},
					{"coingate", pgtypeTimestamp()},
				}}, nil
			},
		}))

		got, err := repo.GetAllProviders(context.Background())

		require.NoError(t, err)
		assert.Equal(t, []entities.Provider{
			{Name: values.ProviderNameCex},
			{Name: values.ProviderNameCoinGate},
		}, got)
	})

	t.Run("wraps query error", func(t *testing.T) {
		t.Parallel()
		repo := postgres.NewRepository(db.New(fakeDBTX{
			queryFn: func(context.Context, string, ...any) (pgx.Rows, error) { return nil, rateErr },
		}))

		got, err := repo.GetAllProviders(context.Background())

		assert.Nil(t, got)
		assert.ErrorIs(t, err, rateErr)
	})
}
