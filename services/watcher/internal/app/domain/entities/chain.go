package entities

type Driver string

const (
	DriverRPC Driver = "rpc"
	DriverP2P Driver = "p2p"
)

func (d Driver) Equals(drv Driver) bool {
	return d == drv
}

type Type string

const (
	TypeEvm  Type = "evm"
	TypeUtxo Type = "utxo"
)

func (t Type) Equals(typ Type) bool {
	return t == typ
}

type Chain string
