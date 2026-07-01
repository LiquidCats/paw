package entities

import (
	"errors"

	v1 "github.com/LiquidCats/paw/protos/gen/go/services/litehsm/v1"
)

type CurveType string

const (
	CurveTypeSecp256k1 CurveType = "secp256k1"
)

var ErrInvalidCurveType = errors.New("invalid curve type")

func CurveTypeFromProto(s v1.Curve) (CurveType, error) {
	switch s { //nolint:exhaustive
	case v1.Curve_CURVE_SECP256K1:
		return CurveTypeSecp256k1, nil
	default:
		return "", ErrInvalidCurveType
	}
}

func (c CurveType) ToProto() v1.Curve {
	switch c {
	case CurveTypeSecp256k1:
		return v1.Curve_CURVE_SECP256K1
	default:
		return v1.Curve_CURVE_UNSPECIFIED
	}
}

type AlgorithmType string

const (
	AlgorithmTypeECDSA AlgorithmType = "ecdsa"
)

var ErrInvalidAlgorithmType = errors.New("invalid algorithm type")

func AlgorithmTypeFromProto(s v1.Algorithm) (AlgorithmType, error) {
	switch s { //nolint:exhaustive
	case v1.Algorithm_ALGORITHM_ECDSA:
		return AlgorithmTypeECDSA, nil
	default:
		return "", ErrInvalidAlgorithmType
	}
}

func (a AlgorithmType) ToProto() v1.Algorithm {
	switch a {
	case AlgorithmTypeECDSA:
		return v1.Algorithm_ALGORITHM_ECDSA
	default:
		return v1.Algorithm_ALGORITHM_UNSPECIFIED
	}
}
