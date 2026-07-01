package hsm_test

import (
	"context"
	"testing"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/usecase/hsm"
	mocks "github.com/LiquidCats/paw/services/litehsm/test/mocks/litehsm"
	"github.com/google/uuid"
)

func BenchmarkKeyManagerSetStatus(b *testing.B) {
	b.Run("Handle", func(b *testing.B) {
		keyID := uuid.New()
		repo := mocks.NewMockKeyManagerRepository(b)
		usecase := hsm.NewKeyManagerSetStatus(repo)
		ctx := context.Background()

		b.ReportAllocs()
		for b.Loop() {
			if err := usecase.Handle(ctx, keyID, entities.KeyStatusEnabled); err != nil {
				b.Fatalf("Handle() error = %v", err)
			}
		}
	})

	b.Run("HandleValidation", func(b *testing.B) {
		repo := mocks.NewMockKeyManagerRepository(b)
		usecase := hsm.NewKeyManagerSetStatus(repo)
		ctx := context.Background()

		b.ReportAllocs()
		for b.Loop() {
			err := usecase.Handle(ctx, entities.KeyID{}, entities.KeyStatusEnabled)
			if err == nil {
				b.Fatal("Handle() error = nil, want error")
			}
		}
	})
}
