package values_test

import (
	"testing"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	"github.com/stretchr/testify/require"
)

func TestProviderNameString(t *testing.T) {
	require.Equal(t, "cex", values.ProviderNameCex.String())
	require.Equal(t, "coinmarketcap", values.ProviderNameCoinMarketCap.String())
}

func TestProviderNameMarshalUnmarshalText(t *testing.T) {
	for _, p := range []values.ProviderName{
		values.ProviderNameCex,
		values.ProviderNameCoinApi,
		values.ProviderNameCoinGate,
		values.ProviderNameCoinGecko,
		values.ProviderNameCoinMarketCap,
	} {
		text, err := p.MarshalText()
		require.NoError(t, err)

		var got values.ProviderName
		require.NoError(t, got.UnmarshalText(text))
		require.Equal(t, p, got)
	}
}

func TestProviderNameUnmarshalTextInvalid(t *testing.T) {
	var p values.ProviderName
	err := p.UnmarshalText([]byte("unknown"))
	require.ErrorIs(t, err, values.ErrInvalidProviderName)
}

func TestProviderNameMarshalUnmarshalJSON(t *testing.T) {
	data, err := values.ProviderNameCoinGecko.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, `"coingecko"`, string(data))

	var got values.ProviderName
	require.NoError(t, got.UnmarshalJSON(data))
	require.Equal(t, values.ProviderNameCoinGecko, got)
}

func TestProviderNameUnmarshalJSONInvalid(t *testing.T) {
	var p values.ProviderName
	err := p.UnmarshalJSON([]byte(`"unknown"`))
	require.ErrorIs(t, err, values.ErrInvalidProviderName)
}
