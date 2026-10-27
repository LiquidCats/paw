package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/inbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/outbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/timeutils"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/errors"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

var _ inbound.RateService = (*RateService)(nil)

type RateService struct {
	apiRegsitry inbound.APIRepositroyRegistry
	rateDB      outbound.RateDatabaseRepository
}

func NewRateService(
	apiRegsitry inbound.APIRepositroyRegistry,
	rateDB outbound.RateDatabaseRepository,
) *RateService {
	return &RateService{
		apiRegsitry: apiRegsitry,
		rateDB:      rateDB,
	}
}

// Current implements [inbound.RateService].
func (s *RateService) Current(ctx context.Context, pair entities.Pair) (*entities.Rate, values.ProviderName, error) {
	var (
		err      error
		price    decimal.Decimal
		provider values.ProviderName
	)
	start := time.Now()

	logger := zerolog.Ctx(ctx).
		With().
		Str("name", "service.rate.current").
		Any("pair", pair).
		Logger()

	s.apiRegsitry.Iterate(func(key values.ProviderName, value outbound.RateAPIRepository) {
		price, err = value.GetRate(ctx, pair)
		provider = key
		if err != nil {
			var providerErr *errors.ProviderRequestFailedError
			if eris.As(err, &providerErr) {
				logger = logger.With().
					Int("provider_err_code", providerErr.StatusCode).
					Str("provider_err_body", providerErr.Body).
					Logger()
			}

			logger.Error().
				Any("err", eris.ToJSON(err, true)).
				Any("provider", key).
				Msg("cant get rate")
		}

		if !price.IsZero() {
			return
		}
	})

	if price.IsZero() {
		logger.Error().
			Any("err", eris.ToJSON(err, true)).
			Any("provider", provider).
			Msg("rate not available")
		return nil, provider, errors.ErrRateNotAvailable
	}

	rate := entities.Rate{
		Pair:     pair,
		Price:    price,
		Provider: provider,
	}

	logger.Debug().
		Any("rate", rate).
		Msg("get historical rate")

	hasRate, err := s.rateDB.HasRate(ctx, rate.Pair.Symbol, timeutils.RoundToNearest(start, timeutils.FiveMinuteBucket))

	if err == nil && hasRate {
		return &rate, provider, nil
	}

	if err != nil {
		logger.Warn().
			Any("err", eris.ToJSON(err, true)).
			Msg("cant check rate in database")
		return &rate, provider, nil //nolint:nilerr
	}

	_, err = s.rateDB.SaveRate(ctx, rate, timeutils.RoundToNearest(start, timeutils.FiveMinuteBucket))

	if err != nil {
		logger.Error().
			Any("err", eris.ToJSON(err, true)).
			Any("provider", provider).
			Any("rate", rate).
			Msg("cant save rate to database")
	}

	return &rate, provider, nil
}

// Historical implements [inbound.RateService].
func (s *RateService) Historical(ctx context.Context, pair entities.Pair, ts time.Time) (*entities.Rate, error) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("name", "service.rate.historical").
		Any("pair", pair).
		Logger()

	historical, err := s.rateDB.GetRate(ctx, pair.Symbol, ts)
	if err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			logger.Error().
				Any("err", eris.ToJSON(err, true)).
				Msg("no historical rate found")
			return nil, errors.ErrNoHistoricalRate
		}
		logger.Error().
			Any("err", eris.ToJSON(err, true)).
			Msg("get historical rate")
		return nil, eris.Wrap(err, "cant get rate from database")
	}

	logger.Debug().
		Any("rate", historical).
		Msg("get historical rate")

	return historical, nil
}
