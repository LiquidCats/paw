package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres/db"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type RateRepositorySuite struct {
	suite.Suite
}

func TestRateRepositorySuite(t *testing.T) {
	suite.Run(t, new(RateRepositorySuite))
}

func (s *RateRepositorySuite) TestGetRateReturnsMappedRate() {
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(context.Context, string, ...any) pgx.Row {
			return fakeRow{values: []any{
				int32(1), decimal.NewFromInt(100), "BTC_USD", "cex", pgtypeTimestamp(), pgtypeTimestamp(),
			}}
		},
	}))

	rate, err := repo.GetRate(context.Background(), values.Symbol("BTC_USD"), time.Now())

	s.Require().NoError(err)
	s.Equal(values.ProviderNameCex, rate.Provider)
	s.True(rate.Price.Equal(decimal.NewFromInt(100)))
	s.Equal(entities.Pair{From: "BTC", To: "USD", Symbol: "BTC_USD"}, rate.Pair)
}

func (s *RateRepositorySuite) TestGetRateWrapsQueryError() {
	wantErr := errors.New("no rows")
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(context.Context, string, ...any) pgx.Row {
			return fakeRow{err: wantErr}
		},
	}))

	rate, err := repo.GetRate(context.Background(), values.Symbol("BTC_USD"), time.Now())

	s.Require().ErrorIs(err, wantErr)
	s.Nil(rate)
}

func (s *RateRepositorySuite) TestGetRateWrapsUnknownProviderName() {
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(context.Context, string, ...any) pgx.Row {
			return fakeRow{values: []any{
				int32(1), decimal.NewFromInt(100), "BTC_USD", "not-a-real-provider", pgtypeTimestamp(), pgtypeTimestamp(),
			}}
		},
	}))

	rate, err := repo.GetRate(context.Background(), values.Symbol("BTC_USD"), time.Now())

	s.Require().ErrorIs(err, values.ErrInvalidProviderName)
	s.Nil(rate)
}

func (s *RateRepositorySuite) TestHasRateReturnsBool() {
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(context.Context, string, ...any) pgx.Row {
			return fakeRow{values: []any{true}}
		},
	}))

	has, err := repo.HasRate(context.Background(), values.Symbol("BTC_USD"), time.Now())

	s.Require().NoError(err)
	s.True(has)
}

func (s *RateRepositorySuite) TestHasRateWrapsQueryError() {
	wantErr := errors.New("connection refused")
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(context.Context, string, ...any) pgx.Row {
			return fakeRow{err: wantErr}
		},
	}))

	has, err := repo.HasRate(context.Background(), values.Symbol("BTC_USD"), time.Now())

	s.Require().ErrorIs(err, wantErr)
	s.False(has)
}

func (s *RateRepositorySuite) TestSaveRateStoresProviderNameAsText() {
	var gotArgs []any
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(_ context.Context, _ string, args ...any) pgx.Row {
			gotArgs = args
			return fakeRow{values: []any{
				int32(1), decimal.NewFromInt(100), "BTC_USD", "cex", pgtypeTimestamp(), pgtypeTimestamp(),
			}}
		},
	}))
	rate := entities.Rate{
		Pair:     entities.Pair{Symbol: "BTC_USD"},
		Price:    decimal.NewFromInt(100),
		Provider: values.ProviderNameCex,
	}

	got, err := repo.SaveRate(context.Background(), rate, time.Now())

	s.Require().NoError(err)
	s.Equal(values.ProviderNameCex, got.Provider)
	s.Require().Len(gotArgs, 4)
	s.Equal("cex", gotArgs[2])
}

func (s *RateRepositorySuite) TestSaveRateWrapsQueryError() {
	wantErr := errors.New("constraint violation")
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(context.Context, string, ...any) pgx.Row {
			return fakeRow{err: wantErr}
		},
	}))
	rate := entities.Rate{Pair: entities.Pair{Symbol: "BTC_USD"}, Price: decimal.NewFromInt(100), Provider: values.ProviderNameCex}

	got, err := repo.SaveRate(context.Background(), rate, time.Now())

	s.Require().ErrorIs(err, wantErr)
	s.Nil(got)
}
