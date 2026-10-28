package values_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

func TestCurrencyISO(t *testing.T) {
	t.Parallel()

	c := values.CurrencyISO("bTc")

	assert.Equal(t, values.CurrencyISO("BTC"), c.ToUpper())
	assert.Equal(t, values.CurrencyISO("btc"), c.ToLower())
	assert.Equal(t, "bTc", c.String())
}

func TestProviderName_String(t *testing.T) {
	t.Parallel()

	tests := map[values.ProviderName]string{
		values.ProviderNameCex:           "cex",
		values.ProviderNameCoinApi:       "coinapi",
		values.ProviderNameCoinGate:      "coingate",
		values.ProviderNameCoinGecko:     "coingecko",
		values.ProviderNameCoinMarketCap: "coinmarketcap",
	}

	for name, want := range tests {
		t.Run(want, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, want, name.String())
		})
	}
}
