package entities

import (
	"github.com/shopspring/decimal"

	"github.com/LiquidCats/paw/internal/kernel/domain/values"
)

type Transaction[M Movement] struct {
	TxID     values.TxID
	Fee      decimal.Decimal
	Movement M
}
