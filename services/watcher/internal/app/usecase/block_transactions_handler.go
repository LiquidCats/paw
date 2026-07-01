package usecase

import (
	"context"

	"github.com/LiquidCats/paw/services/watcher/configs"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/bus"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/metrics"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/rpc"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/state"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

type BlockTransactionsHandler[TxIn any] struct {
	cfg            configs.ChainConfig
	rpc            rpc.Client[TxIn]
	transactionPub bus.Publisher[entities.Transaction[TxIn]]
	blockPub       bus.Publisher[entities.Block]
	inflightState  state.MapState[entities.BlockHash, bool]
	metrics        BlockTransactionsHandlerMetrics
}

type BlockTransactionsHandlerMetrics struct {
	RequestToNodeCounter metrics.RequestToNodeCounter
}

func NewBlockTransactionsHandler[TxIn any](
	cfg configs.ChainConfig,
	rpc rpc.Client[TxIn],
	transactionPub bus.Publisher[entities.Transaction[TxIn]],
	blockPub bus.Publisher[entities.Block],
	inflightState state.MapState[entities.BlockHash, bool],
	metrics BlockTransactionsHandlerMetrics,
) *BlockTransactionsHandler[TxIn] {
	return &BlockTransactionsHandler[TxIn]{
		cfg:            cfg,
		rpc:            rpc,
		transactionPub: transactionPub,
		blockPub:       blockPub,
		inflightState:  inflightState,
		metrics:        metrics,
	}
}

func (uc *BlockTransactionsHandler[TxIn]) Handle(ctx context.Context, block *entities.Block) error {
	defer uc.inflightState.Del(block.Hash)

	logger := zerolog.Ctx(ctx).With().
		Any("driver", uc.cfg.Driver).
		Any("type", uc.cfg.Type).
		Any("chain", uc.cfg.Chain).
		Any("block_hash", block.Hash).
		Any("block_num", block.Height).
		Logger()

	blockWithTransactions, err := uc.rpc.GetBlockByHashWithTransactions(ctx, block.Hash)
	if err != nil {
		return eris.Wrap(err, "get block by hash")
	}
	uc.metrics.RequestToNodeCounter.Inc(uc.cfg.Chain)

	for _, transaction := range blockWithTransactions.Transactions {
		logger.Debug().Any("txid", transaction.TxID).Msg("publish transaction")
		err = uc.transactionPub.PublishTo(ctx, uc.cfg.Topics.Transactions, transaction)
		if err != nil {
			logger.Error().
				Any("err", eris.ToJSON(err, true)).
				Any("txid", transaction.TxID).
				Msg("publish transaction")
			return eris.Wrap(err, "publish transaction")
		}
	}

	err = uc.blockPub.PublishTo(ctx, uc.cfg.Topics.Blocks, *block)
	if err != nil {
		logger.Error().Any("err", eris.ToJSON(err, true)).Msg("publish block")
		return eris.Wrap(err, "publish block")
	}

	logger.Debug().Int("txs", len(block.Transactions)).Msg("block transactions handled")

	return nil
}
