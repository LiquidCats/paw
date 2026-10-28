package bootstrap

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/LiquidCats/libraries/cfg"
	"github.com/LiquidCats/libraries/graceful"
	"github.com/LiquidCats/paw/internal/kernel/ports"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/inbound/presentation/grpc"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api"
	postgresRepos "github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres/db"
	redisRepos "github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/redis"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/config"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/outbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/usecase"
	service "github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/services"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	"github.com/LiquidCats/paw/internal/shares/clock"
	"github.com/LiquidCats/paw/internal/shares/regsitry"
)

const ModuleName = "rater"

func Bootstrap(ctx context.Context, logger ports.Logger, pool *pgxpool.Pool, redisClient *redis.Client) {
	logger = logger.With(ports.Str("module", ModuleName))

	cfg, err := cfg.Load[config.Config](cfg.WithFilePath("configs/rater.config.yaml"))
	if err != nil {
		logger.Panic(
			"failed to load config",
			ports.Err(err),
		)
	}

	clock := clock.NewClock()

	redisRepo := redisRepos.NewRepository(redisClient)

	queries := db.New(pool)
	dbRepo := postgresRepos.NewRepository(queries)

	apiRegistry := regsitry.NewRegistry[values.ProviderName, outbound.RateAPIRepository]()
	for _, providerCfg := range cfg.Providers {
		if !providerCfg.Enabled {
			continue
		}
		provider, providerErr := api.Factory(providerCfg.Name, providerCfg)
		if providerErr != nil {
			logger.Panic(
				"failed to create provider",
				ports.Err(providerErr),
				ports.Str("provider", string(providerCfg.Name)),
			)
		}

		apiRegistry.Add(providerCfg.Name, provider)

		logger.Info("provider added", ports.Str("provider", string(providerCfg.Name)))
	}

	rateService := service.NewRateService(logger, clock, apiRegistry, dbRepo)

	rateUseCase := usecase.NewRateUseCase(logger, clock, redisRepo, rateService)
	rateGrpcService := grpc.NewServer(rateUseCase)

	logger.Info("starting service")
	defer logger.Info("service stopped")

	err = graceful.WaitContext(
		ctx,
		graceful.Signals,
		graceful.GRPCRunner(rateGrpcService),
	)
	if err != nil {
		logger.Panic("failed to wait for graceful shutdown", ports.Err(err))
	}
}
