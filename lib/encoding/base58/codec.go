package base58

type Codec struct {
	alphabet     string
	alphabetIdx0 byte
	lookupTable  *[256]byte
}

func NewCodec(alphabet string, lookupTable *[256]byte) *Codec {
	return &Codec{
		alphabet:     alphabet,
		alphabetIdx0: alphabet[0],
		lookupTable:  lookupTable,
	}
}

var (
	BitcoinCodec = NewCodec(bitcoinAlphabet, &bitcoin58)
	RippleCodec  = NewCodec(rippleAlphabet, &ripple58)
)
