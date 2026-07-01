package entities

type (
	BlockHeight uint64
	BlockHash   string
	RawBlock    string
)

type BlockHeader struct {
	Height   BlockHeight
	Hash     BlockHash
	PrevHash BlockHash
}

type Block struct {
	BlockHeader
	Transactions []TxID
}

type BlockWithTransactions[TxIn any] struct {
	BlockHeader
	Transactions []Transaction[TxIn]
}
