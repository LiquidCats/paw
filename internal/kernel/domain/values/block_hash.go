package values

type BlockHash string

func (b BlockHash) String() string {
	return string(b)
}
