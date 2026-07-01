package entities

import "github.com/shopspring/decimal"

type (
	TxID           string
	RawTransaction string
)

type TxIDs []TxID

type Transaction[In any] struct {
	TxID      TxID
	BlockHash BlockHash
	Inputs    []In
	Outputs   []TransactionOutput
	Fee       decimal.Decimal
}

type TransactionUtxoInput struct {
	TxID TxID
	N    uint32
}

type TransactionAccountInput struct {
	Address Address
}

type TransactionOutput struct {
	N        uint32
	Value    decimal.Decimal
	Ticker   Ticker
	Contract Address
	Address  Address
}
