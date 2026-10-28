package errors_test

import (
	"testing"

	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/assert"

	domainerrors "github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/errors"
)

func TestProviderRequestFailedError(t *testing.T) {
	t.Parallel()

	err := &domainerrors.ProviderRequestFailedError{StatusCode: 503, Body: "down"}

	assert.EqualError(t, err, "provider request failed with code 503")

	var target *domainerrors.ProviderRequestFailedError
	require := assert.New(t)
	require.True(eris.As(eris.Wrap(err, "wrapped"), &target))
	require.Equal("down", target.Body)
}
