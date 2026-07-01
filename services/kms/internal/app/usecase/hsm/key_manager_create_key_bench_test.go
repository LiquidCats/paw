package hsm_test

import (
	"context"
	"testing"
	"time"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/usecase/hsm"
	mocks "github.com/LiquidCats/paw/services/litehsm/test/mocks/litehsm"
)

func BenchmarkKeyManagerCreateKey(b *testing.B) {
	b.Run("Handle", func(b *testing.B) {
		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		repo := mocks.NewMockKeyManagerRepository(b)
		usecase := hsm.NewKeyManagerCreateKey(repo)
		entry := validKeyEntry(&expiresAt)
		ctx := context.Background()

		b.ReportAllocs()
		for b.Loop() {
			_, err := usecase.Handle(ctx, entry)
			if err != nil {
				b.Fatalf("Handle() error = %v", err)
			}
		}
	})

	b.Run("HandleValidation", func(b *testing.B) {
		repo := mocks.NewMockKeyManagerRepository(b)
		usecase := hsm.NewKeyManagerCreateKey(repo)
		entry := validKeyEntry(nil)
		entry.Alias = ""
		ctx := context.Background()

		b.ReportAllocs()
		for b.Loop() {
			_, err := usecase.Handle(ctx, entry)
			if err == nil {
				b.Fatal("Handle() error = nil, want error")
			}
		}
	})
}
