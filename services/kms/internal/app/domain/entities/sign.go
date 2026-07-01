package entities

type SignatureEncoding string

const (
	SignatureEncodingHex    SignatureEncoding = "hex"
	SignatureEncodingBase58 SignatureEncoding = "base58"
)

type Signature []byte
type PublicKey []byte
