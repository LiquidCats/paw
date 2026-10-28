package cex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/cex/dto"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/config"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/entities"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/errors"
)

type Repository struct {
	cfg config.ProviderConfig
}

func NewRepository(cfg config.ProviderConfig) *Repository {
	return &Repository{cfg: cfg}
}

func (r *Repository) GetRate(ctx context.Context, pair entities.Pair) (decimal.Decimal, error) {
	parsedURL, err := url.ParseRequestURI(
		fmt.Sprintf(
			"%s/%s/%s",
			r.cfg.URL,
			pair.From,
			pair.To,
		),
	)
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

	var resp dto.APIResponse

	if err = decoder.Decode(&resp); err != nil {
		return decimal.Zero, eris.Wrap(err, "repo: could not unmarshal response")
	}

	value, err := decimal.NewFromString(resp.Last)
	if err != nil {
		return decimal.Zero, eris.Wrap(err, "repo: could not parse value")
	}

	return value, nil
}
