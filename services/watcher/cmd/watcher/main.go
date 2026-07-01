package main

import (
	"context"
	"os"

	"github.com/LiquidCats/paw/watcher/configs"
	"github.com/LiquidCats/paw/watcher/internal/adapter/database"
	"github.com/LiquidCats/paw/watcher/internal/app"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	_ "github.com/lib/pq"
	_ "go.uber.org/automaxprocs"
)

func main() {
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Caller().
		Logger()
	ctx := logger.WithContext(context.Background())

	cfg, err := configs.Load(app.ApplicationName)
	if err != nil {
		logger.Fatal().Stack().Err(err).Msg("failed to load config")
	}
	//
	zerolog.DefaultContextLogger = &logger // nolint:reassign
	zerolog.SetGlobalLevel(cfg.App.LogLevel)
	//
	poolConfig, err := pgxpool.ParseConfig(cfg.DB.ToDSN())
	if err != nil {
		logger.Fatal().Err(err).Msg("parse db config")
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("database config")
	}
	if err != nil {
		logger.Fatal().Stack().Err(err).Msg("connect to database")
	}
	defer pool.Close()

	migrationConn, err := pool.Acquire(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("acquire pool connection")
	}

	if err = database.Migrate(migrationConn.Conn()); err != nil {
		logger.Fatal().Stack().Err(err).Msg("migrate")
	}

	logger.Info().Msg("starting application")

	if err = app.Run(ctx, cfg, pool); err != nil {
		logger.Fatal().Stack().Err(err).Msg("run app")
	}

	logger.Info().Msg("shutting down")
}
