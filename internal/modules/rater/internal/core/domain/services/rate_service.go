package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"

	kernelvalues "github.com/LiquidCats/paw/internal/kernel/domain/values"
	"github.com/LiquidCats/paw/internal/kernel/ports"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/inbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/outbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/errors"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

var _ inbound.RateService = (*RateService)(nil)

type RateService struct {
	logger      ports.Logger
	clock       ports.Clock
	apiRegsitry inbound.APIRepositroyRegistry
	rateDB      outbound.RateDatabaseRepository
}

func NewRateService(
	logger ports.Logger,
	clock ports.Clock,
	apiRegsitry inbound.APIRepositroyRegistry,
	rateDB outbound.RateDatabaseRepository,
) *RateService {
	return &RateService{
		logger:      logger,
		clock:       clock,
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

	logger := s.logger.
		With(
			ports.Any("pair", pair),
		)

	s.apiRegsitry.Iterate(func(key values.ProviderName, value outbound.RateAPIRepository) {
		price, err = value.GetRate(ctx, pair)
		provider = key
		if err != nil {
			var providerErr *errors.ProviderRequestFailedError
			if eris.As(err, &providerErr) {
				logger = logger.With(
					ports.Int("provider_err_code", providerErr.StatusCode),
					ports.Str("provider_err_body", providerErr.Body),
				)
			}

			logger.Error("cant get rate",
				ports.Any("err", eris.ToJSON(err, true)),
				ports.Any("provider", key),
			)
		}

		if !price.IsZero() {
			return
		}
	})

	if price.IsZero() {
		logger.Error(
			"rate not available",
			ports.Err(err),
			ports.Stringer("provider", provider),
		)
		return nil, provider, errors.ErrRateNotAvailable
	}

	rate := entities.Rate{
		Pair:     pair,
		Price:    price,
		Provider: provider,
	}

	logger.Debug("get historical rate", ports.Any("rate", rate))
	hasRate, err := s.rateDB.HasRate(ctx, rate.Pair.Symbol, s.clock.RoundToNearest(start, kernelvalues.FiveMinuteBucket))

	if err == nil && hasRate {
		return &rate, provider, nil
	}

	if err != nil {
		logger.Warn("cant check rate in database", ports.Err(err))

		return &rate, provider, nil //nolint:nilerr
	}

	_, err = s.rateDB.SaveRate(ctx, rate, s.clock.RoundToNearest(start, kernelvalues.FiveMinuteBucket))

	if err != nil {
		logger.Error(
			"cant save rate to database",
			ports.Err(err),
			ports.Stringer("provider", provider),
			ports.Any("rate", rate),
		)
	}

	return &rate, provider, nil
}

// Historical implements [inbound.RateService].
func (s *RateService) Historical(ctx context.Context, pair entities.Pair, ts time.Time) (*entities.Rate, error) {
	logger := s.logger.
		With(
			ports.Any("pair", pair),
		)

	historical, err := s.rateDB.GetRate(ctx, pair.Symbol, ts)
	if err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			logger.Error(
				"no historical rate found",
				ports.Err(err),
			)

			return nil, errors.ErrNoHistoricalRate
		}

		logger.Error("get historical rate", ports.Err(err))

		return nil, eris.Wrap(err, "cant get rate from database")
	}

	logger.Debug("get historical rate", ports.Any("rate", historical))

	return historical, nil
}
