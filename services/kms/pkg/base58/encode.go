package base58

import (
	"math/big"

	"github.com/LiquidCats/paw/services/litehsm/pkg/sha"
)

// CheckEncode prepends a version byte, appends a four-byte checksum, and returns
// the Base58Check encoding of the input byte slice.
func (c *Codec) CheckEncode(input []byte, prefix ...byte) string {
	b := make([]byte, 0, 1+len(input)+4)
	b = append(b, prefix...)
	b = append(b, input...)

	cksum := sha.DoubleSHA256Checksum(b)
	b = append(b, cksum[:]...)
	return c.Encode(b)
}

// Encode encodes a byte slice to a base58 string using the XRP alphabet.
func (c *Codec) Encode(b []byte) string {
	x := new(big.Int)
	x.SetBytes(b)

	// Maximum length of output is log58(2^(8*len(b))) == len(b) * 8 / log(58)
	maxlen := int(float64(len(b))*1.365658237309761) + 1
	answer := make([]byte, 0, maxlen)
	mod := new(big.Int)

	for x.Sign() > 0 {
		// Calculating with big.Int is slow for each iteration.
		//    x, mod = x / 58, x % 58
		//
		// Instead we can try to do as much calculations on int64.
		//    x, mod = x / 58^10, x % 58^10
		//
		// Which will give us mod, which is 10 digit base58 number.
		// We'll loop that 10 times to convert to the answer.

		x.DivMod(x, bigRadix10, mod)

		if x.Sign() == 0 {
			// When x = 0, we need to ensure we don't add any extra zeros.
			m := mod.Int64()
			for m > 0 {
				answer = append(answer, c.alphabet[m%58])
				m /= 58
			}
		} else {
			m := mod.Int64()
			for range 10 {
				answer = append(answer, c.alphabet[m%58])
				m /= 58
			}
		}
	}
	// leading zero bytes
	for _, i := range b {
		if i != 0 {
			break
		}
		answer = append(answer, c.alphabetIdx0)
	}

	// reverse
	alen := len(answer)
	for i := range alen / 2 {
		answer[i], answer[alen-1-i] = answer[alen-1-i], answer[i]
	}

	return string(answer)
}
