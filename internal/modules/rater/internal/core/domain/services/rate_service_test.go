package service_test

import (
	"context"
	"database/sql"
	"io"
	"testing"
	"time"

	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	kernelvalues "github.com/LiquidCats/paw/internal/kernel/domain/values"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/outbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	domainerrors "github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/errors"
	service "github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/services"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/tests/mocks"
	"github.com/LiquidCats/paw/internal/shares/logs"
	"github.com/LiquidCats/paw/internal/shares/regsitry"
	kernelmocks "github.com/LiquidCats/paw/tests/mocks/kernelmocks"
)

var (
	ctx    = context.Background()
	symbol = values.Symbol("BTC_USD")
	pair   = entities.NewPairFromSymbol(symbol)
	bucket = time.Date(2026, 1, 2, 3, 5, 0, 0, time.UTC)
	errDB  = eris.New("db down")
)

type fixture struct {
	svc *service.RateService
	api *mocks.MockRateAPIRepository
	db  *mocks.MockRateDatabaseRepository
}

// newFixture wires one provider so iteration order is deterministic.
func newFixture(t *testing.T) fixture {
	t.Helper()

	api := mocks.NewMockRateAPIRepository(t)
	db := mocks.NewMockRateDatabaseRepository(t)
	clock := kernelmocks.NewMockClock(t)
	clock.EXPECT().RoundToNearest(mock.Anything, kernelvalues.FiveMinuteBucket).Return(bucket).Maybe()

	registry := regsitry.NewRegistry[values.ProviderName, outbound.RateAPIRepository]()
	registry.Add(values.ProviderNameCex, api)

	return fixture{
		svc: service.NewRateService(logs.NewSlogAdapter(io.Discard), clock, registry, db),
		api: api,
		db:  db,
	}
}

func TestRateService_Current(t *testing.T) {
	t.Parallel()

	price := decimal.NewFromInt(42)
	want := entities.Rate{Pair: pair, Price: price, Provider: values.ProviderNameCex}

	t.Run("rate already stored is not saved again", func(t *testing.T) {
		t.Parallel()
		f := newFixture(t)
		f.api.EXPECT().GetRate(ctx, pair).Return(price, nil).Once()
		f.db.EXPECT().HasRate(ctx, symbol, bucket).Return(true, nil).Once()

		got, provider, err := f.svc.Current(ctx, pair)

		require.NoError(t, err)
		assert.Equal(t, &want, got)
		assert.Equal(t, values.ProviderNameCex, provider)
	})

	t.Run("new rate is saved", func(t *testing.T) {
		t.Parallel()
		f := newFixture(t)
		f.api.EXPECT().GetRate(ctx, pair).Return(price, nil).Once()
		f.db.EXPECT().HasRate(ctx, symbol, bucket).Return(false, nil).Once()
		f.db.EXPECT().SaveRate(ctx, want, bucket).Return(&want, nil).Once()

		got, _, err := f.svc.Current(ctx, pair)

		require.NoError(t, err)
		assert.Equal(t, &want, got)
	})

	t.Run("has-rate failure still returns the rate without saving", func(t *testing.T) {
		t.Parallel()
		f := newFixture(t)
		f.api.EXPECT().GetRate(ctx, pair).Return(price, nil).Once()
		f.db.EXPECT().HasRate(ctx, symbol, bucket).Return(false, errDB).Once()

		got, _, err := f.svc.Current(ctx, pair)

		require.NoError(t, err)
		assert.Equal(t, &want, got)
	})

	t.Run("save failure still returns the rate", func(t *testing.T) {
		t.Parallel()
		f := newFixture(t)
		f.api.EXPECT().GetRate(ctx, pair).Return(price, nil).Once()
		f.db.EXPECT().HasRate(ctx, symbol, bucket).Return(false, nil).Once()
		f.db.EXPECT().SaveRate(ctx, want, bucket).Return(nil, errDB).Once()

		got, _, err := f.svc.Current(ctx, pair)

		require.NoError(t, err)
		assert.Equal(t, &want, got)
	})

	t.Run("provider errors", func(t *testing.T) {
		t.Parallel()

		tests := map[string]error{
			"generic":        errDB,
			"request failed": eris.Wrap(&domainerrors.ProviderRequestFailedError{StatusCode: 500, Body: "x"}, "api"),
		}

		for name, apiErr := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				f := newFixture(t)
				f.api.EXPECT().GetRate(ctx, pair).Return(decimal.Zero, apiErr).Once()

				got, provider, err := f.svc.Current(ctx, pair)

				assert.Nil(t, got)
				assert.Equal(t, values.ProviderNameCex, provider)
				assert.ErrorIs(t, err, domainerrors.ErrRateNotAvailable)
			})
		}
	})

	t.Run("zero price is not available", func(t *testing.T) {
		t.Parallel()
		f := newFixture(t)
		f.api.EXPECT().GetRate(ctx, pair).Return(decimal.Zero, nil).Once()

		_, _, err := f.svc.Current(ctx, pair)

		assert.ErrorIs(t, err, domainerrors.ErrRateNotAvailable)
	})

	t.Run("no providers registered", func(t *testing.T) {
		t.Parallel()
		clock := kernelmocks.NewMockClock(t)
		svc := service.NewRateService(
			logs.NewSlogAdapter(io.Discard), clock,
			regsitry.NewRegistry[values.ProviderName, outbound.RateAPIRepository](),
			mocks.NewMockRateDatabaseRepository(t),
		)

		got, _, err := svc.Current(ctx, pair)

		assert.Nil(t, got)
		assert.ErrorIs(t, err, domainerrors.ErrRateNotAvailable)
	})
}

func TestRateService_Historical(t *testing.T) {
	t.Parallel()

	want := &entities.Rate{Pair: pair, Price: decimal.NewFromInt(7)}

	tests := []struct {
		name    string
		rate    *entities.Rate
		dbErr   error
		wantErr error
	}{
		{name: "found", rate: want},
		{name: "no rows", dbErr: eris.Wrap(sql.ErrNoRows, "query"), wantErr: domainerrors.ErrNoHistoricalRate},
		{name: "db failure", dbErr: errDB, wantErr: errDB},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newFixture(t)
			f.db.EXPECT().GetRate(ctx, symbol, bucket).Return(tt.rate, tt.dbErr).Once()

			got, err := f.svc.Historical(ctx, pair, bucket)

			if tt.wantErr != nil {
				assert.Nil(t, got)
				assert.ErrorIs(t, err, tt.wantErr)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.rate, got)
		})
	}
}
