package data

import (
	"strings"

	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
	"golang.org/x/crypto/sha3"
)

func AddressToChecksumAddress[T ~string](address T) entities.Address {
	addr := string(address[len(address)-40:])

	addr = strings.ToLower(addr)

	// Keccak-256 hash of the lowercase address
	hash := keccak256([]byte(addr))

	// Build checksummed address
	var result strings.Builder
	result.WriteString("0x")

	for i, c := range addr {
		if c >= '0' && c <= '9' { //nolint:nestif
			// Digits stay as-is
			result.WriteByte(addr[i])
		} else {
			// Letters: uppercase if corresponding hash nibble >= 8
			hashNibble := hash[i/2]
			if i%2 == 0 {
				hashNibble >>= 4 //nolint:mnd
			} else {
				hashNibble &= 0x0F //nolint:mnd
			}

			if hashNibble >= 8 { //nolint:mnd
				result.WriteByte(addr[i] - 32) //nolint:mnd    // to uppercase
			} else {
				result.WriteByte(addr[i])
			}
		}
	}

	return entities.Address(result.String())
}

func keccak256(data []byte) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write(data)
	return h.Sum(nil)
}
