package entities_test

import (
	"testing"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	"github.com/stretchr/testify/require"
)

func TestNewPairFromSymbolUppercasesParts(t *testing.T) {
	pair := entities.NewPairFromSymbol(values.Symbol("btc_usd"))

	require.Equal(t, values.CurrencyISO("BTC"), pair.From)
	require.Equal(t, values.CurrencyISO("USD"), pair.To)
	require.Equal(t, values.Symbol("btc_usd"), pair.Symbol)
}

func TestPairJoin(t *testing.T) {
	pair := entities.Pair{From: "BTC", To: "USD"}

	require.Equal(t, "BTC-USD", pair.Join("-"))
}
