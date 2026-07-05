package main

import (
	"context"
	"log/slog"
	"os"

	db "github.com/LiquidCats/paw/lib/database"
	"github.com/LiquidCats/paw/lib/graceful"
	"github.com/LiquidCats/paw/services/litehsm/configs"
	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/keychain"
	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/postgresql"
	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/postgresql/database"
	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/postgresql/database/migrations"
	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/sealer"
	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/transport/grpc"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/usecase/hsm"
	"github.com/LiquidCats/paw/services/litehsm/internal/bootstrap"
	"github.com/awnumar/memguard"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	memguard.CatchInterrupt()
	defer memguard.Purge()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	cfg, err := configs.Load(bootstrap.AppName)
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     cfg.Common.Logging.Level,
	}))
	slog.SetDefault(logger)

	poolConfig, err := pgxpool.ParseConfig(cfg.Common.DB.ToDSN())
	if err != nil {
		logger.ErrorContext(ctx, "create db pool", slog.String("err", err.Error()))

		return
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.ErrorContext(ctx, "create db pool", slog.String("err", err.Error()))

		return
	}
	defer pool.Close()

	if err = db.MigrateUp(ctx, pool, migrations.FS); err != nil {
		logger.ErrorContext(ctx, "migrate db", slog.String("err", err.Error()))

		return
	}

	queries := database.New(pool)
	txManager := db.NewTxManager(pool)
	queriesManager := db.NewQueriesTxManager[database.Queries](txManager, queries)

	repo := postgresql.NewRepository(queriesManager)

	seal := sealer.NewDefault(bootstrap.AppMagic)

	env, err := entities.EnvelopeFromBuffer(cfg.App.KeyManager.Seed.Sealing.LockedBuffer)
	if err != nil {
		logger.ErrorContext(ctx, "failed to load seed key", slog.String("error", err.Error()))

		return
	}

	seed, err := seal.Unseal(env, cfg.App.KeyManager.Seed.Passphrase.Enclave)
	if err != nil {
		logger.ErrorContext(ctx, "failed to load seed key", slog.String("error", err.Error()))

		return
	}

	kch, err := keychain.NewSecp256k1Keychain(seed)
	if err != nil {
		logger.ErrorContext(ctx, "failed to load seed key", slog.String("error", err.Error()))

		return
	}

	_ = kch

	keyManagerCreateKey := hsm.NewKeyManagerCreateKey(repo)
	keyManagerSetExpiration := hsm.NewKeyManagerSetExpiration(repo)
	keyManagerSetStatus := hsm.NewKeyManagerSetStatus(repo)

	var runners []graceful.Runner

	grpcServer := grpc.NewKeyManagerServiceServer(
		keyManagerCreateKey,
		keyManagerSetExpiration,
		keyManagerSetStatus,
	)

	runners = append(
		runners,
		graceful.GRPCRunner(
			grpcServer,
			graceful.WithGRPCPort(cfg.Common.GRPC.Port),
			graceful.WithConnectionTimeout(cfg.Common.GRPC.ConnTimeout),
		),
	)

	if err := graceful.WaitContext(ctx, runners...); err != nil {
		logger.Error("application terminated", slog.String("err", err.Error()))

		return
	}

	logger.Info("shutting down")
}
