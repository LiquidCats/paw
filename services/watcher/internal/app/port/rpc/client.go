package rpc

import (
	"context"

	"github.com/LiquidCats/paw/watcher/internal/app/domain/entities"
)

type Client[TxIn any] interface {
	GetLatestBlock(ctx context.Context) (*entities.Block, error)
	GetBlockByHash(ctx context.Context, hash entities.BlockHash) (*entities.Block, error)
	GetBlockByHashWithTransactions(
		ctx context.Context,
		hash entities.BlockHash,
	) (*entities.BlockWithTransactions[TxIn], error)
	GetTransactionByTxID(ctx context.Context, hash entities.TxID) (*entities.Transaction[TxIn], error)
	GetMempool(ctx context.Context) ([]entities.TxID, error)
}
