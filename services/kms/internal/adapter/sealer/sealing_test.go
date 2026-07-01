package sealer_test

import (
	"errors"
	"testing"

	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/sealer"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	"github.com/awnumar/memguard"
)

const (
	testMagic      = "TST1"
	testPassphrase = "correct horse battery staple"
	testPlaintext  = "seed material to encrypt"
)

func newPassphrase(t testing.TB, value string) *memguard.Enclave {
	t.Helper()

	passphrase := memguard.NewEnclave([]byte(value))
	if passphrase == nil {
		t.Fatal("NewEnclave() = nil, want passphrase enclave")
	}

	return passphrase
}

func newPlaintext(t testing.TB, value string) *memguard.LockedBuffer { //nolint:unparam
	t.Helper()

	data := memguard.NewBufferFromBytes([]byte(value))
	if data == nil || data.Size() == 0 {
		t.Fatal("NewBufferFromBytes() = nil or empty, want plaintext buffer")
	}
	t.Cleanup(data.Destroy)

	return data
}

func openEnclave(t testing.TB, enclave *memguard.Enclave) *memguard.LockedBuffer {
	t.Helper()

	if enclave == nil {
		t.Fatal("enclave = nil, want sealed plaintext")
	}

	buf, err := enclave.Open()
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	t.Cleanup(buf.Destroy)

	return buf
}

func sealEnvelope(t *testing.T) (*sealer.Sealing, *entities.Envelope) {
	t.Helper()

	sealing := sealer.NewDefault(testMagic)
	env, err := sealing.Seal(newPassphrase(t, testPassphrase), newPlaintext(t, testPlaintext))
	if err != nil {
		t.Fatalf("Seal() error = %v, want nil", err)
	}
	t.Cleanup(env.Destroy)

	return sealing, env
}

func flipFirstByte(buf *memguard.LockedBuffer) {
	buf.Melt()
	buf.CopyAt(0, []byte{buf.Bytes()[0] ^ 0xFF})
	buf.Freeze()
}

func TestNewDefault_SealEnvelopeMetadata(t *testing.T) {
	_, env := sealEnvelope(t)

	if got, want := string(env.Magic[:]), testMagic; got != want {
		t.Fatalf("Magic = %q, want %q", got, want)
	}
	if got, want := env.Version, entities.DefaultVersion; got != want {
		t.Fatalf("Version = %d, want %d", got, want)
	}
	if got, want := env.KDF, entities.KDFArgon2id; got != want {
		t.Fatalf("KDF = %d, want %d", got, want)
	}
	if got, want := env.AEAD, entities.AEADxchacha20poly1305; got != want {
		t.Fatalf("AEAD = %d, want %d", got, want)
	}
	if got, want := env.KDFParams, entities.DefaultKDFParams; got != want {
		t.Fatalf("KDFParams = %+v, want %+v", got, want)
	}
}

func TestSealing_SealPopulatesEnvelopeBuffers(t *testing.T) {
	_, env := sealEnvelope(t)

	tests := []struct {
		name string
		buf  *memguard.LockedBuffer
		size int
	}{
		{"salt", env.Salt, entities.SaltSize},
		{"nonce kek", env.NonceKEK, entities.NonceSize},
		{"nonce dek", env.NonceDEK, entities.NonceSize},
		{"wrapped dek", env.WrappedDEK, entities.WrappedDEKSize},
		{"ciphertext", env.Ciphertext, len(testPlaintext) + entities.TagSize},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.buf == nil {
				t.Fatal("buffer = nil, want locked buffer")
			}
			if got := tc.buf.Size(); got != tc.size {
				t.Fatalf("Size() = %d, want %d", got, tc.size)
			}
		})
	}
}

func TestSealing_SealUnsealRoundTrip(t *testing.T) {
	sealing, env := sealEnvelope(t)

	got, err := sealing.Unseal(env, newPassphrase(t, testPassphrase))
	if err != nil {
		t.Fatalf("Unseal() error = %v, want nil", err)
	}

	plain := openEnclave(t, got)
	if !plain.EqualTo([]byte(testPlaintext)) {
		t.Fatalf("Unseal() plaintext = %q, want %q", plain.Bytes(), testPlaintext)
	}
}

func TestSealing_SealRejectsInvalidInput(t *testing.T) {
	data := newPlaintext(t, testPlaintext)
	passphrase := newPassphrase(t, testPassphrase)

	tests := []struct {
		name       string
		passphrase *memguard.Enclave
		data       *memguard.LockedBuffer
		wantErr    error
	}{
		{"nil passphrase", nil, data, sealer.ErrEmptyPassphrase},
		{"nil data", passphrase, nil, sealer.ErrEmptyData},
		{"empty data", passphrase, memguard.NewBuffer(0), sealer.ErrEmptyData},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.data != nil && tc.data.Size() == 0 {
				t.Cleanup(tc.data.Destroy)
			}

			env, err := sealer.NewDefault(testMagic).Seal(tc.passphrase, tc.data)
			if env != nil {
				t.Cleanup(env.Destroy)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Seal() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestSealing_SealUsesFreshRandomMaterial(t *testing.T) {
	sealing := sealer.NewDefault(testMagic)
	first, err := sealing.Seal(newPassphrase(t, testPassphrase), newPlaintext(t, testPlaintext))
	if err != nil {
		t.Fatalf("first Seal() error = %v, want nil", err)
	}
	t.Cleanup(first.Destroy)

	second, err := sealing.Seal(newPassphrase(t, testPassphrase), newPlaintext(t, testPlaintext))
	if err != nil {
		t.Fatalf("second Seal() error = %v, want nil", err)
	}
	t.Cleanup(second.Destroy)

	if first.Salt.EqualTo(second.Salt.Bytes()) {
		t.Fatal("Seal() reused salt across calls")
	}
	if first.NonceKEK.EqualTo(second.NonceKEK.Bytes()) {
		t.Fatal("Seal() reused KEK nonce across calls")
	}
	if first.NonceDEK.EqualTo(second.NonceDEK.Bytes()) {
		t.Fatal("Seal() reused DEK nonce across calls")
	}
	if first.WrappedDEK.EqualTo(second.WrappedDEK.Bytes()) {
		t.Fatal("Seal() reused wrapped DEK across calls")
	}
	if first.Ciphertext.EqualTo(second.Ciphertext.Bytes()) {
		t.Fatal("Seal() reused ciphertext across calls")
	}
}

func TestSealing_UnsealRejectsWrongPassphrase(t *testing.T) {
	sealing, env := sealEnvelope(t)

	got, err := sealing.Unseal(env, newPassphrase(t, "wrong passphrase"))
	if got != nil {
		t.Fatalf("Unseal() enclave = %v, want nil", got)
	}
	if !errors.Is(err, sealer.ErrFailedUnseal) {
		t.Fatalf("Unseal() error = %v, want %v", err, sealer.ErrFailedUnseal)
	}
}

func TestSealing_UnsealRejectsTamperedEnvelope(t *testing.T) {
	tests := []struct {
		name   string
		tamper func(*entities.Envelope)
	}{
		{
			name: "magic",
			tamper: func(env *entities.Envelope) {
				env.Magic[0] ^= 0xFF
			},
		},
		{
			name: "version",
			tamper: func(env *entities.Envelope) {
				env.Version++
			},
		},
		{
			name: "salt",
			tamper: func(env *entities.Envelope) {
				flipFirstByte(env.Salt)
			},
		},
		{
			name: "kek nonce",
			tamper: func(env *entities.Envelope) {
				flipFirstByte(env.NonceKEK)
			},
		},
		{
			name: "wrapped dek",
			tamper: func(env *entities.Envelope) {
				flipFirstByte(env.WrappedDEK)
			},
		},
		{
			name: "dek nonce",
			tamper: func(env *entities.Envelope) {
				flipFirstByte(env.NonceDEK)
			},
		},
		{
			name: "ciphertext",
			tamper: func(env *entities.Envelope) {
				flipFirstByte(env.Ciphertext)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sealing, env := sealEnvelope(t)
			tc.tamper(env)

			got, err := sealing.Unseal(env, newPassphrase(t, testPassphrase))
			if got != nil {
				t.Fatalf("Unseal() enclave = %v, want nil", got)
			}
			if !errors.Is(err, sealer.ErrFailedUnseal) {
				t.Fatalf("Unseal() error = %v, want %v", err, sealer.ErrFailedUnseal)
			}
		})
	}
}

func TestSealing_UnsealRejectsInvalidKDFParams(t *testing.T) {
	sealing, env := sealEnvelope(t)
	env.KDFParams.Parallelism = 0

	got, err := sealing.Unseal(env, newPassphrase(t, testPassphrase))
	if got != nil {
		t.Fatalf("Unseal() enclave = %v, want nil", got)
	}
	if err == nil {
		t.Fatal("Unseal() error = nil, want invalid kdf params error")
	}
}
