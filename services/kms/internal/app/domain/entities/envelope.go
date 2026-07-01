package entities

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/awnumar/memguard"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	SaltSize       = 32
	NonceSize      = chacha20poly1305.NonceSizeX
	DEKSize        = chacha20poly1305.KeySize
	TagSize        = 16
	WrappedDEKSize = DEKSize + TagSize

	// HeaderSize
	// Offset  Size  Field
	// ------  ----  -----
	// 0       4     magic "HSM1"
	// 4       1     version (currently 1)
	// 5       1     kdf_id (1 = argon2id)
	// 6       1     aead_id (1 = XChaCha20-Poly1305)
	// 7       4     argon2 memory (KiB, big-endian uint32)
	// 11      4     argon2 iterations (big-endian uint32)
	// 15      1     argon2 parallelism
	// 16      32    kdf salt
	// 48      24    nonce_kek (XChaCha20-Poly1305)
	// 72      48    wrapped DEK (32 bytes key + 16 bytes tag)
	// 120     24    nonce_dek
	// 144     N+16  sealed payload (seed + tag).
	HeaderSize = 4 + 1 + 1 + 1 + 4 + 4 + 1 + SaltSize + NonceSize + WrappedDEKSize + NonceSize
)

type (
	Version byte
	KDF     byte
	AEAD    byte
)

const (
	KDFUnknown KDF = iota
	KDFArgon2id
)

const (
	AEADUnknown AEAD = iota
	AEADxchacha20poly1305
)

const (
	DefaultVersion Version = 1
)

type KDFParams struct {
	Iterations  uint32
	MemoryKiB   uint32
	Parallelism uint8
}

const (
	ParamsMinParallelism = 1
	ParamsMaxParallelism = 16
	ParamsMinMemoryKiB   = 64
	ParamsMaxMemoryKiB   = 4 * 1024 * 1024 // 4GiB
	ParamsMinIterations  = 1
	ParamsMaxIterations  = 100
)

func (p *KDFParams) Validate() error {
	if p.Parallelism < ParamsMinParallelism {
		return fmt.Errorf("parallelism must be >= %d", ParamsMinParallelism)
	}
	if p.Parallelism > ParamsMaxParallelism {
		return fmt.Errorf("parallelism must be <= %d", ParamsMaxParallelism)
	}
	if p.MemoryKiB < ParamsMinMemoryKiB {
		return fmt.Errorf("memory must be >= %d KiB", ParamsMinMemoryKiB)
	}
	if p.MemoryKiB > ParamsMaxMemoryKiB {
		return fmt.Errorf("memory must be <= %d KiB", ParamsMaxMemoryKiB)
	}
	if p.Iterations < ParamsMinIterations {
		return fmt.Errorf("iterations must be >= %d", ParamsMinIterations)
	}
	if p.Iterations > ParamsMaxIterations {
		return fmt.Errorf("iterations must be <= %d", ParamsMaxIterations)
	}

	return nil
}

// DefaultKDFParams is a reasonable starting point. Tune for your hardware:
// run a benchmark on your weakest target and pick params that take ~1 second.
var DefaultKDFParams = KDFParams{
	MemoryKiB:   256 * 1024, // 256 MiB
	Iterations:  3,
	Parallelism: 4,
}

var (
	_ encoding.BinaryMarshaler   = (*Envelope)(nil)
	_ encoding.BinaryUnmarshaler = (*Envelope)(nil)
)

type Envelope struct {
	Magic   [4]byte
	Version Version
	KDF     KDF
	AEAD    AEAD

	KDFParams KDFParams

	Salt       *memguard.LockedBuffer
	NonceKEK   *memguard.LockedBuffer
	NonceDEK   *memguard.LockedBuffer
	WrappedDEK *memguard.LockedBuffer
	Ciphertext *memguard.LockedBuffer
}

func EnvelopeFromBuffer(buf *memguard.LockedBuffer) (*Envelope, error) {
	env := new(Envelope)

	if err := env.UnmarshalBinary(buf.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to unmarshal enclave: %w", err)
	}

	return env, nil
}

// AAD returns the static envelope metadata authenticated by both AEAD
// operations. KDFParams and salt are bound through the derived KEK, and nonces are
// bound by the AEAD calls themselves.
func (e *Envelope) AAD() []byte {
	buf := make([]byte, 0, 7)
	buf = append(buf, e.Magic[:]...)
	buf = append(buf, byte(e.Version), byte(e.KDF), byte(e.AEAD))
	return buf
}

// MarshalBinary serializes to the on-disk format.
func (e *Envelope) MarshalBinary() ([]byte, error) {
	buf := make([]byte, HeaderSize+e.Ciphertext.Size())
	o := 0
	copy(buf[o:], e.AAD())
	o += 7
	binary.BigEndian.PutUint32(buf[o:], e.KDFParams.MemoryKiB)
	o += 4
	binary.BigEndian.PutUint32(buf[o:], e.KDFParams.Iterations)
	o += 4
	buf[o] = e.KDFParams.Parallelism
	o++
	copy(buf[o:], e.Salt.Bytes())
	o += SaltSize
	copy(buf[o:], e.NonceKEK.Bytes())
	o += NonceSize
	copy(buf[o:], e.WrappedDEK.Bytes())
	o += WrappedDEKSize
	copy(buf[o:], e.NonceDEK.Bytes())
	o += NonceSize
	copy(buf[o:], e.Ciphertext.Bytes())
	return buf, nil
}

// UnmarshalBinary parses the on-disk format with strict validation.
func (e *Envelope) UnmarshalBinary(data []byte) error {
	if len(data) < HeaderSize+TagSize {
		return errors.New("envelope: truncated")
	}

	if !bytes.Equal(data[0:4], e.Magic[:]) {
		return errors.New("envelope: bad magic")
	}
	o := 4

	if e.Version != Version(data[o]) {
		return fmt.Errorf("envelope: unsupported version %d", data[o])
	}
	o++

	if data[o] != byte(e.KDF) {
		return fmt.Errorf("envelope: unsupported kdf %d", data[o])
	}
	o++

	if data[o] != byte(e.AEAD) {
		return fmt.Errorf("envelope: unsupported aead %d", data[o])
	}
	o++

	e.KDFParams.MemoryKiB = binary.BigEndian.Uint32(data[o:])
	o += 4
	e.KDFParams.Iterations = binary.BigEndian.Uint32(data[o:])
	o += 4
	e.KDFParams.Parallelism = data[o]
	o++

	// Sanity-bound the KDF params to prevent DoS on a hostile file
	if err := e.KDFParams.Validate(); err != nil {
		return fmt.Errorf("envelope: kdf params out of range: %w", err)
	}

	e.Salt = memguard.NewBufferFromBytes(data[o : o+SaltSize])
	o += SaltSize
	e.NonceKEK = memguard.NewBufferFromBytes(data[o : o+NonceSize])
	o += NonceSize
	e.WrappedDEK = memguard.NewBufferFromBytes(data[o : o+WrappedDEKSize])
	o += WrappedDEKSize
	e.NonceDEK = memguard.NewBufferFromBytes(data[o : o+NonceSize])
	o += NonceSize
	e.Ciphertext = memguard.NewBufferFromBytes(data[o:])

	return nil
}

func (e *Envelope) Destroy() {
	if e.Salt != nil {
		e.Salt.Destroy()
	}
	if e.NonceKEK != nil {
		e.NonceKEK.Destroy()
	}
	if e.NonceDEK != nil {
		e.NonceDEK.Destroy()
	}
	if e.WrappedDEK != nil {
		e.WrappedDEK.Destroy()
	}
	if e.Ciphertext != nil {
		e.Ciphertext.Destroy()
	}
}
