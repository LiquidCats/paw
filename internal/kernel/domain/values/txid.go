package values

type TxID string

func (t TxID) String() string {
	return string(t)
}
