package cctp

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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
	Nonce             uint64
}

// CircleDomain represents Circle's domain identifiers
type CircleDomain uint32

const (
	DomainEthereum  CircleDomain = 0
	DomainAvalanche CircleDomain = 1
	DomainOptimism  CircleDomain = 2
	DomainArbitrum  CircleDomain = 3
	DomainSolana    CircleDomain = 5
	DomainBase      CircleDomain = 6
	DomainPolygon   CircleDomain = 7
)

// AttestationResponse represents Circle's attestation API response
type AttestationResponse struct {
	Status      string `json:"status"`
	Attestation string `json:"attestation"`
}

// GetCircleDomain returns the Circle domain for a chain
func GetCircleDomain(chain types.Chain) (CircleDomain, error) {
	switch chain {
	case types.Ethereum, types.Sepolia:
		return DomainEthereum, nil
	case types.Avalanche:
		return DomainAvalanche, nil
	case types.Optimism:
		return DomainOptimism, nil
	case types.Arbitrum:
		return DomainArbitrum, nil
	case types.Solana:
		return DomainSolana, nil
	case types.Base:
		return DomainBase, nil
	case types.Polygon:
		return DomainPolygon, nil
	default:
		return 0, fmt.Errorf("chain %s not supported by CCTP", chain)
	}
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
	// Validate chain compatibility
	if signer.Chain() != c.chain {
		return "", fmt.Errorf("signer chain %s does not match CCTP chain %s", signer.Chain(), c.chain)
	}

	// Get destination domain
	destDomain, err := GetCircleDomain(recipient.Chain)
	if err != nil {
		return "", fmt.Errorf("failed to get destination domain: %w", err)
	}

	// Convert recipient address to bytes32 format
	recipientAddr, err := recipient.ToUniversalAddress()
	if err != nil {
		return "", fmt.Errorf("failed to convert recipient address: %w", err)
	}

	// In a real implementation, this would:
	// 1. Get USDC token address for this chain
	// 2. Approve TokenMessenger to spend USDC
	// 3. Call depositForBurn or depositForBurnWithCaller on TokenMessenger
	// 4. Parse MessageSent event from transaction logs
	// 5. Extract message hash and nonce

	// Pseudo-code for the actual implementation:
	// usdcToken := getUSDCAddress(c.chain)
	// approveTx := approveToken(usdcToken, c.tokenMessengerAddr, amount)
	// depositTx := depositForBurn(amount, destDomain, recipientAddr, usdcToken)
	// messageHash := parseMessageSentEvent(depositTx.Logs)

	// For now, return placeholder
	return types.TxHash(fmt.Sprintf("0x%s", hex.EncodeToString(recipientAddr[:8]))), nil
}

// CompleteTransfer completes a native USDC transfer
func (c *EVMCCTP) CompleteTransfer(ctx context.Context, message []byte, attestation []byte, signer types.Signer) (types.TxHash, error) {
	// Validate inputs
	if len(message) == 0 {
		return "", fmt.Errorf("message cannot be empty")
	}
	if len(attestation) == 0 {
		return "", fmt.Errorf("attestation cannot be empty")
	}

	// Validate signer chain
	if signer.Chain() != c.chain {
		return "", fmt.Errorf("signer chain %s does not match CCTP chain %s", signer.Chain(), c.chain)
	}

	// In a real implementation, this would:
	// 1. Call receiveMessage on MessageTransmitter contract
	// 2. Pass the message and attestation
	// 3. Contract verifies attestation signature
	// 4. Contract mints USDC to recipient
	// 5. Return transaction hash

	// Pseudo-code:
	// tx := messageTransmitter.receiveMessage(message, attestation)
	// waitForConfirmation(tx)
	// return tx.Hash

	// For now, return placeholder
	return types.TxHash(fmt.Sprintf("0x%s", hex.EncodeToString(message[:8]))), nil
}

// GetAttestation fetches the attestation for a message from Circle's API
func (c *EVMCCTP) GetAttestation(ctx context.Context, messageHash []byte) ([]byte, error) {
	if len(messageHash) != 32 {
		return nil, fmt.Errorf("message hash must be 32 bytes, got %d", len(messageHash))
	}

	// Circle's attestation API endpoint
	// Testnet: https://iris-api-sandbox.circle.com
	// Mainnet: https://iris-api.circle.com
	apiURL := "https://iris-api-sandbox.circle.com/v1/attestations/" + hex.EncodeToString(messageHash)

	// Poll for attestation with timeout
	maxAttempts := 60
	pollInterval := 2 * time.Second

	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := http.Get(apiURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch attestation: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response: %w", err)
			}

			var attestResp AttestationResponse
			if err := json.Unmarshal(body, &attestResp); err != nil {
				return nil, fmt.Errorf("failed to parse attestation response: %w", err)
			}

			if attestResp.Status == "complete" && attestResp.Attestation != "" {
				// Decode hex attestation
				attestation, err := hex.DecodeString(attestResp.Attestation)
				if err != nil {
					return nil, fmt.Errorf("failed to decode attestation: %w", err)
				}
				return attestation, nil
			}
		}

		// Wait before next attempt
		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("attestation not available after %d attempts", maxAttempts)
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
	// Validate chain compatibility
	if signer.Chain() != s.chain {
		return "", fmt.Errorf("signer chain %s does not match CCTP chain %s", signer.Chain(), s.chain)
	}

	// Get destination domain
	destDomain, err := GetCircleDomain(recipient.Chain)
	if err != nil {
		return "", fmt.Errorf("failed to get destination domain: %w", err)
	}

	// Convert recipient address
	recipientAddr, err := recipient.ToUniversalAddress()
	if err != nil {
		return "", fmt.Errorf("failed to convert recipient address: %w", err)
	}

	// In a real implementation, this would:
	// 1. Get USDC token mint address for Solana
	// 2. Create approve instruction for Token Messenger
	// 3. Create depositForBurn instruction
	// 4. Build and sign transaction
	// 5. Send transaction to Solana network
	// 6. Parse MessageSent event from transaction
	// 7. Extract message hash

	// Pseudo-code:
	// usdcMint := getUSDCMint()
	// approveIx := createApproveInstruction(signer, s.tokenMessengerAddr, amount)
	// depositIx := createDepositForBurnInstruction(amount, destDomain, recipientAddr)
	// tx := buildTransaction([approveIx, depositIx])
	// signature := signer.Sign(tx)
	// txHash := sendTransaction(signature)

	// For now, return placeholder
	return types.TxHash(hex.EncodeToString(recipientAddr[:8])), nil
}

// CompleteTransfer completes a native USDC transfer
func (s *SolanaCCTP) CompleteTransfer(ctx context.Context, message []byte, attestation []byte, signer types.Signer) (types.TxHash, error) {
	// Validate inputs
	if len(message) == 0 {
		return "", fmt.Errorf("message cannot be empty")
	}
	if len(attestation) == 0 {
		return "", fmt.Errorf("attestation cannot be empty")
	}

	// Validate signer chain
	if signer.Chain() != s.chain {
		return "", fmt.Errorf("signer chain %s does not match CCTP chain %s", signer.Chain(), s.chain)
	}

	// In a real implementation, this would:
	// 1. Create receiveMessage instruction for Message Transmitter
	// 2. Pass message and attestation as instruction data
	// 3. Build and sign transaction
	// 4. Send transaction to Solana network
	// 5. Message Transmitter program verifies attestation
	// 6. Program mints USDC to recipient
	// 7. Return transaction signature

	// Pseudo-code:
	// receiveIx := createReceiveMessageInstruction(message, attestation)
	// tx := buildTransaction([receiveIx])
	// signature := signer.Sign(tx)
	// txHash := sendTransaction(signature)

	// For now, return placeholder
	return types.TxHash(hex.EncodeToString(message[:8])), nil
}

// GetAttestation fetches the attestation for a message from Circle's API
func (s *SolanaCCTP) GetAttestation(ctx context.Context, messageHash []byte) ([]byte, error) {
	if len(messageHash) != 32 {
		return nil, fmt.Errorf("message hash must be 32 bytes, got %d", len(messageHash))
	}

	// Circle's attestation API endpoint
	// Testnet: https://iris-api-sandbox.circle.com
	// Mainnet: https://iris-api.circle.com
	apiURL := "https://iris-api-sandbox.circle.com/v1/attestations/" + hex.EncodeToString(messageHash)

	// Poll for attestation with timeout
	maxAttempts := 60
	pollInterval := 2 * time.Second

	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch attestation: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response: %w", err)
			}

			var attestResp AttestationResponse
			if err := json.Unmarshal(body, &attestResp); err != nil {
				return nil, fmt.Errorf("failed to parse attestation response: %w", err)
			}

			if attestResp.Status == "complete" && attestResp.Attestation != "" {
				// Decode hex attestation
				attestation, err := hex.DecodeString(attestResp.Attestation)
				if err != nil {
					return nil, fmt.Errorf("failed to decode attestation: %w", err)
				}
				return attestation, nil
			}
		}

		// Wait before next attempt
		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("attestation not available after %d attempts", maxAttempts)
}
