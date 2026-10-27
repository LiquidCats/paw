package grpc

import (
	"github.com/rs/zerolog"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/usecase"
)

type RaterService struct {
	logger zerolog.Logger
	rate   *usecase.RateUseCase
}

func NewServer(
	logger zerolog.Logger,
	rate *usecase.RateUseCase,
) *RaterService {
	return &RaterService{
		logger: logger,
		rate:   rate,
	}
}
