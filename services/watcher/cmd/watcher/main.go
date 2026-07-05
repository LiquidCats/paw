package main

import (
	"context"
	"os"

	migrator "github.com/LiquidCats/paw/lib/database"
	"github.com/LiquidCats/paw/services/watcher/configs"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/postgresql/database/migrations"
	"github.com/LiquidCats/paw/services/watcher/internal/app"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	_ "github.com/lib/pq"
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

	if err = migrator.MigrateUp(ctx, pool, migrations.FS); err != nil {
		logger.Fatal().Stack().Err(err).Msg("migrate")
	}

	logger.Info().Msg("starting application")

	if err = app.Run(ctx, cfg, pool); err != nil {
		logger.Fatal().Stack().Err(err).Msg("run app")
	}

	logger.Info().Msg("shutting down")
}
