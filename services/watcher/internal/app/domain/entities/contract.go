package entities

type Ticker string

type Contract struct {
	Address  Address
	Ticker   Ticker
	Decimals int32
}
