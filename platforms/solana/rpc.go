package solana

import (
	"context"
	"fmt"

	solanago "github.com/gagliardetto/solana-go"
	solanarpc "github.com/gagliardetto/solana-go/rpc"
	"github.com/wormhole-foundation/wormhole-go-sdk/rpc"
	"github.com/wormhole-foundation/wormhole-go-sdk/types"
)

type RPCClient struct {
	client *solanarpc.Client
	config rpc.Config
	chain  types.Chain
}

func NewRPCClient(chain types.Chain, config rpc.Config) (*RPCClient, error) {
	client := solanarpc.New(config.URL)

	return &RPCClient{
		client: client,
		config: config,
		chain:  chain,
	}, nil
}

func (c *RPCClient) GetBlockHeight(ctx context.Context) (uint64, error) {
	slot, err := c.client.GetSlot(ctx, solanarpc.CommitmentFinalized)
	if err != nil {
		return 0, fmt.Errorf("failed to get slot: %w", err)
	}
	return slot, nil
}

func (c *RPCClient) GetBalance(ctx context.Context, address string) (string, error) {
	pubKey, err := solanago.PublicKeyFromBase58(address)
	if err != nil {
		return "", fmt.Errorf("invalid address: %w", err)
	}

	balance, err := c.client.GetBalance(ctx, pubKey, solanarpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("failed to get balance: %w", err)
	}

	return fmt.Sprintf("%d", balance.Value), nil
}

func (c *RPCClient) GetAccountInfo(ctx context.Context, address string) (*solanarpc.GetAccountInfoResult, error) {
	pubKey, err := solanago.PublicKeyFromBase58(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	accountInfo, err := c.client.GetAccountInfo(ctx, pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}

	return accountInfo, nil
}

func (c *RPCClient) GetTransaction(ctx context.Context, signature string) (*solanarpc.GetTransactionResult, error) {
	sig, err := solanago.SignatureFromBase58(signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	tx, err := c.client.GetTransaction(ctx, sig, &solanarpc.GetTransactionOpts{
		Commitment: solanarpc.CommitmentFinalized,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return tx, nil
}

func (c *RPCClient) GetRecentBlockhash(ctx context.Context) (solanago.Hash, error) {
	recent, err := c.client.GetRecentBlockhash(ctx, solanarpc.CommitmentFinalized)
	if err != nil {
		return solanago.Hash{}, fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	return recent.Value.Blockhash, nil
}

func (c *RPCClient) GetLatestBlockhash(ctx context.Context) (*solanarpc.GetLatestBlockhashResult, error) {
	blockhash, err := c.client.GetLatestBlockhash(ctx, solanarpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest blockhash: %w", err)
	}

	return blockhash, nil
}

func (c *RPCClient) SendTransaction(ctx context.Context, tx *solanago.Transaction) (solanago.Signature, error) {
	sig, err := c.client.SendTransactionWithOpts(ctx, tx, solanarpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: solanarpc.CommitmentFinalized,
	})
	if err != nil {
		return solanago.Signature{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return sig, nil
}

func (c *RPCClient) GetMinimumBalanceForRentExemption(ctx context.Context, dataSize uint64) (uint64, error) {
	balance, err := c.client.GetMinimumBalanceForRentExemption(ctx, dataSize, solanarpc.CommitmentFinalized)
	if err != nil {
		return 0, fmt.Errorf("failed to get minimum balance: %w", err)
	}

	return balance, nil
}

func (c *RPCClient) GetTokenAccountBalance(ctx context.Context, tokenAccount string) (*solanarpc.GetTokenAccountBalanceResult, error) {
	pubKey, err := solanago.PublicKeyFromBase58(tokenAccount)
	if err != nil {
		return nil, fmt.Errorf("invalid token account: %w", err)
	}

	balance, err := c.client.GetTokenAccountBalance(ctx, pubKey, solanarpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("failed to get token account balance: %w", err)
	}

	return balance, nil
}

func (c *RPCClient) IsHealthy(ctx context.Context) bool {
	_, err := c.client.GetHealth(ctx)
	return err == nil
}

func (c *RPCClient) Close() error {
	return nil
}

func (c *RPCClient) Client() *solanarpc.Client {
	return c.client
}
