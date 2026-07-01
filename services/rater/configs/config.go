package configs

type Config struct {
	App   AppConfig   `yaml:"app" envconfig:"APP"`
	Redis RedisConfig `yaml:"redis" envconfig:"REDIS"`
	DB    DB          `yaml:"db" envconfig:"DB"`

	CoinGate      CoinGateConfig      `yaml:"coingate" envconfig:"COIN_GATE"`
	Cex           CexConfig           `yaml:"cex" envconfig:"CEX"`
	CoinApi       CoinApiConfig       `yaml:"coin_api" envconfig:"COIN_API"` // nolint:revive
	CoinMarketCap CoinMarketCapConfig `yaml:"coin_market_cap" envconfig:"COIN_MARKET_CAP"`
	CoinGecko     CoinGeckoConfig     `yaml:"coin_gecko" envconfig:"COIN_GECKO"`
}
