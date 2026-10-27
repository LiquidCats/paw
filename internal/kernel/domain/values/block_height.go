package values

import "fmt"

type BlockHeight uint64

func (h BlockHeight) Int() int {
	return int(h) //nolint:gosec
}

func (h BlockHeight) Uint() uint {
	return uint(h)
}

func (h BlockHeight) Hex() string {
	return fmt.Sprintf("0x%x", h)
}

func (h BlockHeight) String() string {
	return fmt.Sprintf("%d", h)
}
