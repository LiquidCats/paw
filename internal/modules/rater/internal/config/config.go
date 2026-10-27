package config

import (
	"github.com/go-playground/sensitive"
)

type Config struct {
	App AppConfig `envconfig:"APP"`

	Providers ProviderConfig `envconfig:"PROVIDERS"`
}

type AppConfig struct {
	CollectSchedule string `envconfig:"COLLECT_SCHEDULE" default:"*/5 * * * *"`
}

type ProviderConfig struct {
	Cex       CexConfig       `envconfig:"CEX"`
	CoinGate  CoinGateConfig  `envconfig:"COIN_GATE"`
	CoinGecko CoinGeckoConfig `envconfig:"COIN_GECKO"`

	CoinApi       CoinApiConfig       `envconfig:"COIN_API"`
	CoinMarketCap CoinMarketCapConfig `envconfig:"COIN_MARKET_CAP"`
}

type CexConfig struct {
	URL string `envconfig:"URL"`
}

type CoinApiConfig struct { // nolint:revive
	URL    string           `envconfig:"URL"`
	Secret sensitive.String `envconfig:"SECRET"`
}

type CoinGateConfig struct {
	URL string `envconfig:"URL"`
}

type CoinGeckoConfig struct {
	URL string `envconfig:"URL"`
}

type CoinMarketCapConfig struct {
	URL    string           `envconfig:"URL"`
	Secret sensitive.String `envconfig:"SECRET"`
}
