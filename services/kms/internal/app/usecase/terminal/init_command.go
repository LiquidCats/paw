package terminal

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/ports"
	"github.com/awnumar/memguard"
)

type InitCommand struct {
	sealer    ports.Sealer
	providers []ports.PassphraseProvider
}

func NewInitCommand(sealer ports.Sealer, providers ...ports.PassphraseProvider) (*InitCommand, error) {
	if sealer == nil {
		return nil, errors.New("sealer is required")
	}
	if len(providers) == 0 {
		return nil, errors.New("at least one passphrase provider is required")
	}
	for i, provider := range providers {
		if provider == nil {
			return nil, fmt.Errorf("passphrase provider %d is nil", i)
		}
	}

	return &InitCommand{
		sealer:    sealer,
		providers: providers,
	}, nil
}

func (uc *InitCommand) Run() error {
	return uc.RunArgs(os.Args[2:])
}

func (uc *InitCommand) RunArgs(args []string) error {
	initArgsSet := flag.NewFlagSet(uc.Name(), flag.ContinueOnError)
	initArgsSet.SetOutput(io.Discard)

	typeNames := make([]string, len(uc.providers))
	for i, provider := range uc.providers {
		typeNames[i] = provider.Name()
	}
	sourceTypeDesc := "type of passphrase source: " + strings.Join(typeNames, "|")

	sourceTypeFlag := initArgsSet.String("from", uc.providers[0].Name(), sourceTypeDesc)
	sourceDestFlag := initArgsSet.String("input", "", "path to passphrase file or env variable name")
	outputFlag := initArgsSet.String("output", "", "path where envelope will be stored")

	if err := initArgsSet.Parse(args); err != nil {
		return fmt.Errorf("init args: %w", err)
	}

	sourceType := *sourceTypeFlag
	sourceDest := *sourceDestFlag
	output := *outputFlag
	if output == "" {
		return errors.New("output path is required")
	}

	passphrase, err := uc.extractPassphrase(sourceType, sourceDest)
	if err != nil {
		return fmt.Errorf("could not extract passphrase: %w", err)
	}
	if passphrase.Size() == 0 {
		return errors.New("could not extract passphrase: provider returned empty passphrase")
	}

	seed, err := memguard.NewBufferFromReader(rand.Reader, 32)
	if err != nil {
		return fmt.Errorf("init seed: %w", err)
	}
	defer seed.Destroy()

	env, err := uc.sealer.Seal(passphrase, seed)
	if err != nil {
		return fmt.Errorf("init seal: %w", err)
	}
	if env == nil {
		return errors.New("sealer returned nil envelope")
	}
	defer env.Destroy()

	data, err := env.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal enclave: %w", err)
	}

	dir := filepath.Dir(output)
	tmp, err := os.CreateTemp(dir, ".seal-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	tmpName := tmp.Name()
	defer func(name string) {
		_ = os.Remove(name)
		_ = tmp.Close()
	}(tmpName)

	if err := tmp.Chmod(0o600); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if n, err := tmp.Write(data); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	} else if n != len(data) {
		return fmt.Errorf("write temp file: %w", io.ErrShortWrite)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpName, output); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

func (uc *InitCommand) Name() string {
	return "init"
}

func (uc *InitCommand) extractPassphrase(source, dest string) (*memguard.Enclave, error) {
	for _, provider := range uc.providers {
		if source == provider.Name() {
			p, err := provider.Get(dest)
			if err != nil {
				return nil, fmt.Errorf("get passphrase %s: %w", dest, err)
			}
			if p == nil {
				return nil, fmt.Errorf("passphrase not found for %s", dest)
			}

			return p.Seal(), nil
		}
	}

	return nil, fmt.Errorf("unsupported source type %q", source)
}
