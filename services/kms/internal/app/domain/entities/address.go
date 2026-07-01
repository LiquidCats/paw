package entities

type Address string
type AddressEncoding string

const (
	AddressEncodingKeccakEvm AddressEncoding = "keccak:evm"

	AddressEncodingBase58Tron   AddressEncoding = "base58:tron"
	AddressEncodingBase58Ripple AddressEncoding = "base58:ripple"
	AddressEncodingBase58P2SH   AddressEncoding = "base58:p2sh"
	AddressEncodingBase58P2PKH  AddressEncoding = "base58:p2pkh"

	AddressEncodingBech32P2WSH    AddressEncoding = "bech32:p2wsh"
	AddressEncodingBech32P2WPKH   AddressEncoding = "bech32:p2wpkh"
	AddressEncodingBech32CashAddr AddressEncoding = "bech32:cashaddr"
)
