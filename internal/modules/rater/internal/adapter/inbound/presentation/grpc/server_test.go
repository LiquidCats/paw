package grpc_test

import (
	"io"
	"testing"
	"time"

	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	kernelvalues "github.com/LiquidCats/paw/internal/kernel/domain/values"
	ratergrpc "github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/inbound/presentation/grpc"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/usecase"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/tests/mocks"
	"github.com/LiquidCats/paw/internal/shares/logs"
	pawv1 "github.com/LiquidCats/paw/pkg/paw/v1"
	kernelmocks "github.com/LiquidCats/paw/tests/mocks/kernelmocks"
)

type fixture struct {
	srv     *ratergrpc.RaterService
	clock   *kernelmocks.MockClock
	cache   *mocks.MockRateCacheRepository
	service *mocks.MockRateService
}

func newFixture(t *testing.T) fixture {
	t.Helper()

	clock := kernelmocks.NewMockClock(t)
	cache := mocks.NewMockRateCacheRepository(t)
	service := mocks.NewMockRateService(t)
	uc := usecase.NewRateUseCase(logs.NewSlogAdapter(io.Discard), clock, cache, service)

	return fixture{srv: ratergrpc.NewServer(uc), clock: clock, cache: cache, service: service}
}

func TestRaterService_GetRate(t *testing.T) {
	t.Parallel()

	pair := entities.NewPairFromSymbol("BTC_USD")
	rate := &entities.Rate{Pair: pair, Price: decimal.RequireFromString("123.45")}

	t.Run("today is served from cache and mapped", func(t *testing.T) {
		t.Parallel()
		f := newFixture(t)
		now := time.Now()
		f.clock.EXPECT().RoundToNearest(mock.Anything, kernelvalues.FiveMinuteBucket).Return(now).Twice()
		f.cache.EXPECT().GetRate(t.Context(), pair).Return(rate, nil).Once()

		got, err := f.srv.GetRate(t.Context(), &pawv1.GetRateRequest{Symbol: "btc_usd"})

		require.NoError(t, err)
		assert.Equal(t, &pawv1.GetRateResponse{
			Pair:  &pawv1.Pair{From: "BTC", To: "USD", Symbol: "BTC_USD"},
			Price: "123.45",
		}, got)
	})

	t.Run("explicit date is historical", func(t *testing.T) {
		t.Parallel()
		f := newFixture(t)
		date := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
		f.clock.EXPECT().RoundToNearest(date, kernelvalues.FiveMinuteBucket).Return(date).Once()
		f.clock.EXPECT().RoundToNearest(mock.Anything, kernelvalues.FiveMinuteBucket).Return(time.Now()).Once()
		f.service.EXPECT().Historical(t.Context(), pair, date).Return(rate, nil).Once()

		got, err := f.srv.GetRate(t.Context(), &pawv1.GetRateRequest{Symbol: "BTC_USD", Date: proto.String("2020-01-02")})

		require.NoError(t, err)
		assert.Equal(t, "123.45", got.GetPrice())
	})

	t.Run("bad date is invalid argument with field violation", func(t *testing.T) {
		t.Parallel()
		f := newFixture(t)

		got, err := f.srv.GetRate(t.Context(), &pawv1.GetRateRequest{Symbol: "BTC_USD", Date: proto.String("02/01/2020")})

		assert.Nil(t, got)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		require.Len(t, st.Details(), 1)
		detail, ok := st.Details()[0].(*epb.BadRequest)
		require.True(t, ok)
		assert.Equal(t, "date", detail.GetFieldViolations()[0].GetField())
	})

	t.Run("use case failure is internal", func(t *testing.T) {
		t.Parallel()
		f := newFixture(t)
		date := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
		f.clock.EXPECT().RoundToNearest(date, kernelvalues.FiveMinuteBucket).Return(date).Once()
		f.clock.EXPECT().RoundToNearest(mock.Anything, kernelvalues.FiveMinuteBucket).Return(time.Now()).Once()
		f.service.EXPECT().Historical(t.Context(), pair, date).Return(nil, eris.New("boom")).Once()

		got, err := f.srv.GetRate(t.Context(), &pawv1.GetRateRequest{Symbol: "BTC_USD", Date: proto.String("2020-01-02")})

		assert.Nil(t, got)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestRaterService_AttachToGRPC(t *testing.T) {
	t.Parallel()

	server := grpc.NewServer()
	newFixture(t).srv.AttachToGRPC(server)

	assert.Contains(t, server.GetServiceInfo(), pawv1.RaterService_ServiceDesc.ServiceName)
}
