package bootstrap

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rotisserie/eris"

	"github.com/LiquidCats/libraries/cfg"
	"github.com/LiquidCats/libraries/db/postgres"
	"github.com/LiquidCats/libraries/graceful"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/inbound/cron"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/cex"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coinapi"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coingate"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coingecko"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coinmarketcap"
	postgresRepos "github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres/db"
	redisRepos "github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/redis"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/config"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/outbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/usecase"
	service "github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/services"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/migrations"
	"github.com/LiquidCats/paw/internal/shares/regsitry"

	"github.com/rs/zerolog"
)

const ModuleName = "rater"

func Bootstrap(ctx context.Context, logger zerolog.Logger, pool *pgxpool.Pool, redisClient *redis.Client) {
	logger = logger.With().Str("module", ModuleName).Logger()

	cfg, err := cfg.Load[config.Config](cfg.WithPrefix(ModuleName))

	logger.Info().Msg("connected to database")

	err = postgres.MigrateUp(ctx, pool, migrations.FS)
	if err != nil {
		logger.Panic().
			Any("err", eris.ToJSON(err, true)).
			Msg("database migration failed")
	}
	logger.Info().Msg("database migrated")

	redisRepo := redisRepos.NewRepository(redisClient)

	queries := db.New(pool)
	dbRepo := postgresRepos.NewRepository(queries)

	apiRegistry := regsitry.NewRegistry[values.ProviderName, outbound.RateAPIRepository]()
	apiRegistry.Add(values.ProviderNameCex, cex.NewRepository(cfg.Providers.Cex))
	apiRegistry.Add(values.ProviderNameCoinApi, coinapi.NewRepository(cfg.Providers.CoinApi))
	apiRegistry.Add(values.ProviderNameCoinGate, coingate.NewRepository(cfg.Providers.CoinGate))
	apiRegistry.Add(values.ProviderNameCoinGecko, coingecko.NewRepository(cfg.Providers.CoinGecko))
	apiRegistry.Add(values.ProviderNameCoinMarketCap, coinmarketcap.NewRepository(cfg.Providers.CoinMarketCap))

	rateService := service.NewRateService(apiRegistry, dbRepo)

	rateUseCase := usecase.NewRateUseCase(redisRepo, rateService)

	rateTaskLogger := logger.With().Str("task", "collect_rate").Logger()

	collectRateTask := cron.NewCollectRateTask(rateTaskLogger, cfg.App, dbRepo, rateUseCase)
	// rateGrpcService := grpc.NewServer(logger, rateUseCase)

	logger.Info().Msg("starting service")
	err = graceful.WaitContext(
		ctx,
		graceful.Signals,
		graceful.ScheduleRunner(collectRateTask),
		// graceful.GRPCRunner(rateGrpcService),
	)
	if err != nil {
		logger.Panic().
			Any("err", eris.ToJSON(err, true)).
			Msg("failed to wait for graceful shutdown")
	}

	logger.Info().Msg("service stopped")
}
