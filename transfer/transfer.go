package transfer

import (
"context"
"crypto/sha256"
"encoding/hex"
"fmt"
"math/big"
"time"

"github.com/ethereum/go-ethereum/common"
// "github.com/ethereum/go-ethereum/core/types"
"github.com/ethereum/go-ethereum/ethclient"
"github.com/gagliardetto/solana-go"
"github.com/gagliardetto/solana-go/rpc"
"github.com/wormhole-foundation/wormhole-go-sdk/protocols/core"
"github.com/wormhole-foundation/wormhole-go-sdk/protocols/tokenbridge"
wormholetypes "github.com/wormhole-foundation/wormhole-go-sdk/types"
)

// TransferState represents the state of a transfer
type TransferState string

const (
StateCreated   TransferState = "Created"
StateInitiated TransferState = "Initiated"
StateAttested  TransferState = "Attested"
StateCompleted TransferState = "Completed"
StateFailed    TransferState = "Failed"
)

// WormholeTransfer represents a cross-chain transfer
type WormholeTransfer struct {
	ID           string
	Token        wormholetypes.TokenID
	Amount       wormholetypes.Amount
	Source       wormholetypes.ChainAddress
	Destination  wormholetypes.ChainAddress
	Automatic    bool
	Payload      []byte
	NativeGas    string
	State        TransferState
	SourceTxHash wormholetypes.TxHash
	DestTxHash   wormholetypes.TxHash
	VAA          []byte
	Sequence     uint64
	CreatedAt    time.Time
	InitiatedAt  *time.Time
	AttestedAt   *time.Time
	CompletedAt  *time.Time
	Error        string
}

// TransferQuote represents a quote for a transfer
type TransferQuote struct {
	SourceToken      TokenAmount
	DestinationToken TokenAmount
	RelayFee         string
	GasFee           string
	TotalFee         string
	EstimatedTime    time.Duration
	ExchangeRate     string
}

// TokenAmount represents a token amount with metadata
type TokenAmount struct {
	Token    wormholetypes.TokenID
	Amount   wormholetypes.Amount
	USDValue string
	Decimals uint8
}

// Wormhole interface for dependency injection
type Wormhole interface {
	GetTokenBridge(chain wormholetypes.Chain) (tokenbridge.TokenBridge, error)
	GetCoreBridge(chain wormholetypes.Chain) (core.CoreBridge, error)
	GetGuardianRPC() string
	GetEVMClient(chain wormholetypes.Chain) (*ethclient.Client, error)
	GetSolanaClient(chain wormholetypes.Chain) (*rpc.Client, error)
}

// TokenTransfer provides high-level token transfer functionality
type TokenTransfer struct {
	wormhole Wormhole
	transfer *WormholeTransfer
}

// NewTokenTransfer creates a new token transfer
func NewTokenTransfer(
wormhole Wormhole,
token wormholetypes.TokenID,
amount wormholetypes.Amount,
source wormholetypes.ChainAddress,
destination wormholetypes.ChainAddress,
automatic bool,
payload []byte,
nativeGas string,
) *TokenTransfer {
	return &TokenTransfer{
		wormhole: wormhole,
		transfer: &WormholeTransfer{
			ID:          generateTransferID(),
			Token:       token,
			Amount:      amount,
			Source:      source,
			Destination: destination,
			Automatic:   automatic,
			Payload:     payload,
			NativeGas:   nativeGas,
			State:       StateCreated,
			CreatedAt:   time.Now(),
		},
	}
}

// QuoteTransfer returns a quote for the transfer
func (t *TokenTransfer) QuoteTransfer(ctx context.Context) (*TransferQuote, error) {
	destBridge, err := t.wormhole.GetTokenBridge(t.transfer.Destination.Chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination bridge: %w", err)
	}

	destToken := t.transfer.Token
	isWrapped, err := destBridge.IsWrappedAsset(ctx, t.transfer.Token)
	if err == nil && !isWrapped && t.transfer.Token.Chain != t.transfer.Destination.Chain {
		wrappedAddr, err := destBridge.GetWrappedAsset(ctx, t.transfer.Token)
		if err == nil && wrappedAddr != "" {
			destToken = wormholetypes.TokenID{
				Chain:   t.transfer.Destination.Chain,
				Address: wrappedAddr,
			}
		}
	}

	coreBridge, err := t.wormhole.GetCoreBridge(t.transfer.Source.Chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get core bridge: %w", err)
	}

	messageFee, err := coreBridge.GetMessageFee(ctx)
	if err != nil {
		messageFee = "0"
	}

	relayFee := "0"
	if t.transfer.Automatic {
		relayFee = estimateRelayFee(t.transfer.Destination.Chain, t.transfer.NativeGas)
	}

	totalFee := addFees(messageFee, relayFee)
	estimatedTime := estimateTransferTime(t.transfer.Source.Chain, t.transfer.Destination.Chain)

	return &TransferQuote{
		SourceToken: TokenAmount{
			Token:    t.transfer.Token,
			Amount:   t.transfer.Amount,
			Decimals: 18,
		},
		DestinationToken: TokenAmount{
			Token:    destToken,
			Amount:   t.transfer.Amount,
			Decimals: 18,
		},
		RelayFee:      relayFee,
		GasFee:        messageFee,
		TotalFee:      totalFee,
		EstimatedTime: estimatedTime,
		ExchangeRate:  "1.0",
	}, nil
}

// InitiateTransfer initiates the transfer on the source chain
func (t *TokenTransfer) InitiateTransfer(ctx context.Context, signer wormholetypes.Signer) ([]wormholetypes.TxHash, error) {
	if t.transfer.State != StateCreated {
		return nil, fmt.Errorf("transfer already initiated, current state: %s", t.transfer.State)
	}

	if signer.Chain() != t.transfer.Source.Chain {
		return nil, fmt.Errorf("signer chain %s does not match source chain %s", signer.Chain(), t.transfer.Source.Chain)
	}

	bridge, err := t.wormhole.GetTokenBridge(t.transfer.Source.Chain)
	if err != nil {
		t.transfer.State = StateFailed
		t.transfer.Error = fmt.Sprintf("failed to get token bridge: %v", err)
		return nil, err
	}

	coreBridge, err := t.wormhole.GetCoreBridge(t.transfer.Source.Chain)
	if err != nil {
		t.transfer.State = StateFailed
		t.transfer.Error = fmt.Sprintf("failed to get core bridge: %v", err)
		return nil, err
	}

	var txHashes []wormholetypes.TxHash

	transferTx, err := bridge.Transfer(
ctx,
t.transfer.Token,
t.transfer.Amount,
t.transfer.Destination,
signer,
)
	if err != nil {
		t.transfer.State = StateFailed
		t.transfer.Error = fmt.Sprintf("failed to transfer: %v", err)
		return nil, err
	}

	txHashes = append(txHashes, transferTx)

	sequence, err := t.getSequenceFromTx(ctx, coreBridge, transferTx)
	if err != nil {
		sequence = 0
	}

	now := time.Now()
	t.transfer.State = StateInitiated
	t.transfer.InitiatedAt = &now
	t.transfer.SourceTxHash = transferTx
	t.transfer.Sequence = sequence

	return txHashes, nil
}

// GetTransfer returns the transfer details
func (t *TokenTransfer) GetTransfer() *WormholeTransfer {
	return t.transfer
}

func (t *TokenTransfer) getSequenceFromTx(ctx context.Context, coreBridge core.CoreBridge, txHash wormholetypes.TxHash) (uint64, error) {
	switch coreBridge.(type) {
	case *core.EVMCoreBridge:
		return t.getSequenceFromEVMTx(ctx, coreBridge, txHash)
	case *core.SolanaCoreBridge:
		return t.getSequenceFromSolanaTx(ctx, coreBridge, txHash)
	default:
		return 0, fmt.Errorf("unsupported chain type")
	}
}

func (t *TokenTransfer) getSequenceFromEVMTx(ctx context.Context, coreBridge core.CoreBridge, txHash wormholetypes.TxHash) (uint64, error) {
	client, err := t.wormhole.GetEVMClient(t.transfer.Source.Chain)
	if err != nil {
		return 0, err
	}

	hash := common.HexToHash(string(txHash))
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-timeout:
			return 0, fmt.Errorf("timeout waiting for transaction receipt")
		case <-ticker.C:
			receipt, err := client.TransactionReceipt(ctx, hash)
			if err != nil {
				continue
			}
			msg, err := coreBridge.ParseMessageFromLogs(receipt.Logs)
			if err != nil {
				continue
			}
			return msg.Sequence, nil
		}
	}
}

func (t *TokenTransfer) getSequenceFromSolanaTx(ctx context.Context, coreBridge core.CoreBridge, txHash wormholetypes.TxHash) (uint64, error) {
	client, err := t.wormhole.GetSolanaClient(t.transfer.Source.Chain)
	if err != nil {
		return 0, err
	}

	sig, err := solana.SignatureFromBase58(string(txHash))
	if err != nil {
		return 0, err
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-timeout:
			return 0, fmt.Errorf("timeout waiting for transaction")
		case <-ticker.C:
			tx, err := client.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
				Commitment: rpc.CommitmentConfirmed,
			})
			if err != nil {
				continue
			}
			if tx == nil || tx.Meta == nil || tx.Meta.LogMessages == nil {
				continue
			}
			msg, err := coreBridge.ParseMessageFromLogs(tx.Meta.LogMessages)
			if err != nil {
				continue
			}
			return msg.Sequence, nil
		}
	}
}

func generateTransferID() string {
	timestamp := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", timestamp)))
	return fmt.Sprintf("transfer-%s", hex.EncodeToString(hash[:8]))
}

func estimateRelayFee(chain wormholetypes.Chain, nativeGas string) string {
	baseFee := map[wormholetypes.Chain]string{
		wormholetypes.Ethereum:  "50000000000000000",
		wormholetypes.Polygon:   "10000000000000000",
		wormholetypes.BSC:       "10000000000000000",
		wormholetypes.Solana:    "5000000",
		wormholetypes.Avalanche: "10000000000000000",
		wormholetypes.Arbitrum:  "5000000000000000",
		wormholetypes.Optimism:  "5000000000000000",
	}

	if fee, ok := baseFee[chain]; ok {
		if nativeGas != "" && nativeGas != "0" {
			gasAmount := new(big.Int)
			gasAmount.SetString(nativeGas, 10)
			baseFeeAmount := new(big.Int)
			baseFeeAmount.SetString(fee, 10)
			total := new(big.Int).Add(baseFeeAmount, gasAmount)
			return total.String()
		}
		return fee
	}
	return "10000000000000000"
}

func addFees(fee1, fee2 string) string {
	f1 := new(big.Int)
	f1.SetString(fee1, 10)
	f2 := new(big.Int)
	f2.SetString(fee2, 10)
	total := new(big.Int).Add(f1, f2)
	return total.String()
}

func estimateTransferTime(source, dest wormholetypes.Chain) time.Duration {
	baseTime := 15 * time.Minute
	finalityTime := map[wormholetypes.Chain]time.Duration{
		wormholetypes.Ethereum:  15 * time.Minute,
		wormholetypes.Polygon:   5 * time.Minute,
		wormholetypes.BSC:       3 * time.Minute,
		wormholetypes.Solana:    1 * time.Minute,
		wormholetypes.Avalanche: 2 * time.Minute,
		wormholetypes.Arbitrum:  2 * time.Minute,
		wormholetypes.Optimism:  2 * time.Minute,
	}
	sourceTime := finalityTime[wormholetypes.Ethereum]
	if t, ok := finalityTime[source]; ok {
		sourceTime = t
	}
	_ = dest
	return baseTime + sourceTime
}
