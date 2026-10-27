package entities

import "github.com/LiquidCats/paw/internal/kernel/domain/values"

type Contract struct {
	Address  values.Address
	Ticker   values.Ticker
	Decimals uint
}
