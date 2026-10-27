package usecase

import (
	"context"
	"time"

	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/inbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/outbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/timeutils"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

type RateUseCase struct {
	cacheRepositroy outbound.RateCacheRepository
	rateService     inbound.RateService
}

func NewRateUseCase(
	cacheRepositroy outbound.RateCacheRepository,
	rateService inbound.RateService,
) *RateUseCase {
	return &RateUseCase{
		cacheRepositroy: cacheRepositroy,
		rateService:     rateService,
	}
}

func (e *RateUseCase) CollectRate(ctx context.Context, symbol values.Symbol) error {
	logger := zerolog.Ctx(ctx).
		With().
		Str("name", "use_case.collect_rate").
		Any("symbol", symbol).
		Logger()

	rate, provider, err := e.rateService.Current(ctx, entities.NewPairFromSymbol(symbol))
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

func (e *RateUseCase) GetRate(ctx context.Context, symbol values.Symbol, date time.Time) (*entities.Rate, error) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("name", "use_case.get_rate").
		Any("symbol", symbol).
		Logger()

	pairEntity := entities.NewPairFromSymbol(symbol)
	logger.Debug().Msg("get rate")

	ts := timeutils.RoundToNearest(date, timeutils.FiveMinuteBucket)

	if ts.Compare(timeutils.RoundToNearest(time.Now(), timeutils.FiveMinuteBucket)) < 0 {
		rate, err := e.rateService.Historical(ctx, pairEntity, ts)
		if err != nil {
			logger.Error().
				Any("err", eris.ToJSON(err, true)).
				Msg("get historical rate")

			return nil, err
		}

		return rate, nil
	}

	rate, err := e.cacheRepositroy.GetRate(ctx, pairEntity)
	if err != nil {
		logger.Error().
			Any("err", eris.ToJSON(err, true)).
			Msg("cant get rate value from cache")
	}

	if rate != nil {
		return rate, nil
	}

	rate, provider, err := e.rateService.Current(ctx, pairEntity)
	if err != nil {
		logger.Error().
			Any("err", eris.ToJSON(err, true)).
			Any("provider", provider).
			Msg("cant get rate")
	}

	if err = e.cacheRepositroy.PutRate(ctx, *rate); nil != err {
		logger.Error().
			Any("err", eris.ToJSON(err, true)).
			Any("provider", provider).
			Msg("cant put rate value into cache")

		return nil, err
	}

	return rate, nil
}
