package coinmarketcap_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coinmarketcap"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/config"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	domainerrors "github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/errors"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

func TestRepository_GetRateErrors(t *testing.T) {
	t.Parallel()

	closed := httptest.NewServer(http.NotFoundHandler())
	closedURL := closed.URL
	closed.Close()

	tests := []struct {
		name    string
		status  int
		body    string
		url     string
		from    string
		wantAPI bool
	}{
		{name: "provider 5xx", status: http.StatusInternalServerError, body: `"boom"`, wantAPI: true},
		{name: "client error without body", status: http.StatusBadRequest, body: ``},
		{name: "missing quote currency", status: http.StatusOK, body: `{"data":{"quote":{}}}`},
		{name: "malformed body", status: http.StatusOK, body: "not a rate"},
		{name: "invalid url", url: "http://a b"},
		{name: "unsupported scheme", url: "ftp://example.com"},
		{name: "connection refused", url: closedURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			url := srv.URL
			if tt.url != "" {
				url = tt.url
			}
			from := "BTC"
			if tt.from != "" {
				from = tt.from
			}

			rate, err := coinmarketcap.NewRepository(config.ProviderConfig{URL: url}).
				GetRate(t.Context(), entities.Pair{From: values.CurrencyISO(from), To: "USD"})

			require.Error(t, err)
			assert.True(t, rate.IsZero())

			if tt.wantAPI {
				var apiErr *domainerrors.ProviderRequestFailedError
				require.ErrorAs(t, err, &apiErr)
				assert.Equal(t, tt.status, apiErr.StatusCode)
				assert.Equal(t, "boom", apiErr.Body)
			}
		})
	}
}
