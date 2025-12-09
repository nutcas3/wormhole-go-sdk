package cctp

import (
	"context"
	"fmt"

	"github.com/wormhole-foundation/wormhole-go-sdk/types"
)

type CCTP interface {
	Transfer(ctx context.Context, amount types.Amount, recipient types.ChainAddress, signer types.Signer) (types.TxHash, error)

	CompleteTransfer(ctx context.Context, message []byte, attestation []byte, signer types.Signer) (types.TxHash, error)

	GetAttestation(ctx context.Context, messageHash []byte) ([]byte, error)
}

type TransferParams struct {
	Amount            types.Amount
	Recipient         types.ChainAddress
	DestinationDomain uint32
}

// EVMCCTP implements CCTP for EVM chains
type EVMCCTP struct {
	chain              types.Chain
	tokenMessengerAddr string
	messageTransmitter string
	rpcClient          any
}

// NewEVMCCTP creates a new EVM CCTP client
func NewEVMCCTP(chain types.Chain, tokenMessengerAddr, messageTransmitter string, rpcClient any) *EVMCCTP {
	return &EVMCCTP{
		chain:              chain,
		tokenMessengerAddr: tokenMessengerAddr,
		messageTransmitter: messageTransmitter,
		rpcClient:          rpcClient,
	}
}

// Transfer initiates a native USDC transfer
func (c *EVMCCTP) Transfer(ctx context.Context, amount types.Amount, recipient types.ChainAddress, signer types.Signer) (types.TxHash, error) {
	return "", fmt.Errorf("not implemented")
}

// CompleteTransfer completes a native USDC transfer
func (c *EVMCCTP) CompleteTransfer(ctx context.Context, message []byte, attestation []byte, signer types.Signer) (types.TxHash, error) {
	return "", fmt.Errorf("not implemented")
}

// GetAttestation fetches the attestation for a message
func (c *EVMCCTP) GetAttestation(ctx context.Context, messageHash []byte) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

// SolanaCCTP implements CCTP for Solana
type SolanaCCTP struct {
	chain              types.Chain
	tokenMessengerAddr string
	messageTransmitter string
	rpcClient          any
}

// NewSolanaCCTP creates a new Solana CCTP client
func NewSolanaCCTP(chain types.Chain, tokenMessengerAddr, messageTransmitter string, rpcClient any) *SolanaCCTP {
	return &SolanaCCTP{
		chain:              chain,
		tokenMessengerAddr: tokenMessengerAddr,
		messageTransmitter: messageTransmitter,
		rpcClient:          rpcClient,
	}
}

// Transfer initiates a native USDC transfer
func (s *SolanaCCTP) Transfer(ctx context.Context, amount types.Amount, recipient types.ChainAddress, signer types.Signer) (types.TxHash, error) {
	return "", fmt.Errorf("not implemented")
}

// CompleteTransfer completes a native USDC transfer
func (s *SolanaCCTP) CompleteTransfer(ctx context.Context, message []byte, attestation []byte, signer types.Signer) (types.TxHash, error) {
	return "", fmt.Errorf("not implemented")
}

// GetAttestation fetches the attestation for a message
func (s *SolanaCCTP) GetAttestation(ctx context.Context, messageHash []byte) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}
