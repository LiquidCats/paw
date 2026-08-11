package base58

import (
	"errors"
)

var (
	// ErrChecksum indicates that the checksum of a check-encoded string does not verify against
	// the checksum.
	ErrChecksum = errors.New("checksum error")
	// ErrInvalidFormat indicates that the check-encoded string has an invalid format.
	ErrInvalidFormat = errors.New("invalid format: version and/or checksum bytes missing")
)
