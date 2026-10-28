package entities

import (
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

var ErrRateNotFound = eris.New("rate not found")

type Rate struct {
	Pair     Pair
	Price    decimal.Decimal
	Provider values.ProviderName
}
