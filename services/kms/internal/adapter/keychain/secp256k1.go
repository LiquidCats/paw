package keychain

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/ports"
	"github.com/LiquidCats/paw/services/litehsm/pkg/unsafe"
	"github.com/awnumar/memguard"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const maxDepth = 0xFF

var (
	ErrEmptySeed       = errors.New("empty seed")
	ErrUnusableSeed    = errors.New("unusable seed")
	ErrInvalidChildKey = errors.New("invalid child key")
	ErrInvalidSeedLen  = errors.New("invalid seed length")
	ErrMaxDepth        = errors.New("max depth")
)

// masterHMACKey is the HMAC-SHA512 key used to derive the master node (BIP-32).
const masterHMACKey = "Bitcoin seed"

const (
	// MinSeedBytes is the minimum number of bytes allowed for a seed to a master node.
	MinSeedBytes = 16 // 128 bits
	// MaxSeedBytes is the maximum number of bytes allowed for a seed to a master node.
	MaxSeedBytes = 64 // 512 bits
)

// Secp256k1Keychain represents a BIP-32 extended private key.
// chainCode and keyData are stored in memguard locked buffers — memory-locked,
// guard-paged, and wiped on Destroy.
type Secp256k1Keychain struct {
	depth     uint8
	chainCode *memguard.LockedBuffer
	keyData   *memguard.LockedBuffer
}

// NewSecp256k1Keychain derives the BIP-32 master extended key from a BIP-39 mnemonic and
// optional passphrase. All intermediate key material is wiped after use.
func NewSecp256k1Keychain(seed *memguard.Enclave) (*Secp256k1Keychain, error) {
	if seed == nil {
		return nil, ErrEmptySeed
	}

	if seed.Size() < MinSeedBytes || seed.Size() > MaxSeedBytes {
		return nil, ErrInvalidSeedLen
	}

	seedBuf, err := seed.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open seed buffer: %w", err)
	}

	hmac512 := hmac.New(sha512.New, unsafe.StringToBytes(masterHMACKey))
	hmac512.Write(seedBuf.Bytes())
	lr := hmac512.Sum(nil)
	defer memguard.WipeBytes(lr)

	// Il = master secret key (lr[:32]), Ir = master chain code (lr[32:])
	secretKey := lr[:32]
	chainCode := lr[32:]

	// Validate the key before moving it to locked memory.
	var priv secp256k1.ModNScalar
	overflow := priv.SetByteSlice(secretKey)
	isZero := priv.IsZero()
	priv.Zero()
	if overflow || isZero {
		return nil, ErrUnusableSeed
	}

	return &Secp256k1Keychain{
		// NewBufferFromBytes copies into locked memory and wipes the source slice.
		chainCode: memguard.NewBufferFromBytes(chainCode),
		keyData:   memguard.NewBufferFromBytes(secretKey),
	}, nil
}

// Destroy wipes and frees all key material held by the extended key.
func (k *Secp256k1Keychain) Destroy() {
	if k.chainCode != nil {
		k.chainCode.Destroy()
		k.chainCode = nil
	}
	if k.keyData != nil {
		k.keyData.Destroy()
		k.keyData = nil
	}
}

// Derive produces a child BIP-32 extended key at the given index.
// All intermediate buffers containing key material are wiped before return.
func (k *Secp256k1Keychain) Derive(
	index entities.Index,
) (ports.Keychain[secp256k1.PublicKey, secp256k1.PrivateKey], error) {
	if k.depth == maxDepth {
		return nil, ErrMaxDepth
	}

	// data is the 37-byte input to HMAC-SHA512: either 0x00||privkey or pubkey,
	// followed by a 4-byte big-endian index.
	data := make([]byte, 37) // 33 key bytes + 4 index bytes
	defer memguard.WipeBytes(data)

	if index.IsHardened() {
		// Hardened child: 0x00 || privkey (right-aligned into 33 bytes).
		offset := 33 - len(k.keyData.Bytes())
		copy(data[offset:], k.keyData.Bytes())
	} else {
		// Normal child: compressed public key.
		pubKey := secp256k1.PrivKeyFromBytes(k.keyData.Bytes()).PubKey().SerializeCompressed()
		copy(data, pubKey)
	}
	binary.BigEndian.PutUint32(data[33:], index.Uint32())

	hmac512 := hmac.New(sha512.New, k.chainCode.Bytes())
	hmac512.Write(data)
	ilr := hmac512.Sum(nil)
	defer memguard.WipeBytes(ilr)

	// Il = intermediate key (ilr[:32]), Ir = child chain code (ilr[32:])
	il := ilr[:32]
	childChainCode := ilr[32:]

	var ilModN secp256k1.ModNScalar
	if overflow := ilModN.SetByteSlice(il); overflow || ilModN.IsZero() {
		return nil, ErrInvalidChildKey
	}

	// childKey = parse256(Il) + parentKey  (mod n)
	var parentPrivKeyModN secp256k1.ModNScalar
	parentPrivKeyModN.SetByteSlice(k.keyData.Bytes())
	ilModN.Add(&parentPrivKeyModN)
	parentPrivKeyModN.Zero()

	childKeyArr := ilModN.Bytes() // [32]byte on the stack
	ilModN.Zero()

	childKey := childKeyArr[:]
	// Strip leading zeroes (legacy behaviour matching prior big.Int usage).
	for len(childKey) > 0 && childKey[0] == 0x00 {
		childKey = childKey[1:]
	}

	// Copy into a separate slice so NewBufferFromBytes wipes only that copy,
	// then zero the full stack array explicitly.
	keyMaterial := make([]byte, len(childKey))
	copy(keyMaterial, childKey)
	keyDataBuf := memguard.NewBufferFromBytes(keyMaterial) // wipes keyMaterial
	for i := range childKeyArr {
		childKeyArr[i] = 0
	}

	return &Secp256k1Keychain{
		depth:     k.depth + 1,
		chainCode: memguard.NewBufferFromBytes(childChainCode), // wipes ilr[32:]
		keyData:   keyDataBuf,
	}, nil
}

// PublicKey derives and returns the corresponding public key,
// optionally in compressed format, from the extended private key.
func (k *Secp256k1Keychain) PublicKey() *secp256k1.PublicKey {
	return k.PrivateKey().PubKey()
}

// PrivateKey returns the secp256k1 private key derived from the extended key's key material.
func (k *Secp256k1Keychain) PrivateKey() *secp256k1.PrivateKey {
	return secp256k1.PrivKeyFromBytes(k.keyData.Bytes())
}
