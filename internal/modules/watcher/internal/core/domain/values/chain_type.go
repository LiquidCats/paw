package values

import "slices"

type ChainType string

const (
	TypeEvm  ChainType = "evm"
	TypeUtxo ChainType = "utxo"
)

func (t ChainType) Equals(typ ChainType) bool {
	return t == typ
}

func (c ChainType) OneOf(other ...ChainType) bool {
	return slices.ContainsFunc(other, c.Equals)
}

func (t ChainType) IsValid() bool {
	switch t {
	case TypeEvm, TypeUtxo:
		return true
	default:
		return false
	}
}

func (t ChainType) String() string {
	return string(t)
}
