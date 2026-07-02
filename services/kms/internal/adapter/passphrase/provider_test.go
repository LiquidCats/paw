package passphrase_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/passphrase"
	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/require"
)

func requireLockedBufferEqual(t *testing.T, got *memguard.LockedBuffer, want []byte) {
	t.Helper()

	if got == nil {
		t.Fatal("Get() buffer = nil, want locked buffer")
	}
	if !got.EqualTo(want) {
		t.Fatalf("Get() buffer = %q, want %q", got.Bytes(), want)
	}
}

func TestEnvPassphraseProvider_Name(t *testing.T) {
	provider := new(passphrase.EnvPassphraseProvider)

	if got, want := provider.Name(), "env"; got != want {
		t.Fatalf("Name() = %q, want %q", got, want)
	}
}

func TestEnvPassphraseProvider_Get(t *testing.T) {
	t.Run("reads configured environment variable", func(t *testing.T) {
		const name = "LITEHSM_TEST_ENV_PASSPHRASE"
		want := []byte("correct horse battery staple")
		t.Setenv(name, string(want))

		provider := new(passphrase.EnvPassphraseProvider)
		got, err := provider.Get(name)
		if err != nil {
			t.Fatalf("Get() error = %v, want nil", err)
		}
		defer got.Destroy()

		requireLockedBufferEqual(t, got, want)
	})

	t.Run("returns empty buffer for empty environment variable", func(t *testing.T) {
		const name = "LITEHSM_TEST_EMPTY_ENV_PASSPHRASE"
		t.Setenv(name, "")

		provider := new(passphrase.EnvPassphraseProvider)
		got, err := provider.Get(name)
		if err != nil {
			t.Fatalf("Get() error = %v, want nil", err)
		}
		defer got.Destroy()

		requireLockedBufferEqual(t, got, []byte{})
	})

	t.Run("rejects missing environment variable", func(t *testing.T) {
		const name = "LITEHSM_TEST_MISSING_ENV_PASSPHRASE"
		if err := os.Unsetenv(name); err != nil {
			t.Fatalf("Unsetenv() error = %v", err)
		}

		provider := new(passphrase.EnvPassphraseProvider)
		got, err := provider.Get(name)
		if err == nil {
			t.Fatal("Get() error = nil, want error")
		}
		if got != nil {
			t.Fatalf("Get() buffer = %v, want nil", got)
		}
		if got, want := err.Error(), "environment variable not set"; got != want {
			t.Fatalf("Get() error = %q, want %q", got, want)
		}
	})
}

func TestFilePassphraseProvider_Name(t *testing.T) {
	provider := new(passphrase.FilePassphraseProvider)

	if got, want := provider.Name(), "file"; got != want {
		t.Fatalf("Name() = %q, want %q", got, want)
	}
}

func TestFilePassphraseProvider_Get(t *testing.T) {
	t.Run("reads entire file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "passphrase.txt")
		want := []byte("secret\nwith spaces")
		if err := os.WriteFile(path, want, 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		provider := new(passphrase.FilePassphraseProvider)
		got, err := provider.Get(path)
		if err != nil {
			t.Fatalf("Get() error = %v, want nil", err)
		}
		defer got.Destroy()

		requireLockedBufferEqual(t, got, want)
	})

	t.Run("reads empty file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "empty-passphrase.txt")
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		provider := new(passphrase.FilePassphraseProvider)
		got, err := provider.Get(path)
		if err != nil {
			t.Fatalf("Get() error = %v, want nil", err)
		}
		defer got.Destroy()

		requireLockedBufferEqual(t, got, []byte{})
	})

	t.Run("wraps open error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing-passphrase.txt")

		provider := new(passphrase.FilePassphraseProvider)
		got, err := provider.Get(path)
		if err == nil {
			t.Fatal("Get() error = nil, want error")
		}
		if got != nil {
			t.Fatalf("Get() buffer = %v, want nil", got)
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Get() error = %v, want wrapped os.ErrNotExist", err)
		}
	})
}

func TestStdInPassphraseProvider_Name(t *testing.T) {
	provider := new(passphrase.StdInPassphraseProvider)

	if got, want := provider.Name(), "stdin"; got != want {
		t.Fatalf("Name() = %q, want %q", got, want)
	}
}

func TestStdInPassphraseProvider_Get_NonTerminalStdin(t *testing.T) {
	originalStdin := os.Stdin
	originalStdout := os.Stdout
	t.Cleanup(func() {
		os.Stdin = originalStdin   //nolint:reassign
		os.Stdout = originalStdout //nolint:reassign
	})

	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("Open(os.DevNull) for stdin error = %v", err)
	}
	defer func() {
		_ = stdin.Close()
	}()

	stdoutReader, stdoutWriter, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		_ = stdoutReader.Close()
	}()

	os.Stdin = stdin         //nolint:reassign
	os.Stdout = stdoutWriter //nolint:reassign

	provider := new(passphrase.StdInPassphraseProvider)
	got, err := provider.Get("")

	if closeErr := stdoutWriter.Close(); closeErr != nil {
		t.Fatalf("stdout pipe close error = %v", closeErr)
	}
	output, readErr := io.ReadAll(stdoutReader)
	if readErr != nil {
		t.Fatalf("ReadAll(stdout) error = %v", readErr)
	}

	if err == nil {
		t.Fatal("Get() error = nil, want error")
	}
	if got != nil {
		t.Fatalf("Get() buffer = %v, want nil", got)
	}
	if got, want := string(output), "enter passphrase:\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}
