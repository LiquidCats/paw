package entities

import (
	"errors"
	"fmt"

	"github.com/LiquidCats/paw/services/litehsm/pkg/unsafe"
	"gopkg.in/yaml.v3"
)

var ErrInvalidChain = errors.New("invalid chain")

type Chain string //nolint:recvcheck

const (
	ChainBitcoin     = Chain("bitcoin")
	ChainLitecoin    = Chain("litecoin")
	ChainDogecoin    = Chain("dogecoin")
	ChainBitcoinCash = Chain("bitcoin_cash")

	ChainEthereum = Chain("ethereum")
	ChainArbitrum = Chain("arbitrum")
	ChainBase     = Chain("base")

	ChainTron   = Chain("tron")
	ChainRipple = Chain("ripple")
)

func (c *Chain) UnmarshalText(text []byte) error {
	chn := Chain(unsafe.BytesToString(text))
	if !chn.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidChain, chn)
	}

	*c = chn

	return nil
}

func (c *Chain) UnmarshalYAML(d *yaml.Node) error {
	var value string

	if err := d.Decode(value); err != nil {
		return fmt.Errorf("decoding YAML: %w", err)
	}

	chn := Chain(value)

	if !chn.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidChain, chn)
	}

	*c = chn

	return nil
}

func (c Chain) IsValid() bool {
	switch c {
	case ChainBitcoin:
	case ChainLitecoin:
	case ChainDogecoin:
	case ChainEthereum:
	case ChainArbitrum:
	case ChainBase:
	case ChainTron:
	case ChainRipple:
	case ChainBitcoinCash:
		return true
	}

	return false
}

func (c Chain) Family() ChainFamily {
	switch c {
	case ChainBitcoin, ChainLitecoin, ChainDogecoin, ChainBitcoinCash:
		return ChainFamilyUTXO
	case ChainEthereum, ChainArbitrum, ChainBase:
		return ChainFamilyEVM
	case ChainTron, ChainRipple:
		return ChainFamilyOther
	}

	return ChainFamilyOther
}

func (c Chain) String() string {
	return string(c)
}

type ChainFamily string

const (
	ChainFamilyOther ChainFamily = "other"
	ChainFamilyEVM   ChainFamily = "evm"
	ChainFamilyUTXO  ChainFamily = "utxo"
)
