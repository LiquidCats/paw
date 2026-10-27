package coinapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coinapi"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/config"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
)

const response = `{"time":"2023-07-24T11:31:56.0000000Z","asset_id_base":"BTC","asset_id_quote":"USD","rate":29295.929694597355}`

func TestCoinApiRepository_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "test1234", r.Header.Get("X-CoinAPI-Key")) //nolint:canonicalheader

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(response))
		assert.NoError(t, err)
	}))
	defer ts.Close()

	repo := coinapi.NewRepository(config.CoinApiConfig{
		URL:    ts.URL,
		Secret: "test1234",
	})

	rate, err := repo.GetRate(t.Context(), entities.Pair{
		From: "BTC",
		To:   "USD",
	})
	require.NoError(t, err)
	require.NotNil(t, rate)

	require.Equal(t, "29295.929694597355", rate.String())
}
