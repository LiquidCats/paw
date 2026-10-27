package cron

import (
	"context"
	"time"

	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/config"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/outbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/usecase"
)

type CollectRateTask struct {
	logger zerolog.Logger

	cfg     config.AppConfig
	pairDB  outbound.PairDatabaseRepository
	useCase *usecase.RateUseCase
}

func NewCollectRateTask(
	logger zerolog.Logger,
	cfg config.AppConfig,
	pairDB outbound.PairDatabaseRepository,
	useCase *usecase.RateUseCase,
) *CollectRateTask {
	return &CollectRateTask{
		logger: logger,

		cfg:     cfg,
		pairDB:  pairDB,
		useCase: useCase,
	}
}

func (t *CollectRateTask) Spec() string {
	return t.cfg.CollectSchedule
}

func (t *CollectRateTask) Run() {
	t.logger.Info().Msg("start collect rate task")
	defer t.logger.Info().Msg("end collect rate task")

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute) //nolint:mnd
	defer cancel()

	pairs, err := t.pairDB.GetAllPairs(ctx)
	if err != nil {
		t.logger.Error().
			Any("err", eris.ToJSON(err, true)).
			Msg("invalid pairs")
		return
	}

	for _, pair := range pairs {
		if err = t.useCase.CollectRate(ctx, pair.Symbol); err != nil {
			t.logger.Error().
				Any("err", eris.ToJSON(err, true)).
				Str("symbol", string(pair.Symbol)).
				Msg("invalid rate")
		}
	}
}
