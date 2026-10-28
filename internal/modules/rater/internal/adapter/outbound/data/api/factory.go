package api

import (
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/cex"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coinapi"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coingate"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coingecko"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coinmarketcap"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/config"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/outbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

func Factory(provider values.ProviderName, cfg config.ProviderConfig) (outbound.RateAPIRepository, error) {
	switch provider {
	case values.ProviderNameCex:
		return cex.NewRepository(cfg), nil
	case values.ProviderNameCoinApi:
		return coinapi.NewRepository(cfg), nil
	case values.ProviderNameCoinGate:
		return coingate.NewRepository(cfg), nil
	case values.ProviderNameCoinGecko:
		return coingecko.NewRepository(cfg), nil
	case values.ProviderNameCoinMarketCap:
		return coinmarketcap.NewRepository(cfg), nil
	default:
		return nil, values.ErrUnsupportedProvider
	}
}
