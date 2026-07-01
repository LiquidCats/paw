package data

import (
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/rpc/evm/data/common"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
)

type Block[T any] struct {
	Hash         entities.BlockHash `json:"hash"`
	Number       *common.Big        `json:"number"`
	ParentHash   entities.BlockHash `json:"parentHash"`
	Transactions []T                `json:"transactions"`
}
