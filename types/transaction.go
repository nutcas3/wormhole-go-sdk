package types

type UnsignedTransaction struct {
	Chain       Chain
	Transaction any /// Platform-specific transaction data
	Description string
}

type SignedTransaction struct {
	Chain       Chain
	Transaction []byte // Serialized signed transaction
	TxHash      string
}

type TxHash string

func (h TxHash) String() string {
	return string(h)
}
