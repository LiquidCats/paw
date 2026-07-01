package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/bits"
	"reflect"
)

// Errors.
var (
	ErrEmptyString   = &decError{"empty hex string"}
	ErrSyntax        = &decError{"invalid hex string"}
	ErrMissingPrefix = &decError{"hex string without 0x prefix"}
	ErrOddLength     = &decError{"hex string of odd length"}
	ErrEmptyNumber   = &decError{"hex string \"0x\""}
	ErrLeadingZero   = &decError{"hex number with leading zero digits"}
	ErrUint64Range   = &decError{"hex number > 64 bits"}
	ErrUintRange     = &decError{fmt.Sprintf("hex number > %d bits", bits.UintSize)}
	ErrBig256Range   = &decError{"hex number > 256 bits"}
)

func wrapTypeError(err error, typ reflect.Type) error {
	decError := &decError{}
	if errors.As(err, &decError) {
		return &json.UnmarshalTypeError{Value: err.Error(), Type: typ}
	}
	return err
}

type decError struct{ msg string }

func (err decError) Error() string { return err.msg }

func errNonString(typ reflect.Type) error {
	return &json.UnmarshalTypeError{Value: "non-string", Type: typ}
}
