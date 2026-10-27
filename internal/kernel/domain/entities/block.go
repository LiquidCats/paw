package entities

import "github.com/LiquidCats/paw/internal/kernel/domain/values"

type BlockHeader struct {
	Height   values.BlockHeight
	Hash     values.BlockHash
	PrevHash values.BlockHash
}

type Block[T any] struct {
	BlockHeader
	Transactions []T
}
