package coingecko_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coingecko"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/config"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
)

const response = `{"bitcoin": {"eur": 60715.43364698}}`

func TestCoinGateRepository_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer ts.Close()

	repo := coingecko.NewRepository(config.CoinGeckoConfig{
		URL: ts.URL,
	})

	rate, err := repo.GetRate(t.Context(), entities.Pair{
		From: "BTC",
		To:   "EUR",
	})
	require.NoError(t, err)
	require.NotNil(t, rate)

	require.Equal(t, "60715.43364698", rate.String())
}
