package main

import (
	"context"
	"net"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"

	"github.com/LiquidCats/libraries/cfg"
	"github.com/LiquidCats/libraries/db/postgres"
	"github.com/LiquidCats/paw/internal/config"
	"github.com/LiquidCats/paw/internal/modules/rater/bootstrap"
)

func main() {
	logger := zerolog.New(os.Stdout).
		With().
		Caller().
		Stack().
		Timestamp().
		Logger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redisCfg, err := cfg.Load[config.RedisConfig]()
	if err != nil {
		logger.Panic().Err(err).Msg("failed to load redis config")
	}
	pgCfg, err := cfg.Load[config.PostgresConfig]()
	if err != nil {
		logger.Panic().Err(err).Msg("failed to load postgres config")
	}
	zerologCfg, err := cfg.Load[config.ZerologConfig]()
	if err != nil {
		logger.Panic().Err(err).Msg("failed to load zerolog config")
	}

	ctx = logger.WithContext(ctx)

	zerolog.DefaultContextLogger = &logger // nolint:reassign
	zerolog.SetGlobalLevel(zerologCfg.Level)

	pool, err := postgres.Connect(
		ctx,
		postgres.WithDisableSSL(),
		postgres.WithHost(pgCfg.Host),
		postgres.WithPort(pgCfg.Port),
		postgres.WithDatabase(pgCfg.Database),
		postgres.WithUser(pgCfg.User),
		postgres.WithPassword(string(pgCfg.Password)),
	)

	if err != nil {
		logger.Panic().
			Any("err", eris.ToJSON(err, true)).
			Msg("connect to database")
	}
	defer pool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(redisCfg.Host, redisCfg.Port),
		DB:       redisCfg.DB,
		Protocol: redisCfg.Protocol,
		Password: string(redisCfg.Password),
	})
	defer redisClient.Close()

	bootstrap.Bootstrap(ctx, logger, pool, redisClient)
}
