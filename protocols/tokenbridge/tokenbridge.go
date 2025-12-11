package tokenbridge

import (
	"context"
	"fmt"

	"github.com/wormhole-foundation/wormhole-go-sdk/types"
)

type TokenBridge interface {
	Transfer(ctx context.Context, token types.TokenID, amount types.Amount, recipient types.ChainAddress, signer types.Signer) (types.TxHash, error)

	CompleteTransfer(ctx context.Context, vaa []byte, signer types.Signer) (types.TxHash, error)

	AttestToken(ctx context.Context, token types.TokenID, signer types.Signer) (types.TxHash, error)

	CreateWrapped(ctx context.Context, vaa []byte, signer types.Signer) (types.TxHash, error)

	GetWrappedAsset(ctx context.Context, token types.TokenID) (string, error)

	IsWrappedAsset(ctx context.Context, token types.TokenID) (bool, error)
}

type TransferParams struct {
	Token      types.TokenID
	Amount     types.Amount
	Recipient  types.ChainAddress
	RelayerFee string
	Payload    []byte
	ArbiterFee string
}

type EVMTokenBridge struct {
	chain           types.Chain
	contractAddress string
	rpcClient       interface{}
}

func NewEVMTokenBridge(chain types.Chain, contractAddress string, rpcClient interface{}) *EVMTokenBridge {
	return &EVMTokenBridge{
		chain:           chain,
		contractAddress: contractAddress,
		rpcClient:       rpcClient,
	}
}

// Transfer initiates a token transfer
func (t *EVMTokenBridge) Transfer(ctx context.Context, token types.TokenID, amount types.Amount, recipient types.ChainAddress, signer types.Signer) (types.TxHash, error) {
	return "", fmt.Errorf("not implemented")
}

// CompleteTransfer completes a token transfer using a VAA
func (t *EVMTokenBridge) CompleteTransfer(ctx context.Context, vaa []byte, signer types.Signer) (types.TxHash, error) {
	return "", fmt.Errorf("not implemented")
}

// AttestToken attests a token for cross-chain transfer
func (t *EVMTokenBridge) AttestToken(ctx context.Context, token types.TokenID, signer types.Signer) (types.TxHash, error) {
	return "", fmt.Errorf("not implemented")
}

// CreateWrapped creates a wrapped token on the target chain
func (t *EVMTokenBridge) CreateWrapped(ctx context.Context, vaa []byte, signer types.Signer) (types.TxHash, error) {
	return "", fmt.Errorf("not implemented")
}

// GetWrappedAsset returns the wrapped asset address for a token
func (t *EVMTokenBridge) GetWrappedAsset(ctx context.Context, token types.TokenID) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// IsWrappedAsset checks if a token is a wrapped asset
func (t *EVMTokenBridge) IsWrappedAsset(ctx context.Context, token types.TokenID) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

// SolanaTokenBridge implements TokenBridge for Solana
type SolanaTokenBridge struct {
	chain     types.Chain
	programID string
	rpcClient any
}

// NewSolanaTokenBridge creates a new Solana token bridge client
func NewSolanaTokenBridge(chain types.Chain, programID string, rpcClient interface{}) *SolanaTokenBridge {
	return &SolanaTokenBridge{
		chain:     chain,
		programID: programID,
		rpcClient: rpcClient,
	}
}

// Transfer initiates a token transfer
func (t *SolanaTokenBridge) Transfer(ctx context.Context, token types.TokenID, amount types.Amount, recipient types.ChainAddress, signer types.Signer) (types.TxHash, error) {
	return "", fmt.Errorf("not implemented")
}

// CompleteTransfer completes a token transfer using a VAA
func (t *SolanaTokenBridge) CompleteTransfer(ctx context.Context, vaa []byte, signer types.Signer) (types.TxHash, error) {
	return "", fmt.Errorf("not implemented")
}

// AttestToken attests a token for cross-chain transfer
func (t *SolanaTokenBridge) AttestToken(ctx context.Context, token types.TokenID, signer types.Signer) (types.TxHash, error) {
	return "", fmt.Errorf("not implemented")
}

// CreateWrapped creates a wrapped token on the target chain
func (t *SolanaTokenBridge) CreateWrapped(ctx context.Context, vaa []byte, signer types.Signer) (types.TxHash, error) {
	return "", fmt.Errorf("not implemented")
}

// GetWrappedAsset returns the wrapped asset address for a token
func (t *SolanaTokenBridge) GetWrappedAsset(ctx context.Context, token types.TokenID) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// IsWrappedAsset checks if a token is a wrapped asset
func (t *SolanaTokenBridge) IsWrappedAsset(ctx context.Context, token types.TokenID) (bool, error) {
	return false, fmt.Errorf("not implemented")
}
