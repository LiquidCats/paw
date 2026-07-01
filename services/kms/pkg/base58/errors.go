package base58

import (
	"github.com/rotisserie/eris"
)

var (
	// ErrChecksum indicates that the checksum of a check-encoded string does not verify against
	// the checksum.
	ErrChecksum = eris.New("checksum error")
	// ErrInvalidFormat indicates that the check-encoded string has an invalid format.
	ErrInvalidFormat = eris.New("invalid format: version and/or checksum bytes missing")
)
