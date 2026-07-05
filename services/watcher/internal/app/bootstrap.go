package app

import (
	"context"

	db "github.com/LiquidCats/paw/lib/database"
	"github.com/LiquidCats/paw/lib/graceful"
	"github.com/LiquidCats/paw/services/watcher/configs"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/bus"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/http/handlers"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/http/router"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/metrics"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/postgresql"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/postgresql/database"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/rpc/evm"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/rpc/utxo"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/state"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/watcher/internal/app/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

const ApplicationName = "watcher"

func Run(ctx context.Context, cfg configs.Config, pool *pgxpool.Pool) error {
	runners := []graceful.Runner{
		graceful.Signals,
		graceful.Server(
			router.NewGinRouter(handlers.NewHealth()),
			graceful.WithPort(cfg.HTTP.Port),
			graceful.WithReadTimeout(cfg.HTTP.ReadTimeout),
			graceful.WithWriteTimeout(cfg.HTTP.WriteTimeout),
		),
		graceful.Server(
			router.NewGinRouter(handlers.NewMetrics()),
			graceful.WithPort(cfg.Metrics.Port),
			graceful.WithReadTimeout(cfg.Metrics.ReadTimeout),
			graceful.WithWriteTimeout(cfg.Metrics.WriteTimeout),
		),
	}

	logger := zerolog.Ctx(ctx)

	requestsToNodeMetric := metrics.NewRequestsToNodeCount(ApplicationName)

	redisClient := redis.NewClient(cfg.Redis.ToConfig(ApplicationName))

	queries := database.New(pool)

	txManager := db.NewTxManager(pool)
	queriesTxManager := db.NewQueriesTxManager[database.Queries](txManager, queries)

	dbRepository := postgresql.NewRepository(queriesTxManager)

	for _, chainConfig := range cfg.Chains {
		blockChan := make(chan *entities.Block, chainConfig.Workers.BlockTransactionsWorkerCount)
		defer close(blockChan)

		txIDChan := make(chan entities.TxID, chainConfig.Workers.TxIDWorkerCount)
		defer close(txIDChan)

		switch chainConfig.Type {
		case entities.TypeEvm:
			chainRunners := bootstrapEvmBased(ctx, chainConfig, redisClient, dbRepository, requestsToNodeMetric, blockChan)
			runners = append(runners, chainRunners...)
		case entities.TypeUtxo:
			chainRunners := bootstrapUtxoBased(
				ctx,
				chainConfig,
				redisClient,
				dbRepository,
				requestsToNodeMetric,
				txIDChan,
				blockChan,
			)
			runners = append(runners, chainRunners...)
		default:
			logger.Error().Msgf("unsupported chain type: %s", chainConfig.Type)
		}
	}

	return graceful.WaitContext(ctx, runners...)
}

func bootstrapEvmBased(
	ctx context.Context,
	chainConfig configs.ChainConfig,
	redisClient *redis.Client,
	dbRepository *postgresql.Repository,
	requestsToNodeMetric *metrics.RequestsToNodeCount,
	blockChan chan *entities.Block,
) []graceful.Runner {
	logger := zerolog.Ctx(ctx)

	blocksState := state.NewSliceState[entities.BlockHash](chainConfig.Persist.Capacity)
	inflightState := state.NewMapState[entities.BlockHash, bool](chainConfig.Persist.Capacity)

	blocksPublisher := bus.NewRedisPublisher[entities.Block](redisClient)
	transactionsPublisher := bus.NewRedisPublisher[entities.Transaction[entities.TransactionAccountInput]](redisClient)

	evmAdapter := evm.NewClient(chainConfig.RPC, chainConfig.ISO, nil)

	blockchainDigger := usecase.NewBlockchainDigger[entities.TransactionAccountInput](
		chainConfig,
		blocksState,
		inflightState,
		evmAdapter,
		nil,
		usecase.BlocksJobMetrics{
			RequestToNodeCounter: requestsToNodeMetric,
		})
	statePersisterJob := usecase.NewBlocksPersister(
		chainConfig,
		dbRepository,
		blocksState,
	)
	blockTransactionsHandler := usecase.NewBlockTransactionsHandler[entities.TransactionAccountInput](
		chainConfig,
		evmAdapter,
		transactionsPublisher,
		blocksPublisher,
		inflightState,
		usecase.BlockTransactionsHandlerMetrics{
			RequestToNodeCounter: requestsToNodeMetric,
		},
	)

	return []graceful.Runner{
		graceful.Ticker(
			chainConfig.Scan.Interval,
			blockchainDigger.Handle,
			graceful.WithTickerLogger(logger),
		),
		graceful.Ticker(
			chainConfig.Persist.Interval,
			statePersisterJob.Handle,
			graceful.WithTickerLogger(logger),
		),
		graceful.Worker(
			blockChan,
			blockTransactionsHandler.Handle,
			graceful.WithWorkerLogger(logger),
		),
	}
}

func bootstrapUtxoBased(
	ctx context.Context,
	chainConfig configs.ChainConfig,
	redisClient *redis.Client,
	dbRepository *postgresql.Repository,
	requestsToNodeMetric *metrics.RequestsToNodeCount,
	txIDChan chan entities.TxID,
	blockChan chan *entities.Block,
) []graceful.Runner {
	logger := zerolog.Ctx(ctx)

	blocksState := state.NewSliceState[entities.BlockHash](chainConfig.Persist.Capacity)
	inflightState := state.NewMapState[entities.BlockHash, bool](chainConfig.Persist.Capacity)

	blocksPublisher := bus.NewRedisPublisher[entities.Block](redisClient)
	transactionsPublisher := bus.NewRedisPublisher[entities.Transaction[entities.TransactionUtxoInput]](redisClient)

	utxoAdapter := utxo.NewClient(chainConfig.RPC, chainConfig.ISO)

	oldMempool, err := utxoAdapter.GetMempool(ctx)
	if err != nil {
		logger.Fatal().Any("error", eris.ToJSON(err, true)).Msg("failed to get mempool")
	}

	mempoolDigger := usecase.NewMempoolDigger[entities.TransactionUtxoInput](
		chainConfig,
		utxoAdapter,
		nil,
		oldMempool,
		usecase.MempoolJobMetrics{
			RequestToNodeCounter: requestsToNodeMetric,
		})
	blockchainDigger := usecase.NewBlockchainDigger[entities.TransactionUtxoInput](
		chainConfig,
		blocksState,
		inflightState,
		utxoAdapter,
		nil,
		usecase.BlocksJobMetrics{
			RequestToNodeCounter: requestsToNodeMetric,
		})
	statePersisterJob := usecase.NewBlocksPersister(
		chainConfig,
		dbRepository,
		blocksState,
	)
	txIDHandler := usecase.NewTxIDHandler[entities.TransactionUtxoInput](
		chainConfig,
		utxoAdapter,
		transactionsPublisher,
		usecase.TxIDHandlerMetrics{
			RequestToNodeCounter: requestsToNodeMetric,
		},
	)
	blockTransactionsHandler := usecase.NewBlockTransactionsHandler[entities.TransactionUtxoInput](
		chainConfig,
		utxoAdapter,
		transactionsPublisher,
		blocksPublisher,
		inflightState,
		usecase.BlockTransactionsHandlerMetrics{
			RequestToNodeCounter: requestsToNodeMetric,
		},
	)
	return []graceful.Runner{
		graceful.Ticker(
			chainConfig.Scan.Interval,
			blockchainDigger.Handle,
			graceful.WithTickerLogger(logger),
		),
		graceful.Ticker(
			chainConfig.Scan.Interval,
			mempoolDigger.Handle,
			graceful.WithTickerLogger(logger),
		),
		graceful.Ticker(
			chainConfig.Persist.Interval,
			statePersisterJob.Handle,
			graceful.WithTickerLogger(logger),
		),
		graceful.Worker(
			txIDChan,
			txIDHandler.Handle,
			graceful.WithWorkerLogger(logger),
		),
		graceful.Worker(
			blockChan,
			blockTransactionsHandler.Handle,
			graceful.WithWorkerLogger(logger),
		),
	}
}
