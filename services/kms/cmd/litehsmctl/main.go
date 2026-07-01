package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/passphrase"
	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/sealer"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/usecase/terminal"
	"github.com/LiquidCats/paw/services/litehsm/internal/bootstrap"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))

	initCmdUseCase, err := terminal.NewInitCommand(
		sealer.NewDefault(bootstrap.AppMagic),
		new(passphrase.StdInPassphraseProvider),
		new(passphrase.EnvPassphraseProvider),
		new(passphrase.FilePassphraseProvider),
	)
	if err != nil {
		logger.Error("create init command", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		logger.Error(fmt.Sprintf("expected '%s' subcommands", initCmdUseCase.Name()))
		os.Exit(1)
	}

	switch os.Args[1] {
	case initCmdUseCase.Name():
		if err := initCmdUseCase.Run(); err != nil {
			logger.Error("initialisation unsuccessful", slog.String("error", err.Error()))
			os.Exit(1)
		}

		logger.Info("initialised successfully")

		os.Exit(0)
	default:
		logger.Error(fmt.Sprintf("unknown command: %s\n", os.Args[1]))

		os.Exit(2)
	}
}
