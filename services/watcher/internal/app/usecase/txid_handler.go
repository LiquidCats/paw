package usecase

import (
	"context"

	"github.com/LiquidCats/paw/services/watcher/configs"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/bus"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/metrics"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/rpc"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

type TxIDHandler[TxIn any] struct {
	cfg       configs.ChainConfig
	rpcClient rpc.Client[TxIn]
	publisher bus.Publisher[entities.Transaction[TxIn]]
	metrics   TxIDHandlerMetrics
}

type TxIDHandlerMetrics struct {
	RequestToNodeCounter metrics.RequestToNodeCounter
}

func NewTxIDHandler[TxIn any](
	cfg configs.ChainConfig,
	rpcClient rpc.Client[TxIn],
	publisher bus.Publisher[entities.Transaction[TxIn]],
	metrics TxIDHandlerMetrics,
) *TxIDHandler[TxIn] {
	return &TxIDHandler[TxIn]{
		cfg:       cfg,
		rpcClient: rpcClient,
		publisher: publisher,
		metrics:   metrics,
	}
}

func (uc *TxIDHandler[TxIn]) Handle(ctx context.Context, txid entities.TxID) error {
	logger := zerolog.Ctx(ctx).With().
		Str("name", "txid_handler").
		Any("driver", uc.cfg.Driver).
		Any("type", uc.cfg.Type).
		Any("chain", uc.cfg.Chain).
		Any("txid", txid).
		Logger()

	tx, err := uc.rpcClient.GetTransactionByTxID(ctx, txid)
	uc.metrics.RequestToNodeCounter.Inc(uc.cfg.Chain)
	if err != nil {
		return eris.Wrap(err, "get transaction by txid")
	}

	logger.Info().Msg("got transaction by hash")

	err = uc.publisher.PublishTo(ctx, uc.cfg.Topics.Transactions, *tx)
	if err != nil {
		return eris.Wrap(err, "publish mempool transaction")
	}

	return nil
}
