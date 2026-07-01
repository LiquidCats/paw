package base58

import (
	"math/big"

	"github.com/LiquidCats/paw/services/litehsm/pkg/sha"
)

// CheckDecode decodes a Base58Check encoded string and verifies the checksum.
func (c *Codec) CheckDecode(input string) (result []byte, err error) { // nolint:nonamedreturns
	decoded := c.Decode(input)
	if len(decoded) < 5 {
		return nil, ErrInvalidFormat
	}

	var cksum [4]byte
	copy(cksum[:], decoded[len(decoded)-4:])
	if sha.DoubleSHA256Checksum(decoded[:len(decoded)-4]) != cksum {
		return nil, ErrChecksum
	}

	result = decoded[:len(decoded)-4]
	return
}

// Decode decodes a modified base58 string to a byte slice.
func (c *Codec) Decode(b string) []byte {
	answer := big.NewInt(0)
	scratch := new(big.Int)

	// Calculating with big.Int is slow for each iteration.
	//    x += b58[b[i]] * j
	//    j *= 58
	//
	// Instead we can try to do as much calculations on int64.
	// We can represent a 10 digit base58 number using an int64.
	//
	// Hence we'll try to convert 10, base58 digits at a time.
	// The rough idea is to calculate `t`, such that:
	//
	//   t := b58[b[i+9]] * 58^9 ... + b58[b[i+1]] * 58^1 + b58[b[i]] * 58^0
	//   x *= 58^10
	//   x += t
	//
	// Of course, in addition, we'll need to handle boundary condition when `b` is not multiple of 58^10.
	// In that case we'll use the bigRadix[n] lookup for the appropriate power.

	for t := b; len(t) > 0; {
		n := min(len(t), 10)

		total := uint64(0)
		for _, v := range t[:n] {
			tmp := c.lookupTable[v]
			if tmp == 255 {
				return []byte("")
			}
			total = total*58 + uint64(tmp)
		}

		answer.Mul(answer, bigRadix[n])
		scratch.SetUint64(total)
		answer.Add(answer, scratch)

		t = t[n:]
	}

	tmpval := answer.Bytes()

	var numZeros int
	for numZeros = 0; numZeros < len(b); numZeros++ { // nolint:intrange
		if b[numZeros] != c.alphabetIdx0 {
			break
		}
	}
	flen := numZeros + len(tmpval)
	val := make([]byte, flen)
	copy(val[numZeros:], tmpval)

	return val
}
