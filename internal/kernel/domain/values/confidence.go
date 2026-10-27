package values

type TransferConfidence int

const (
	// TransferConfidenceLow is used for transfers
	// that are cant be parsed clearly.
	TransferConfidenceLow TransferConfidence = iota
	// TransferConfidenceMedium is used for transfers
	// that are parsed using parser but not clearly
	// identified: internal transfers of native token,
	// transfers via disperse contract, multicall
	// transactions.
	TransferConfidenceMedium
	// TransferConfidenceHigh is used for transfers
	// that are clearly parsed: native token transfers,
	// erc20 transfers.
	TransferConfidenceHigh
)
