package api

import (
	"github.com/LiquidCats/paw/services/rater/configs"
	"github.com/LiquidCats/paw/services/rater/internal/adapter/repository/api/cex"
	"github.com/LiquidCats/paw/services/rater/internal/adapter/repository/api/coinapi"
	"github.com/LiquidCats/paw/services/rater/internal/adapter/repository/api/coingate"
	"github.com/LiquidCats/paw/services/rater/internal/adapter/repository/api/coingecko"
	"github.com/LiquidCats/paw/services/rater/internal/adapter/repository/api/coinmarketcap"
	"github.com/LiquidCats/paw/services/rater/internal/app/domain/entity"
	"github.com/LiquidCats/paw/services/rater/internal/app/port/adapter/repository"
)

func ProviderFactory(cfg configs.Config, provider entity.ProviderName) repository.RateAPI {
	switch provider {
	case entity.ProviderNameCex:
		return cex.NewRepository(cfg.Cex)
	case entity.ProviderNameCoinApi:
		return coinapi.NewRepository(cfg.CoinApi)
	case entity.ProviderNameCoinGate:
		return coingate.NewRepository(cfg.CoinGate)
	case entity.ProviderNameCoinGecko:
		return coingecko.NewRepository(cfg.CoinGecko)
	case entity.ProviderNameCoinMarketCap:
		return coinmarketcap.NewReposiotry(cfg.CoinMarketCap)
	}

	return nil
}
