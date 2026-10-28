package cex_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/cex"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/config"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
)

const response = `{
  "timestamp": "12345678",
  "low": "66830.7",
  "high": "66830.7",
  "last": "66830.7",
  "volume": "0.10000000",
  "volume30d": "1234.00000200",
  "bid": 1234.12,
  "ask": 1234.12,
  "priceChange": "-2058.3",
  "priceChangePercentage": "-9.18",
  "pair": "BTC:USD"
}`

func TestRepository_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "BTC/USD")

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(response))
		assert.NoError(t, err)
	}))
	defer ts.Close()

	repo := cex.NewRepository(config.ProviderConfig{
		URL: ts.URL,
	})

	ctx := t.Context()

	res, err := repo.GetRate(ctx, entities.Pair{
		From: "BTC",
		To:   "USD",
	})

	require.NoError(t, err)

	require.Equal(t, "66830.7", res.String())
}
