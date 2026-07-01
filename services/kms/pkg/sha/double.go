package sha

import "crypto/sha256"

func DoubleSHA256(input []byte) []byte {
	h := sha256.Sum256(input)
	h2 := sha256.Sum256(h[:])
	return h2[:]
}

func DoubleSHA256Checksum(input []byte) (cksum [4]byte) { // nolint:nonamedreturns
	hash := DoubleSHA256(input)

	copy(cksum[:], hash[:4])

	return cksum
}
