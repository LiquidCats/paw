package configs

import (
	"time"

	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
)

type ChainsConfig []ChainConfig

type ChainConfig struct {
	Driver entities.Driver `yaml:"driver"`
	Type   entities.Type   `yaml:"type"`
	ISO    entities.Ticker `yaml:"iso"`
	Chain  entities.Chain  `yaml:"chain"`

	Persist PersistConfig `yaml:"persist"`
	Scan    ScanConfig    `yaml:"scan"`

	Workers WorkersConfig `yaml:"workers"`

	RPC RPCConfig `yaml:"rpc"`

	Topics TopicsConfig `yaml:"topics"`
}

type TopicsConfig struct {
	Transactions string `yaml:"transactions"`
	Blocks       string `yaml:"blocks"`
}

type RPCConfig struct {
	NodeURL string `yaml:"node_url"`
}

type ScanConfig struct {
	Depth    int           `yaml:"depth"`
	Interval time.Duration `yaml:"interval"`
}

type PersistConfig struct {
	Capacity int           `yaml:"capacity"`
	Interval time.Duration `yaml:"interval"`
}

type WorkersConfig struct {
	TxIDWorkerCount              uint `default:"3" yaml:"txid_worker_count"`
	BlockTransactionsWorkerCount uint `default:"5" yaml:"block_transactions_worker_count"`
}
