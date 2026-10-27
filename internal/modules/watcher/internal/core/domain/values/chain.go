package values

type Chain string

func (c Chain) String() string {
	return string(c)
}

func (c Chain) Equals(other Chain) bool {
	return c == other
}
