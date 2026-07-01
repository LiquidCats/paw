//go:build !ziren

package keccak

import (
	"hash"
	"sync"

	"golang.org/x/crypto/sha3"
)

// NewKeccakState creates a new KeccakState.
func NewKeccakState() State {
	return sha3.NewLegacyKeccak256().(State) // nolint:errcheck
}

var hasherPool = sync.Pool{
	New: func() any {
		return sha3.NewLegacyKeccak256().(State) // nolint:errcheck
	},
}

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	b := make([]byte, 32)
	d := hasherPool.Get().(State) // nolint:errcheck
	d.Reset()
	for _, b := range data {
		d.Write(b)
	}
	_, _ = d.Read(b)
	hasherPool.Put(d)
	return b
}

// State wraps sha3.state. In addition to the usual hash methods, it also supports
// Read to get a variable amount of data from the hash state. Read is faster than Sum
// because it doesn't copy the internal state, but also modifies the internal state.
type State interface {
	hash.Hash
	Read([]byte) (int, error)
}
