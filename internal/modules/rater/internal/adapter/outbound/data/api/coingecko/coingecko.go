package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coingecko/dto"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/config"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/errors"
)

type Repository struct {
	cfg config.CoinGeckoConfig
}

func NewRepository(cfg config.CoinGeckoConfig) *Repository {
	return &Repository{
		cfg: cfg,
	}
}

func (r *Repository) GetRate(ctx context.Context, pair entities.Pair) (decimal.Decimal, error) {
	// https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd&precision=8
	id, ok := dto.GetCoinID(pair.From)
	if !ok {
		return decimal.Zero, eris.New("repo: cant find coingecko id")
	}

	fullURL := fmt.Sprintf(
		"%s?ids=%s&vs_currencies=%s&precision=8",
		r.cfg.URL,
		id,
		pair.To.ToLower(),
	)

	parsedURL, err := url.ParseRequestURI(fullURL)
	if err != nil {
		return decimal.Zero, eris.Wrap(err, "repo: incorrect request url")
	}

	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return decimal.Zero, eris.New("repo: unsupported URL scheme")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return decimal.Zero, eris.Wrap(err, "repo: could not create request")
	}

	res, err := http.DefaultClient.Do(req) //nolint:gosec
	if err != nil {
		return decimal.Zero, eris.Wrap(err, "repo: error making http request")
	}
	defer func() {
		_ = res.Body.Close()
	}()

	decoder := json.NewDecoder(res.Body)
	if res.StatusCode >= http.StatusBadRequest {
		var resBody string
		if err = decoder.Decode(&resBody); err != nil && eris.Is(err, io.EOF) {
			return decimal.Zero, eris.Wrap(err, "repo: could not decode response body")
		}

		return decimal.Zero, &errors.ProviderRequestFailedError{
			StatusCode: res.StatusCode,
			Body:       resBody,
		}
	}

	var response dto.APIResponse

	if err = decoder.Decode(&response); err != nil {
		return decimal.Zero, eris.Wrap(err, "repo: could not parse response")
	}

	rateBase, ok := response[id.String()]
	if !ok {
		return decimal.Zero, eris.New("repo: could not get base rate from response")
	}

	rateQuote, ok := rateBase.(map[string]interface{})
	if !ok {
		return decimal.Zero, eris.New("repo: could not get quoted rate from response")
	}

	val, ok := rateQuote[pair.To.ToLower().String()]
	if !ok {
		return decimal.Zero, eris.New("repo: could not get rate value from response")
	}

	floatVal, ok := val.(float64)
	if !ok {
		return decimal.Zero, eris.New("repo: could not get float value from response")
	}

	return decimal.NewFromFloat(floatVal), nil
}
