package dto

import "github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"

type CoinGeckoID string

func (i CoinGeckoID) String() string {
	return string(i)
}

var mapping = map[values.CurrencyISO]CoinGeckoID{ // nolint:gochecknoglobals
	"BTC": "bitcoin",
	"ETH": "ethereum",
}

func GetCoinID(iso values.CurrencyISO) (CoinGeckoID, bool) {
	id, ok := mapping[iso.ToUpper()]

	return id, ok
}
