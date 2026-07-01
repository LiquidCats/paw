package entities

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/rotisserie/eris"
)

type DerivationPath []Index

func ParseDerivationPath(path string) (DerivationPath, error) {
	unprefixed := strings.TrimPrefix(path, "m/")

	parts := strings.Split(unprefixed, "/")

	d := make(DerivationPath, len(parts))

	var err error

	for i, part := range parts {
		d[i], err = ParseIndex(part)
		if err != nil {
			return nil, fmt.Errorf("func=ParseDerivationPath: %w", err)
		}
	}

	return d, nil
}

func (d DerivationPath) String() string {
	res := make([]string, 0, len(d)+1)
	res = append(res, "m")
	for _, index := range d {
		res = append(res, index.String())
	}

	return strings.Join(res, "/")
}

type Index struct { // nolint:recvcheck
	number   uint32
	hardened bool
}

const (
	HardenedKeyIndex uint32 = 1 << 31
)

func RandomHardenedKeyIndex() Index {
	i, _ := rand.Int(rand.Reader, big.NewInt(int64(HardenedKeyIndex-1)))
	return Index{
		number:   uint32(i.Uint64()), //nolint:gosec
		hardened: true,
	}
}

func NewIndex(number uint32, hardened bool) Index {
	return Index{
		number:   number,
		hardened: hardened,
	}
}

func IndexFromUint32(number uint32) Index {
	var isHardened bool
	if number >= HardenedKeyIndex {
		isHardened = true
		number -= HardenedKeyIndex
	}

	return Index{
		number:   number,
		hardened: isHardened,
	}
}

func ParseIndex(index string) (Index, error) {
	idxStr, isHardened := strings.CutSuffix(index, "'")

	idx, err := strconv.ParseUint(idxStr, 10, 32)
	if err != nil {
		return Index{}, eris.Wrap(err, "failed to parse index")
	}

	idxu32 := uint32(idx)

	if idxu32 >= HardenedKeyIndex {
		isHardened = true
	}

	return Index{
		number:   idxu32,
		hardened: isHardened,
	}, nil
}

func (i Index) Uint32() uint32 {
	if i.hardened {
		return i.number + HardenedKeyIndex
	}
	return i.number
}

func (i Index) IsHardened() bool {
	return i.hardened
}

func (i *Index) Incr() {
	i.number++
}

func (i Index) String() string {
	s := strconv.FormatUint(uint64(i.number), 10)
	if i.hardened {
		s += "'"
	}
	return s
}

func (i *Index) UnmarshalText(b []byte) error {
	return i.unmarshalTextType(string(b))
}

func (i *Index) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return eris.Wrap(err, "cannot unmarshal Index from json")
	}

	return i.unmarshalTextType(str)
}

func (i *Index) UnmarshalYAML(b []byte) error {
	var str string
	if err := yaml.Unmarshal(b, &str); err != nil {
		return eris.Wrap(err, "cannot unmarshal Index from yaml")
	}

	return i.unmarshalTextType(str)
}

func (i *Index) unmarshalTextType(str string) error {
	if str == "" || str == "null" || str == "{}" || str == "[]" || str == "nil" {
		return eris.New("cannot unmarshal an empty Index")
	}

	numberStr, isHardened := strings.CutSuffix(str, "'")

	numberUint32, err := strconv.ParseUint(numberStr, 10, 32)
	if err != nil {
		return eris.Wrap(err, "cannot unmarshal an Index")
	}

	*i = Index{
		number:   uint32(numberUint32),
		hardened: isHardened,
	}

	return nil
}
