package redis_test

import (
	"context"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/redis"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	"github.com/LiquidCats/paw/tests/mocks/redismocks"
)

const rateKey = "rate:from:btc:to:usd"

var (
	ctx  = context.Background()
	pair = entities.NewPairFromSymbol(values.Symbol("btc_usd"))
)

func TestRepository_GetRate(t *testing.T) {
	t.Parallel()

	boom := eris.New("redis down")

	tests := []struct {
		name    string
		cmd     *goredis.StringCmd
		want    *entities.Rate
		wantErr error
	}{
		{
			name: "found",
			cmd:  goredis.NewStringResult("12.5", nil),
			want: &entities.Rate{Pair: pair, Price: decimal.RequireFromString("12.5")},
		},
		{name: "missing key", cmd: goredis.NewStringResult("", goredis.Nil), wantErr: entities.ErrRateNotFound},
		{name: "client failure", cmd: goredis.NewStringResult("", boom), wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := redismocks.NewMockUniversalClient(t)
			client.EXPECT().Get(ctx, rateKey).Return(tt.cmd).Once()

			got, err := redis.NewRepository(client).GetRate(ctx, pair)

			if tt.wantErr != nil {
				assert.Nil(t, got)
				assert.ErrorIs(t, err, tt.wantErr)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRepository_PutRate(t *testing.T) {
	t.Parallel()

	rate := entities.Rate{Pair: pair, Price: decimal.RequireFromString("12.5")}
	boom := eris.New("redis down")

	t.Run("stores price with short ttl", func(t *testing.T) {
		t.Parallel()
		client := redismocks.NewMockUniversalClient(t)
		client.EXPECT().Set(ctx, rateKey, "12.5", 5*time.Second).Return(goredis.NewStatusResult("OK", nil)).Once()

		require.NoError(t, redis.NewRepository(client).PutRate(ctx, rate))
	})

	t.Run("client failure", func(t *testing.T) {
		t.Parallel()
		client := redismocks.NewMockUniversalClient(t)
		client.EXPECT().Set(ctx, rateKey, "12.5", 5*time.Second).Return(goredis.NewStatusResult("", boom)).Once()

		assert.ErrorIs(t, redis.NewRepository(client).PutRate(ctx, rate), boom)
	})
}
