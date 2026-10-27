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

type PairRepositorySuite struct {
	suite.Suite
}

func TestPairRepositorySuite(t *testing.T) {
	suite.Run(t, new(PairRepositorySuite))
}

func (s *PairRepositorySuite) TestGetAllPairsReturnsMappedPairs() {
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryFn: func(context.Context, string, ...any) (pgx.Rows, error) {
			return &fakeRows{rows: [][]any{
				{"BTC_USD", "BTC", "USD", pgtypeTimestamp()},
			}}, nil
		},
	}))

	pairs, err := repo.GetAllPairs(context.Background())

	s.Require().NoError(err)
	s.Require().Len(pairs, 1)
	s.Equal(entities.Pair{From: "BTC", To: "USD", Symbol: "BTC_USD"}, pairs[0])
}

func (s *PairRepositorySuite) TestGetAllPairsWrapsQueryError() {
	wantErr := errors.New("connection refused")
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryFn: func(context.Context, string, ...any) (pgx.Rows, error) {
			return nil, wantErr
		},
	}))

	pairs, err := repo.GetAllPairs(context.Background())

	s.Require().ErrorIs(err, wantErr)
	s.Nil(pairs)
}

func (s *PairRepositorySuite) TestGetPairReturnsMappedPair() {
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(context.Context, string, ...any) pgx.Row {
			return fakeRow{values: []any{"BTC_USD", "BTC", "USD", pgtypeTimestamp()}}
		},
	}))

	pair, err := repo.GetPair(context.Background(), values.Symbol("BTC_USD"))

	s.Require().NoError(err)
	s.Equal(&entities.Pair{From: "BTC", To: "USD", Symbol: "BTC_USD"}, pair)
}

func (s *PairRepositorySuite) TestGetPairWrapsQueryError() {
	wantErr := errors.New("not found")
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(context.Context, string, ...any) pgx.Row {
			return fakeRow{err: wantErr}
		},
	}))

	pair, err := repo.GetPair(context.Background(), values.Symbol("BTC_USD"))

	s.Require().ErrorIs(err, wantErr)
	s.Nil(pair)
}

func (s *PairRepositorySuite) TestCountPairsReturnsCount() {
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(context.Context, string, ...any) pgx.Row {
			return fakeRow{values: []any{int64(3)}}
		},
	}))

	count, err := repo.CountPairs(context.Background())

	s.Require().NoError(err)
	s.Equal(int64(3), count)
}

func (s *PairRepositorySuite) TestCountPairsWrapsQueryError() {
	wantErr := errors.New("connection refused")
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(context.Context, string, ...any) pgx.Row {
			return fakeRow{err: wantErr}
		},
	}))

	count, err := repo.CountPairs(context.Background())

	s.Require().ErrorIs(err, wantErr)
	s.Zero(count)
}

func (s *PairRepositorySuite) TestHasPairReturnsBool() {
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(context.Context, string, ...any) pgx.Row {
			return fakeRow{values: []any{true}}
		},
	}))

	has, err := repo.HasPair(context.Background(), values.Symbol("BTC_USD"))

	s.Require().NoError(err)
	s.True(has)
}

func (s *PairRepositorySuite) TestHasPairWrapsQueryError() {
	wantErr := errors.New("connection refused")
	repo := postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(context.Context, string, ...any) pgx.Row {
			return fakeRow{err: wantErr}
		},
	}))

	has, err := repo.HasPair(context.Background(), values.Symbol("BTC_USD"))

	s.Require().ErrorIs(err, wantErr)
	s.False(has)
}
