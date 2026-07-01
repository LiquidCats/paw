package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/LiquidCats/paw/lib/graceful"
	"github.com/LiquidCats/paw/services/litehsm/configs"
	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/keychain"
	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/postgresql"
	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/postgresql/database"
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

	migrationConn, err := pool.Acquire(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "acquire migration connection", slog.String("err", err.Error()))

		return
	}

	if err = database.Migrate(ctx, migrationConn.Conn()); err != nil {
		logger.ErrorContext(ctx, "migration connection", slog.String("err", err.Error()))

		return
	}

	migrationConn.Release()

	queries := database.New(pool)
	tx := postgresql.NewTxManager(pool)
	db := postgresql.NewRepository(queries, tx)

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

	keyManagerCreateKey := hsm.NewKeyManagerCreateKey(db)
	keyManagerSetExpiration := hsm.NewKeyManagerSetExpiration(db)
	keyManagerSetStatus := hsm.NewKeyManagerSetStatus(db)

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
