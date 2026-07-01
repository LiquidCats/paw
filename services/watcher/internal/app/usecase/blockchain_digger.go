package usecase

import (
	"context"
	"slices"

	"github.com/LiquidCats/paw/services/watcher/configs"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/metrics"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/rpc"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/runner"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/state"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

type BlockchainDigger[TxIn any] struct {
	cfg           configs.ChainConfig
	blockState    state.SliceState[entities.BlockHash]
	inflightState state.MapState[entities.BlockHash, bool]
	rpcClient     rpc.Client[TxIn]
	workerCh      runner.ChanWrite[*entities.Block]
	metrics       BlocksJobMetrics
}

type BlocksJobMetrics struct {
	RequestToNodeCounter metrics.RequestToNodeCounter
}

func NewBlockchainDigger[TxIn any](
	cfg configs.ChainConfig,
	blockState state.SliceState[entities.BlockHash],
	inflightState state.MapState[entities.BlockHash, bool],
	rpcClient rpc.Client[TxIn],
	workerCh runner.ChanWrite[*entities.Block],
	metrics BlocksJobMetrics,
) *BlockchainDigger[TxIn] {
	return &BlockchainDigger[TxIn]{
		cfg:           cfg,
		blockState:    blockState,
		inflightState: inflightState,
		rpcClient:     rpcClient,
		workerCh:      workerCh,
		metrics:       metrics,
	}
}

func (uc *BlockchainDigger[TxIn]) Handle(ctx context.Context) error {
	var block *entities.Block
	var err error
	var blocks []*entities.Block

	logger := zerolog.Ctx(ctx).With().
		Str("name", "blocks_job").
		Any("driver", uc.cfg.Driver).
		Any("type", uc.cfg.Type).
		Any("chain", uc.cfg.Chain).
		Logger()

	blocksState := uc.blockState.Get()
	block, err = uc.rpcClient.GetLatestBlock(ctx)
	if err != nil {
		return eris.Wrap(err, "get latest block hash")
	}
	uc.metrics.RequestToNodeCounter.Inc(uc.cfg.Chain)

	if slices.Contains(blocksState, block.Hash) {
		return nil
	}

	if uc.inflightState.Has(block.Hash) {
		logger.Info().
			Any("block_hash", block.Hash).
			Any("block_num", block.Height).
			Msg("block already in flight")
		return nil
	}

	blocks = append(blocks, block)

	for {
		if len(blocks) >= uc.cfg.Scan.Depth {
			logger.Warn().Msg("scan block depth reached")
			break
		}

		blockHash := block.PrevHash

		block, err = uc.rpcClient.GetBlockByHash(ctx, blockHash)
		if err != nil {
			return eris.Wrapf(err, "get block blockhash=%s", blockHash)
		}
		uc.metrics.RequestToNodeCounter.Inc(uc.cfg.Chain)

		if uc.inflightState.Has(block.Hash) {
			logger.Info().
				Any("block_hash", block.Hash).
				Any("block_num", block.Height).
				Msg("block already in flight")
			break
		}

		blocks = append(blocks, block)
		if slices.Contains(blocksState, block.PrevHash) {
			logger.Info().
				Any("block_hash", block.Hash).
				Any("block_num", block.Height).
				Msg("block already known")
			break
		}

		if block.PrevHash == "" {
			logger.Info().
				Any("block_hash", block.Hash).
				Any("block_num", block.Height).
				Msg("block prev hash is empty")
			break
		}
	}

	slices.Reverse(blocks)

	logger.Info().
		Any("blocks_len", len(blocks)).
		Msg("blocks collected")

	for _, b := range blocks {
		uc.workerCh <- b
		uc.inflightState.Set(b.Hash, true)
	}

	return nil
}
