package passphrase_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/passphrase"
	"github.com/stretchr/testify/require"
)

func BenchmarkEnvPassphraseProvider_Get(b *testing.B) {
	const name = "LITEHSM_BENCH_ENV_PASSPHRASE"
	b.Setenv(name, "correct horse battery staple")
	provider := new(passphrase.EnvPassphraseProvider)

	b.ReportAllocs()
	for b.Loop() {
		buf, err := provider.Get(name)
		if err != nil {
			b.Fatalf("Get() error = %v", err)
		}
		buf.Destroy()
	}
}

func BenchmarkEnvPassphraseProvider_Get_Missing(b *testing.B) {
	const name = "LITEHSM_BENCH_MISSING_ENV_PASSPHRASE"
	if err := os.Unsetenv(name); err != nil {
		b.Fatalf("Unsetenv() error = %v", err)
	}
	provider := new(passphrase.EnvPassphraseProvider)

	b.ReportAllocs()
	for b.Loop() {
		buf, err := provider.Get(name)
		if err == nil {
			if buf != nil {
				buf.Destroy()
			}
			b.Fatal("Get() error = nil, want error")
		}
	}
}

func BenchmarkFilePassphraseProvider_Get(b *testing.B) {
	path := filepath.Join(b.TempDir(), "passphrase.txt")
	if err := os.WriteFile(path, []byte("correct horse battery staple"), 0o600); err != nil {
		b.Fatalf("WriteFile() error = %v", err)
	}
	provider := new(passphrase.FilePassphraseProvider)

	b.ReportAllocs()
	for b.Loop() {
		buf, err := provider.Get(path)
		if err != nil {
			b.Fatalf("Get() error = %v", err)
		}
		buf.Destroy()
	}
}

func BenchmarkFilePassphraseProvider_Get_Missing(b *testing.B) {
	path := filepath.Join(b.TempDir(), "missing-passphrase.txt")
	provider := new(passphrase.FilePassphraseProvider)

	b.ReportAllocs()
	for b.Loop() {
		buf, err := provider.Get(path)
		if err == nil {
			if buf != nil {
				buf.Destroy()
			}
			b.Fatal("Get() error = nil, want error")
		}
	}
}

func BenchmarkStdInPassphraseProvider_Get_NonTerminalStdin(b *testing.B) {
	originalStdin := os.Stdin
	originalStdout := os.Stdout
	b.Cleanup(func() {
		os.Stdin = originalStdin   //nolint:reassign
		os.Stdout = originalStdout //nolint:reassign
	})

	stdin, err := os.Open(os.DevNull)
	if err != nil {
		b.Fatalf("Open(os.DevNull) for stdin error = %v", err)
	}
	defer stdin.Close()

	stdout, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	require.NoError(b, err)

	defer stdout.Close()

	os.Stdin = stdin   //nolint:reassign
	os.Stdout = stdout //nolint:reassign

	provider := new(passphrase.StdInPassphraseProvider)

	b.ReportAllocs()
	for b.Loop() {
		buf, err := provider.Get("")
		if err == nil {
			if buf != nil {
				buf.Destroy()
			}
			b.Fatal("Get() error = nil, want error")
		}
	}
}
