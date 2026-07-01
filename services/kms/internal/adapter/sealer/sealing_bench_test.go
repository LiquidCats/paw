package sealer_test

import (
	"testing"

	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/sealer"
)

func BenchmarkSealing_Seal(b *testing.B) {
	sealing := sealer.NewDefault(testMagic)
	passphrase := newPassphrase(b, testPassphrase)
	data := newPlaintext(b, testPlaintext)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		env, err := sealing.Seal(passphrase, data)
		if err != nil {
			b.Fatalf("Seal() error = %v", err)
		}
		env.Destroy()
	}
}

func BenchmarkSealing_Unseal(b *testing.B) {
	sealing := sealer.NewDefault(testMagic)
	passphrase := newPassphrase(b, testPassphrase)
	env, err := sealing.Seal(passphrase, newPlaintext(b, testPlaintext))
	if err != nil {
		b.Fatalf("Seal() error = %v", err)
	}
	b.Cleanup(env.Destroy)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		plain, err := sealing.Unseal(env, passphrase)
		if err != nil {
			b.Fatalf("Unseal() error = %v", err)
		}
		if plain == nil {
			b.Fatal("Unseal() enclave = nil, want plaintext enclave")
		}
	}
}

func BenchmarkSealing_SealUnsealRoundTrip(b *testing.B) {
	sealing := sealer.NewDefault(testMagic)
	passphrase := newPassphrase(b, testPassphrase)
	data := newPlaintext(b, testPlaintext)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		env, err := sealing.Seal(passphrase, data)
		if err != nil {
			b.Fatalf("Seal() error = %v", err)
		}

		plain, err := sealing.Unseal(env, passphrase)
		env.Destroy()
		if err != nil {
			b.Fatalf("Unseal() error = %v", err)
		}
		if plain == nil {
			b.Fatal("Unseal() enclave = nil, want plaintext enclave")
		}
	}
}
