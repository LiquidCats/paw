package main

import (
	"context"
	"net"
	"os"

	"github.com/redis/go-redis/v9"

	"github.com/LiquidCats/libraries/cfg"
	"github.com/LiquidCats/libraries/db/postgres"
	"github.com/LiquidCats/paw/internal/config"
	"github.com/LiquidCats/paw/internal/kernel/ports"
	"github.com/LiquidCats/paw/internal/modules/rater/bootstrap"
	"github.com/LiquidCats/paw/internal/modules/rater/migrations"
	"github.com/LiquidCats/paw/internal/shares/logs"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redisCfg, err := cfg.Load[config.RedisConfig](cfg.WithPrefix("redis"))
	if err != nil {
		panic(err)
	}
	pgCfg, err := cfg.Load[config.PostgresConfig](cfg.WithPrefix("database"))
	if err != nil {
		panic(err)
	}
	zerologCfg, err := cfg.Load[config.ZerologConfig](cfg.WithPrefix("logging"))
	if err != nil {
		panic(err)
	}

	logger := logs.NewZerologAdapter(
		os.Stdout,
		logs.WithCaller(),
		logs.WithStack(),
		logs.WithTimestamp(),
		logs.WithLevel(zerologCfg.Level),
	)

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
		logger.Panic("connect to database", ports.Err(err))
	}
	defer pool.Close()

	logger.Info("connected to database")

	err = postgres.MigrateUp(ctx, pool, migrations.FS)
	if err != nil {
		logger.Panic("database migration failed", ports.Err(err))
	}
	logger.Info("database migrated")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(redisCfg.Host, redisCfg.Port),
		DB:       redisCfg.DB,
		Protocol: redisCfg.Protocol,
		Password: string(redisCfg.Password),
	})
	defer redisClient.Close()

	logger.Info("connected to redis")

	bootstrap.Bootstrap(ctx, logger, pool, redisClient)
}
