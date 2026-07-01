package common

import "math/big"

var bigWordNibbles int // nolint:gochecknoglobals

//nolint:gochecknoinits
func init() {
	// This is a weird way to compute the number of nibbles required for big.Word.
	// The usual way would be to use constant arithmetic but go vet can't handle that.
	b, _ := new(big.Int).SetString("FFFFFFFFFF", 16) // nolint:mnd
	switch len(b.Bits()) {
	case 1:
		bigWordNibbles = 16
	case 2: //nolint:mnd
		bigWordNibbles = 8
	default:
		panic("weird big.Word size")
	}
}

const badNibble = ^uint64(0)

func decodeNibble(in byte) uint64 {
	switch {
	case in >= '0' && in <= '9':
		return uint64(in - '0')
	case in >= 'A' && in <= 'F':
		return uint64(in - 'A' + 10) // nolint:mnd
	case in >= 'a' && in <= 'f':
		return uint64(in - 'a' + 10) // nolint:mnd
	default:
		return badNibble
	}
}

func checkNumberText(input []byte) (raw []byte, err error) { // nolint:nonamedreturns
	if len(input) == 0 {
		return nil, nil // empty strings are allowed
	}
	if !bytesHave0xPrefix(input) {
		return nil, ErrMissingPrefix
	}
	input = input[2:]
	if len(input) == 0 {
		return nil, ErrEmptyNumber
	}
	if len(input) > 1 && input[0] == '0' {
		return nil, ErrLeadingZero
	}
	return input, nil
}

func bytesHave0xPrefix(input []byte) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}
