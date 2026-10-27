package values_test

import (
	"testing"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	"github.com/stretchr/testify/require"
)

func TestSymbolCase(t *testing.T) {
	s := values.Symbol("btc_usd")

	require.Equal(t, "BTC_USD", s.ToUpper().String())
	require.Equal(t, "btc_usd", s.ToLower().String())
}

func TestNewSymbolFromString(t *testing.T) {
	require.Equal(t, values.Symbol("BTC_USD"), values.NewSymbolFromString("btc_usd"))
}

func TestNewSymbol(t *testing.T) {
	require.Equal(t, values.Symbol("BTC_USD"), values.NewSymbol("btc", "usd"))
}
