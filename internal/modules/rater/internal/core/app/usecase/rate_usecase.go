package usecase

import (
	"context"
	"time"

	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"

	kernelvalues "github.com/LiquidCats/paw/internal/kernel/domain/values"
	"github.com/LiquidCats/paw/internal/kernel/ports"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/inbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/outbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

type RateUseCase struct {
	logger          ports.Logger
	cacheRepositroy outbound.RateCacheRepository
	rateService     inbound.RateService
	clock           ports.Clock
}

func NewRateUseCase(
	logger ports.Logger,
	clock ports.Clock,
	cacheRepositroy outbound.RateCacheRepository,
	rateService inbound.RateService,
) *RateUseCase {
	return &RateUseCase{
		logger:          logger,
		clock:           clock,
		cacheRepositroy: cacheRepositroy,
		rateService:     rateService,
	}
}

func (uc *RateUseCase) CollectRate(ctx context.Context, symbol values.Symbol) error {
	logger := zerolog.Ctx(ctx).
		With().
		Str("name", "use_case.collect_rate").
		Any("symbol", symbol).
		Logger()

	rate, provider, err := uc.rateService.Current(ctx, entities.NewPairFromSymbol(symbol))
	if err != nil {
		logger.Error().
			Any("err", eris.ToJSON(err, true)).
			Any("provider", provider).
			Msg("cant get rate")

		return err
	}

	logger.Debug().
		Any("rate_entry", rate).
		Any("provider", provider).
		Msg("rate saved")

	return nil
}

func (uc *RateUseCase) GetRate(ctx context.Context, symbol values.Symbol, date time.Time) (*entities.Rate, error) {
	logger := uc.logger.With(
		ports.Str("name", "use_case.get_rate"),
		ports.Of[string]("symbol", symbol.String()),
	)

	pairEntity := entities.NewPairFromSymbol(symbol)
	logger.Debug("get rate")

	ts := uc.clock.RoundToNearest(date, kernelvalues.FiveMinuteBucket)

	if ts.Compare(uc.clock.RoundToNearest(time.Now(), kernelvalues.FiveMinuteBucket)) < 0 {
		rate, err := uc.rateService.Historical(ctx, pairEntity, ts)
		if err != nil {
			logger.Error("get historical rate", ports.Err(err))

			return nil, err
		}

		return rate, nil
	}

	rate, err := uc.cacheRepositroy.GetRate(ctx, pairEntity)
	if err != nil && !eris.Is(err, entities.ErrRateNotFound) {
		logger.Error("cant get rate value from cache", ports.Err(err))
	}

	if rate != nil {
		return rate, nil
	}

	rate, provider, err := uc.rateService.Current(ctx, pairEntity)
	if err != nil {
		logger.Error("cant get rate", ports.Err(err), ports.Any("provider", provider))

		return nil, err
	}

	if err = uc.cacheRepositroy.PutRate(ctx, *rate); nil != err {
		logger.Error(
			"cant put rate value into cache",
			ports.Err(err),
			ports.Any("provider", provider),
		)

		return nil, err
	}

	return rate, nil
}
