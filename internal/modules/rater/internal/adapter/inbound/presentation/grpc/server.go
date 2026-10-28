package grpc

import (
	"context"
	"time"

	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/LiquidCats/libraries/graceful"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/usecase"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
	pawv1 "github.com/LiquidCats/paw/pkg/paw/v1"
)

var _ graceful.GRPCAttacher = (*RaterService)(nil)

type RaterService struct {
	rate *usecase.RateUseCase

	pawv1.UnimplementedRaterServiceServer
}

// AttachToGRPC implements [graceful.GRPCAttacher].
func (r *RaterService) AttachToGRPC(reg grpc.ServiceRegistrar) {
	pawv1.RegisterRaterServiceServer(reg, r)
}

func NewServer(
	rate *usecase.RateUseCase,
) *RaterService {
	return &RaterService{
		rate: rate,
	}
}

func (srv *RaterService) GetRate(ctx context.Context, req *pawv1.GetRateRequest) (*pawv1.GetRateResponse, error) {
	var err error
	symbolStr := req.GetSymbol()
	dateStr := req.GetDate()

	date := time.Now()
	if dateStr != "" {
		date, err = time.Parse(time.DateOnly, dateStr)
		if err != nil {
			st, stErr := status.New(codes.InvalidArgument, "invalid field").
				WithDetails(&epb.BadRequest{
					FieldViolations: []*epb.BadRequest_FieldViolation{
						{
							Field:       "date",
							Description: "invalid date format",
							Reason:      "date must be in YYYY-MM-DD format",
						},
					},
				})
			if stErr != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}

			return nil, st.Err()
		}
	}

	symbol := values.NewSymbolFromString(symbolStr)

	result, err := srv.rate.GetRate(ctx, symbol, date)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pawv1.GetRateResponse{
		Pair: &pawv1.Pair{
			From:   result.Pair.From.ToUpper().String(),
			To:     result.Pair.To.ToUpper().String(),
			Symbol: result.Pair.Symbol.ToUpper().String(),
		},
		Price: result.Price.String(),
	}, nil
}
