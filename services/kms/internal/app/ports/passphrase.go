package ports

import "github.com/awnumar/memguard"

type PassphraseProvider interface {
	Name() string
	Get(dts string) (*memguard.LockedBuffer, error)
}
