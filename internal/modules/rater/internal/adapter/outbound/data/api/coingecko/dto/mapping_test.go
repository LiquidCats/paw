package dto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api/coingecko/dto"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

func TestGetCoinID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		iso    values.CurrencyISO
		want   string
		wantOK bool
	}{
		{name: "btc", iso: "BTC", want: "bitcoin", wantOK: true},
		{name: "case insensitive", iso: "eth", want: "ethereum", wantOK: true},
		{name: "unknown", iso: "XYZ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			id, ok := dto.GetCoinID(tt.iso)

			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.want, id.String())
		})
	}
}
