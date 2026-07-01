package hsm_test

import (
	"context"
	"testing"
	"time"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/usecase/hsm"
	mocks "github.com/LiquidCats/paw/services/litehsm/test/mocks/litehsm"
	"github.com/google/uuid"
)

func BenchmarkKeyManagerSetExpiration(b *testing.B) {
	b.Run("Handle", func(b *testing.B) {
		keyID := uuid.New()
		expiration := time.Now().Add(24 * time.Hour).UTC()
		repo := mocks.NewMockKeyManagerRepository(b)
		usecase := hsm.NewKeyManagerSetExpiration(repo)
		ctx := context.Background()

		b.ReportAllocs()
		for b.Loop() {
			if err := usecase.Handle(ctx, keyID, expiration); err != nil {
				b.Fatalf("Handle() error = %v", err)
			}
		}
	})

	b.Run("HandleValidation", func(b *testing.B) {
		expiration := time.Now().Add(24 * time.Hour).UTC()
		repo := mocks.NewMockKeyManagerRepository(b)
		usecase := hsm.NewKeyManagerSetExpiration(repo)
		ctx := context.Background()

		b.ReportAllocs()
		for b.Loop() {
			err := usecase.Handle(ctx, entities.KeyID{}, expiration)
			if err == nil {
				b.Fatal("Handle() error = nil, want error")
			}
		}
	})
}
