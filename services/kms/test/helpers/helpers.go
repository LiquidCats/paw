package helpers

import (
	"errors"
	"testing"

	domainErr "github.com/LiquidCats/paw/services/litehsm/internal/app/domain/errors"
)

func RequireValidationError(t *testing.T, err error, wantMsg string) {
	t.Helper()

	if err == nil {
		t.Fatal("Handle() error = nil, want validation error")
	}
	var validationErr *domainErr.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Handle() error = %T, want *ValidationError", err)
	}
	if validationErr.Error() != wantMsg {
		t.Fatalf("Handle() validation error = %q, want %q", validationErr.Error(), wantMsg)
	}
}
