package usecase

import (
	"context"

	"github.com/LiquidCats/paw/watcher/configs"
	"github.com/LiquidCats/paw/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/watcher/internal/app/port/metrics"
	"github.com/LiquidCats/paw/watcher/internal/app/port/rpc"
	"github.com/LiquidCats/paw/watcher/internal/app/port/runner"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

type MempoolDigger[TxIn any] struct {
	cfg       configs.ChainConfig
	rpcClient rpc.Client[TxIn]
	txIDCh    runner.ChanWrite[entities.TxID]

	oldMempool []entities.TxID

	metrics MempoolJobMetrics
}

type MempoolJobMetrics struct {
	RequestToNodeCounter metrics.RequestToNodeCounter
}

func NewMempoolDigger[TxIn any](
	cfg configs.ChainConfig,
	rpcClient rpc.Client[TxIn],
	txIDCh runner.ChanWrite[entities.TxID],
	oldMempool []entities.TxID,
	metrics MempoolJobMetrics,
) *MempoolDigger[TxIn] {
	return &MempoolDigger[TxIn]{
		cfg:        cfg,
		rpcClient:  rpcClient,
		txIDCh:     txIDCh,
		oldMempool: oldMempool,
		metrics:    metrics,
	}
}

func (uc *MempoolDigger[TxIn]) Handle(ctx context.Context) error {
	logger := zerolog.Ctx(ctx).
		With().
		Any("chain", uc.cfg.Chain).
		Any("driver", uc.cfg.Driver).
		Any("type", uc.cfg.Type).
		Str("module", "mempool_processor").
		Logger()

	newMempool, err := uc.rpcClient.GetMempool(ctx)
	uc.metrics.RequestToNodeCounter.Inc(uc.cfg.Chain)
	if err != nil {
		return eris.Wrap(err, "get new mempool")
	}

	if len(newMempool) == 0 {
		return nil
	}

	m := make(map[entities.TxID]struct{}, len(newMempool))

	for _, txID := range uc.oldMempool {
		m[txID] = struct{}{}
	}

	var diff []entities.TxID

	for _, txID := range newMempool {
		_, ok := m[txID]
		if !ok {
			diff = append(diff, txID)
		}
	}

	logger.Info().Any("diff_len", len(diff)).Msg("found new transactions")

	for _, txID := range diff {
		uc.txIDCh <- txID
	}

	uc.oldMempool = newMempool

	return nil
}
