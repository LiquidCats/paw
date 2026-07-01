package data

import (
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/rpc/evm/data/common"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
)

type TransactionReceipts []TransactionReceipt

type TransactionReceipt struct {
	BlockHash         entities.BlockHash `json:"blockHash"`
	BlockNumber       *common.Big        `json:"blockNumber"`
	ContractAddress   entities.Address   `json:"contractAddress"`
	CumulativeGasUsed *common.Big        `json:"cumulativeGasUsed"`
	EffectiveGasPrice *common.Big        `json:"effectiveGasPrice"`
	From              entities.Address   `json:"from"`
	GasUsed           *common.Big        `json:"gasUsed"`
	Logs              []TransactionLog   `json:"logs"`
	TransactionHash   entities.TxID      `json:"transactionHash"`
}

type TransactionLog struct {
	Topics  []string         `json:"topics"`
	Data    string           `json:"data"`
	Address entities.Address `json:"address"`
}
