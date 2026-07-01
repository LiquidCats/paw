package configuration

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/LiquidCats/paw/services/litehsm/pkg/unsafe"
	"github.com/awnumar/memguard"
)

var ErrSensitiveParamMissingType = errors.New("sensitive param must have type and source separated by colon")

const (
	typeFile = "file"
	typeEnvs = "envs"
)

func unmarshalSensitiveText(text []byte) (*memguard.LockedBuffer, error) {
	defer memguard.WipeBytes(text)

	if len(text) == 0 {
		return nil, nil //nolint:nilnil
	}

	typBytes, sourceBytes, ok := bytes.Cut(text, []byte(":"))
	if !ok || len(typBytes) == 0 {
		return nil, ErrSensitiveParamMissingType
	}

	if len(sourceBytes) == 0 {
		return nil, errors.New("sensitive param must have source separated by colon")
	}

	typ := unsafe.BytesToString(typBytes)
	source := unsafe.BytesToString(sourceBytes)

	switch typ {
	case typeEnvs:
		envValue, ok := os.LookupEnv(source)
		if !ok {
			return nil, errors.New("environment variable $" + source + " not set")
		}

		return memguard.NewBufferFromBytes(unsafe.StringToBytes(envValue)), nil
	case typeFile:
		file, err := os.Open(source)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer func() {
			_ = file.Close()
		}()

		buf, err := memguard.NewBufferFromEntireReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		return buf, nil
	default:
		return nil, errors.New("unsupported sensitive param type: " + typ)
	}
}

type LockedParam struct {
	*memguard.LockedBuffer
}

func (sp *LockedParam) UnmarshalText(text []byte) error {
	buf, err := unmarshalSensitiveText(text)
	if err != nil {
		return fmt.Errorf("sensitive param: %w", err)
	}

	sp.LockedBuffer = buf

	return nil
}

type SealedParam struct {
	*memguard.Enclave
}

func (sp *SealedParam) UnmarshalText(text []byte) error {
	buf, err := unmarshalSensitiveText(text)
	if err != nil {
		return fmt.Errorf("sensitive param: %w", err)
	}

	sp.Enclave = buf.Seal()

	return nil
}
