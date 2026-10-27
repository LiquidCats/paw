package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres/db"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/suite"
)

type ProviderRepositorySuite struct {
	suite.Suite
}

func TestProviderRepositorySuite(t *testing.T) {
	suite.Run(t, new(ProviderRepositorySuite))
}

func (s *ProviderRepositorySuite) TestGetAllProvidersReturnsMappedProviders() {
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryFn: func(context.Context, string, ...any) (pgx.Rows, error) {
			return &fakeRows{rows: [][]any{
				{"cex", pgtypeTimestamp()},
				{"coingecko", pgtypeTimestamp()},
			}}, nil
		},
	}))

	providers, err := repo.GetAllProviders(context.Background())

	s.Require().NoError(err)
	s.Equal([]entities.Provider{
		{Name: values.ProviderNameCex},
		{Name: values.ProviderNameCoinGecko},
	}, providers)
}

func (s *ProviderRepositorySuite) TestGetAllProvidersWrapsQueryError() {
	wantErr := errors.New("connection refused")
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryFn: func(context.Context, string, ...any) (pgx.Rows, error) {
			return nil, wantErr
		},
	}))

	providers, err := repo.GetAllProviders(context.Background())

	s.Require().ErrorIs(err, wantErr)
	s.Nil(providers)
}

func (s *ProviderRepositorySuite) TestGetAllProvidersWrapsUnknownProviderName() {
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryFn: func(context.Context, string, ...any) (pgx.Rows, error) {
			return &fakeRows{rows: [][]any{
				{"not-a-real-provider", pgtypeTimestamp()},
			}}, nil
		},
	}))

	providers, err := repo.GetAllProviders(context.Background())

	s.Require().ErrorIs(err, values.ErrInvalidProviderName)
	s.Nil(providers)
}
