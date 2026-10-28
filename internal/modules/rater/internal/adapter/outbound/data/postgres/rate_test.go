package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres/db"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

var (
	rateTS   = time.Date(2026, 1, 2, 3, 5, 0, 0, time.UTC)
	rateErr  = eris.New("db down")
	rateWant = &entities.Rate{
		Pair:     entities.NewPairFromSymbol("BTC_USD"),
		Price:    decimal.NewFromInt(42),
		Provider: values.ProviderNameCex,
	}
)

// rateRow matches the column order of the rates queries.
func rateRow() []any {
	return []any{int32(1), decimal.NewFromInt(42), "BTC_USD", "cex", pgtypeTimestamp(), pgtypeTimestamp()}
}

func rowDB(row fakeRow, capture func(args []any)) *postgres.Repository {
	return postgres.NewRepository(db.New(fakeDBTX{
		queryRowFn: func(_ context.Context, _ string, args ...any) pgx.Row {
			if capture != nil {
				capture(args)
			}

			return row
		},
	}))
}

func TestRepository_GetRate(t *testing.T) {
	t.Parallel()

	t.Run("maps row and passes pair and timestamp", func(t *testing.T) {
		t.Parallel()
		var args []any
		repo := rowDB(fakeRow{values: rateRow()}, func(a []any) { args = a })

		got, err := repo.GetRate(context.Background(), "BTC_USD", rateTS)

		require.NoError(t, err)
		assert.Equal(t, rateWant, got)
		assert.Equal(t, []any{pgtype.Timestamp{Time: rateTS, Valid: true}, "BTC_USD"}, args)
	})

	t.Run("wraps query error", func(t *testing.T) {
		t.Parallel()

		got, err := rowDB(fakeRow{err: rateErr}, nil).GetRate(context.Background(), "BTC_USD", rateTS)

		assert.Nil(t, got)
		assert.ErrorIs(t, err, rateErr)
	})
}

func TestRepository_HasRate(t *testing.T) {
	t.Parallel()

	for _, exists := range []bool{true, false} {
		t.Run("exists="+map[bool]string{true: "true", false: "false"}[exists], func(t *testing.T) {
			t.Parallel()

			got, err := rowDB(fakeRow{values: []any{exists}}, nil).HasRate(context.Background(), "BTC_USD", rateTS)

			require.NoError(t, err)
			assert.Equal(t, exists, got)
		})
	}

	t.Run("wraps query error", func(t *testing.T) {
		t.Parallel()

		got, err := rowDB(fakeRow{err: rateErr}, nil).HasRate(context.Background(), "BTC_USD", rateTS)

		assert.False(t, got)
		assert.ErrorIs(t, err, rateErr)
	})
}

func TestRepository_SaveRate(t *testing.T) {
	t.Parallel()

	t.Run("returns stored rate and passes all columns", func(t *testing.T) {
		t.Parallel()
		var args []any
		repo := rowDB(fakeRow{values: rateRow()}, func(a []any) { args = a })

		got, err := repo.SaveRate(context.Background(), *rateWant, rateTS)

		require.NoError(t, err)
		assert.Equal(t, rateWant, got)
		assert.Equal(t, []any{decimal.NewFromInt(42), "BTC_USD", "cex", pgtype.Timestamp{Time: rateTS, Valid: true}}, args)
	})

	t.Run("wraps query error", func(t *testing.T) {
		t.Parallel()

		got, err := rowDB(fakeRow{err: rateErr}, nil).SaveRate(context.Background(), *rateWant, rateTS)

		assert.Nil(t, got)
		assert.ErrorIs(t, err, rateErr)
	})
}
