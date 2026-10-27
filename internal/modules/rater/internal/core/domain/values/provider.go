package values

import (
	"encoding/json/v2"

	"github.com/rotisserie/eris"
)

type ProviderName int

var (
	ErrInvalidProviderName = eris.New("invalid provider name")
)

const (
	ProviderNameCex           ProviderName = iota
	ProviderNameCoinApi
	ProviderNameCoinGate
	ProviderNameCoinGecko
	ProviderNameCoinMarketCap
)

var providerNameMap = map[ProviderName]string{
	ProviderNameCex:           "cex",
	ProviderNameCoinApi:       "coinapi",
	ProviderNameCoinGate:      "coingate",
	ProviderNameCoinGecko:     "coingecko",
	ProviderNameCoinMarketCap: "coinmarketcap",
}

var providerNameMapReverse = map[string]ProviderName{
	"cex":           ProviderNameCex,
	"coinapi":       ProviderNameCoinApi,
	"coingate":      ProviderNameCoinGate,
	"coingecko":     ProviderNameCoinGecko,
	"coinmarketcap": ProviderNameCoinMarketCap,
}

func (p ProviderName) MarshalJSON() ([]byte, error) {
	return json.Marshal(providerNameMap[p])
}

func (p *ProviderName) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	pn, ok := providerNameMapReverse[s]
	if !ok {
		return ErrInvalidProviderName
	}

	*p = pn

	return nil
}

func (p ProviderName) MarshalText() ([]byte, error) {
	return []byte(providerNameMap[p]), nil
}

func (p *ProviderName) UnmarshalText(text []byte) error {
	pn, ok := providerNameMapReverse[string(text)]
	if !ok {
		return ErrInvalidProviderName
	}

	*p = pn

	return nil
}

func (p ProviderName) String() string {
	return providerNameMap[p]
}
