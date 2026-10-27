package coinmarketcap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coinmarketcap/dto"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/config"
	entity "github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/errors"
)

type Repository struct {
	cfg config.CoinMarketCapConfig
}

func NewRepository(cfg config.CoinMarketCapConfig) *Repository {
	return &Repository{
		cfg: cfg,
	}
}

func (r *Repository) GetRate(ctx context.Context, pair entity.Pair) (decimal.Decimal, error) {
	fullURL := fmt.Sprintf(
		"%s?amount=1&symbol=%s&convert=%s",
		r.cfg.URL,
		pair.From.ToUpper(),
		pair.To.ToUpper(),
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

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-CMC_PRO_API_KEY", string(r.cfg.Secret)) //nolint:canonicalheader

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

	quotePrice, ok := response.Data.Quote[pair.To.ToUpper().String()]
	if !ok {
		return decimal.Zero, eris.Wrap(err, "repo: could not parse response")
	}

	return decimal.NewFromFloat(quotePrice.Price), nil
}
