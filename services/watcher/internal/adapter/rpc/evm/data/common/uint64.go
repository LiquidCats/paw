package common // nolint:revive

import (
	"reflect"
	"strconv"
)

// Uint64 marshals/unmarshals as a JSON string with 0x prefix.
// The zero value marshals as "0x0".
type Uint64 uint64 // nolint:recvcheck

var uint64T = reflect.TypeFor[uint64]() // nolint:gochecknoglobals

// MarshalText implements encoding.TextMarshaler.
func (b Uint64) MarshalText() ([]byte, error) {
	buf := make([]byte, 2, 10) //nolint:mnd
	copy(buf, `0x`)
	buf = strconv.AppendUint(buf, uint64(b), 16) //nolint:mnd
	return buf, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Uint64) UnmarshalJSON(input []byte) error {
	if !isString(input) {
		return errNonString(uint64T)
	}

	return wrapTypeError(b.UnmarshalText(input[1:len(input)-1]), uint64T)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *Uint64) UnmarshalText(input []byte) error {
	raw, err := checkNumberText(input)
	if err != nil {
		return err
	}
	if len(raw) > 16 { //nolint:mnd
		return ErrUint64Range
	}
	var dec uint64
	for _, byte := range raw {
		nib := decodeNibble(byte)
		if nib == badNibble {
			return ErrSyntax
		}
		dec *= 16
		dec += nib
	}
	*b = Uint64(dec)
	return nil
}

// String returns the hex encoding of b.
func (b Uint64) String() string {
	return EncodeUint64(uint64(b))
}

// EncodeUint64 encodes i as a hex string with 0x prefix.
func EncodeUint64(i uint64) string {
	enc := make([]byte, 2, 10) //nolint:mnd
	copy(enc, "0x")
	return string(strconv.AppendUint(enc, i, 16)) //nolint:mnd
}
