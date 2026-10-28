package usecase_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	kernelvalues "github.com/LiquidCats/paw/internal/kernel/domain/values"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/usecase"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/tests/mocks"
	"github.com/LiquidCats/paw/internal/shares/logs"
	kernelmocks "github.com/LiquidCats/paw/tests/mocks/kernelmocks"
)

type RateUseCaseSuite struct {
	suite.Suite

	ctx    context.Context
	symbol values.Symbol
	pair   entities.Pair
	rate   *entities.Rate
	date   time.Time

	clock   *kernelmocks.MockClock
	cache   *mocks.MockRateCacheRepository
	service *mocks.MockRateService
	uc      *usecase.RateUseCase
}

func TestRateUseCaseSuite(t *testing.T) {
	suite.Run(t, new(RateUseCaseSuite))
}

func (s *RateUseCaseSuite) SetupTest() {
	s.ctx = context.Background()
	s.symbol = values.Symbol("BTC_USD")
	s.pair = entities.NewPairFromSymbol(s.symbol)
	s.rate = &entities.Rate{Pair: s.pair, Price: decimal.NewFromInt(10), Provider: values.ProviderNameCex}
	s.date = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	s.clock = kernelmocks.NewMockClock(s.T())
	s.cache = mocks.NewMockRateCacheRepository(s.T())
	s.service = mocks.NewMockRateService(s.T())
	s.uc = usecase.NewRateUseCase(logs.NewSlogAdapter(io.Discard), s.clock, s.cache, s.service)
}

// roundTo pins the two clock calls GetRate makes: the requested date and "now".
func (s *RateUseCaseSuite) roundTo(requested, now time.Time) {
	s.clock.EXPECT().RoundToNearest(s.date, kernelvalues.FiveMinuteBucket).Return(requested).Once()
	s.clock.EXPECT().RoundToNearest(mock.Anything, kernelvalues.FiveMinuteBucket).Return(now).Once()
}

func (s *RateUseCaseSuite) TestCollectRate_Success() {
	s.service.EXPECT().Current(s.ctx, s.pair).Return(s.rate, values.ProviderNameCex, nil).Once()

	s.NoError(s.uc.CollectRate(s.ctx, s.symbol))
}

func (s *RateUseCaseSuite) TestCollectRate_ServiceError() {
	boom := eris.New("boom")
	s.service.EXPECT().Current(s.ctx, s.pair).Return(nil, values.ProviderNameCex, boom).Once()

	s.ErrorIs(s.uc.CollectRate(s.ctx, s.symbol), boom)
}

func (s *RateUseCaseSuite) TestGetRate_Historical() {
	past := s.date.Add(-time.Hour)
	s.roundTo(past, s.date)
	s.service.EXPECT().Historical(s.ctx, s.pair, past).Return(s.rate, nil).Once()

	got, err := s.uc.GetRate(s.ctx, s.symbol, s.date)

	s.Require().NoError(err)
	s.Equal(s.rate, got)
}

func (s *RateUseCaseSuite) TestGetRate_HistoricalError() {
	past := s.date.Add(-time.Hour)
	boom := eris.New("boom")
	s.roundTo(past, s.date)
	s.service.EXPECT().Historical(s.ctx, s.pair, past).Return(nil, boom).Once()

	got, err := s.uc.GetRate(s.ctx, s.symbol, s.date)

	s.Nil(got)
	s.ErrorIs(err, boom)
}

func (s *RateUseCaseSuite) TestGetRate_CacheHit() {
	s.roundTo(s.date, s.date)
	s.cache.EXPECT().GetRate(s.ctx, s.pair).Return(s.rate, nil).Once()

	got, err := s.uc.GetRate(s.ctx, s.symbol, s.date)

	s.Require().NoError(err)
	s.Equal(s.rate, got)
}

func (s *RateUseCaseSuite) TestGetRate_CacheMissFetchesAndCaches() {
	cacheErrors := map[string]error{
		"not found":     entities.ErrRateNotFound,
		"cache failure": eris.New("redis down"),
	}

	for name, cacheErr := range cacheErrors {
		s.Run(name, func() {
			s.SetupTest()
			s.roundTo(s.date, s.date)
			s.cache.EXPECT().GetRate(s.ctx, s.pair).Return(nil, cacheErr).Once()
			s.service.EXPECT().Current(s.ctx, s.pair).Return(s.rate, values.ProviderNameCex, nil).Once()
			s.cache.EXPECT().PutRate(s.ctx, *s.rate).Return(nil).Once()

			got, err := s.uc.GetRate(s.ctx, s.symbol, s.date)

			s.Require().NoError(err)
			s.Equal(s.rate, got)
		})
	}
}

func (s *RateUseCaseSuite) TestGetRate_CurrentError() {
	boom := eris.New("boom")
	s.roundTo(s.date, s.date)
	s.cache.EXPECT().GetRate(s.ctx, s.pair).Return(nil, entities.ErrRateNotFound).Once()
	s.service.EXPECT().Current(s.ctx, s.pair).Return(nil, values.ProviderNameCex, boom).Once()

	got, err := s.uc.GetRate(s.ctx, s.symbol, s.date)

	s.Nil(got)
	s.ErrorIs(err, boom)
}

func (s *RateUseCaseSuite) TestGetRate_PutRateError() {
	boom := eris.New("boom")
	s.roundTo(s.date, s.date)
	s.cache.EXPECT().GetRate(s.ctx, s.pair).Return(nil, entities.ErrRateNotFound).Once()
	s.service.EXPECT().Current(s.ctx, s.pair).Return(s.rate, values.ProviderNameCex, nil).Once()
	s.cache.EXPECT().PutRate(s.ctx, *s.rate).Return(boom).Once()

	got, err := s.uc.GetRate(s.ctx, s.symbol, s.date)

	s.Nil(got)
	s.ErrorIs(err, boom)
}
