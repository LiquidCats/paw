package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

const defaultRateKeyExpiration = 5 * time.Second

type RateKey struct {
	From string
	To   string
}

func newRateKey(pair entities.Pair) RateKey {
	return RateKey{
		From: pair.From.String(),
		To:   pair.To.String(),
	}
}

func (p RateKey) String() string {
	return strings.ToLower(
		fmt.Sprintf(
			"rate:from:%s:to:%s",
			p.From,
			p.To,
		),
	)
}


func (r *Repository) GetRate(ctx context.Context, pair entities.Pair) (*entities.Rate, error) {
	key := newRateKey(pair)
	price, err := r.client.Get(ctx, key.String()).Result()
	if err != nil  {
		return nil, eris.Wrap(err, "failed to get rate from redis")
	}

	return &entities.Rate{
		Pair:  pair,
		Price: decimal.RequireFromString(price),
	}, nil
}

func (r *Repository) PutRate(ctx context.Context, rate entities.Rate) error {
	key := newRateKey(rate.Pair)
	_, err := r.client.Set(ctx, key.String(), rate.Price.String(), defaultRateKeyExpiration).Result()
	if err != nil {
		return eris.Wrap(err, "failed to put rate to redis")
	}

	return nil
}
