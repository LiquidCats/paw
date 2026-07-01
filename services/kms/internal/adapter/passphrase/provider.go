package passphrase

import (
	"errors"
	"fmt"
	"os"

	"github.com/LiquidCats/paw/services/litehsm/pkg/unsafe"
	"github.com/awnumar/memguard"
	"golang.org/x/term"
)

type EnvPassphraseProvider struct{}

func (*EnvPassphraseProvider) Name() string {
	return "env"
}

func (p *EnvPassphraseProvider) Get(dst string) (*memguard.LockedBuffer, error) {
	val, ok := os.LookupEnv(dst)
	if !ok {
		return nil, errors.New("environment variable not set")
	}

	return memguard.NewBufferFromBytes(unsafe.StringToBytes(val)), nil
}

type FilePassphraseProvider struct{}

func (*FilePassphraseProvider) Name() string {
	return "file"
}

func (p *FilePassphraseProvider) Get(dst string) (*memguard.LockedBuffer, error) {
	file, err := os.Open(dst)
	if err != nil {
		return nil, fmt.Errorf("struct=FilePassphraseProvider, method=Get, call=os.Open, dest=%q: %w", dst, err)
	}
	defer func() {
		_ = file.Close()
	}()

	val, err := memguard.NewBufferFromEntireReader(file)
	if err != nil {
		return nil, fmt.Errorf("struct=FilePassphraseProvider, method=Get, call=memguard.NewBufferFromEntireReader: %w", err)
	}

	return val, nil
}

type StdInPassphraseProvider struct{}

func (*StdInPassphraseProvider) Name() string {
	return "stdin"
}

func (p *StdInPassphraseProvider) Get(_ string) (*memguard.LockedBuffer, error) {
	_, err := fmt.Fprintln(os.Stdout, "enter passphrase:")
	if err != nil {
		return nil, fmt.Errorf("struct=StdInPassphraseProvider, method=Get, call=fmt.Fprintln: %w", err)
	}

	b, err := term.ReadPassword(int(os.Stdin.Fd())) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("struct=StdInPassphraseProvider, method=Get, call=term.ReadPassword: %w", err)
	}

	return memguard.NewBufferFromBytes(b), nil
}
