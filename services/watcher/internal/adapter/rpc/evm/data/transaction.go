package data

import (
	"math/big"

	"github.com/LiquidCats/paw/services/watcher/internal/adapter/rpc/evm/data/common"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
)

var (
	TxTypeLegacy0 = big.NewInt(0) // nolint:gochecknoglobals
	TxTypeLegacy1 = big.NewInt(1) // nolint:gochecknoglobals
	TxTypeEIP1559 = big.NewInt(2) // nolint:mnd
)

type Transaction struct {
	BlockHash   entities.BlockHash `json:"blockHash"`
	BlockNumber *common.Uint64     `json:"blockNumber"`
	From        entities.Address   `json:"from"`
	Gas         *common.Uint64     `json:"gas"`
	GasPrice    *common.Big        `json:"gasPrice"`
	Hash        entities.TxID      `json:"hash"`
	Input       string             `json:"input"`
	Nonce       *common.Uint64     `json:"nonce"`
	To          entities.Address   `json:"to"`
	Value       *common.Big        `json:"value"`
	Type        *common.Big        `json:"type"`
}
