package api_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/api"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/config"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

func TestFactory(t *testing.T) {
	t.Parallel()

	supported := []values.ProviderName{
		values.ProviderNameCex,
		values.ProviderNameCoinApi,
		values.ProviderNameCoinGate,
		values.ProviderNameCoinGecko,
		values.ProviderNameCoinMarketCap,
	}

	for _, name := range supported {
		t.Run(name.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := api.Factory(name, config.ProviderConfig{})

			require.NoError(t, err)
			assert.NotNil(t, repo)
		})
	}

	t.Run("unsupported", func(t *testing.T) {
		t.Parallel()

		repo, err := api.Factory("nope", config.ProviderConfig{})

		assert.Nil(t, repo)
		assert.ErrorIs(t, err, values.ErrUnsupportedProvider)
	})
}
