package types

import "context"

type Signer interface {
	Chain() Chain
	Address() string
}

type SignOnlySigner interface {
	Signer
	Sign(ctx context.Context, txs []UnsignedTransaction) ([]SignedTransaction, error)
}

type SignAndSendSigner interface {
	Signer
	SignAndSend(ctx context.Context, txs []UnsignedTransaction) ([]TxHash, error)
}
