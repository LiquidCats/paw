package entities

import (
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	"github.com/shopspring/decimal"
)

type Rate struct {
	Pair     Pair
	Price    decimal.Decimal
	Provider values.ProviderName
}
