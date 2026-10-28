package values

import "github.com/rotisserie/eris"

type ProviderName string

var ErrUnsupportedProvider = eris.New("unsupported provider")

const (
	ProviderNameCex           ProviderName = "cex"
	ProviderNameCoinApi       ProviderName = "coinapi"
	ProviderNameCoinGate      ProviderName = "coingate"
	ProviderNameCoinGecko     ProviderName = "coingecko"
	ProviderNameCoinMarketCap ProviderName = "coinmarketcap"
)

func (p ProviderName) String() string {
	return string(p)
}
