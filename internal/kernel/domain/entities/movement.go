package entities

import (
	"github.com/shopspring/decimal"

	"github.com/LiquidCats/paw/internal/kernel/domain/values"
)

type Movement interface {
	AccountMovement | UTXOMovement
}

type AccountMovement struct {
	Initializer values.Address
	Contract    *values.Address
	Transfers   []Transfer
}

type Transfer struct {
	From       values.Address
	To         values.Address
	Amount     decimal.Decimal
	Contract   *Contract
	Confidence values.TransferConfidence
}

type UTXOMovement struct {
	Inputs  []TransactionInput
	Outputs []TransactionOutput
}

type TransactionInput struct {
	TxID  values.TxID
	Index uint32
}

type TransactionOutput struct {
	Address values.Address
	Amount  decimal.Decimal
	Index   uint32
}
