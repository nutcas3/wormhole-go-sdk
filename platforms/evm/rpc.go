package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/wormhole-foundation/wormhole-go-sdk/rpc"
	"github.com/wormhole-foundation/wormhole-go-sdk/types"
)

type RPCClient struct {
	client *ethclient.Client
	config rpc.Config
	chain  types.Chain
}

func NewRPCClient(chain types.Chain, config rpc.Config) (*RPCClient, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to EVM RPC: %w", err)
	}

	return &RPCClient{
		client: client,
		config: config,
		chain:  chain,
	}, nil
}

func (c *RPCClient) GetBlockHeight(ctx context.Context) (uint64, error) {
	header, err := c.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get block height: %w", err)
	}
	return header.Number.Uint64(), nil
}

func (c *RPCClient) GetBalance(ctx context.Context, address string) (string, error) {
	addr := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get balance: %w", err)
	}
	return balance.String(), nil
}

func (c *RPCClient) GetNonce(ctx context.Context, address string) (uint64, error) {
	addr := common.HexToAddress(address)
	nonce, err := c.client.PendingNonceAt(ctx, addr)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}
	return nonce, nil
}

func (c *RPCClient) GetChainID(ctx context.Context) (*big.Int, error) {
	chainID, err := c.client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}
	return chainID, nil
}

func (c *RPCClient) EstimateGas(ctx context.Context, from, to string, data []byte, value *big.Int) (uint64, error) {
	fromAddr := common.HexToAddress(from)
	toAddr := common.HexToAddress(to)

	msg := ethereum.CallMsg{
		From:  fromAddr,
		To:    &toAddr,
		Data:  data,
		Value: value,
	}

	gas, err := c.client.EstimateGas(ctx, msg)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}
	return gas, nil
}

func (c *RPCClient) GetTransactionReceipt(ctx context.Context, txHash string) (map[string]interface{}, error) {
	hash := common.HexToHash(txHash)
	receipt, err := c.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	return map[string]interface{}{
		"blockNumber":       receipt.BlockNumber.Uint64(),
		"transactionHash":   receipt.TxHash.Hex(),
		"status":            receipt.Status,
		"gasUsed":           receipt.GasUsed,
		"cumulativeGasUsed": receipt.CumulativeGasUsed,
		"contractAddress":   receipt.ContractAddress.Hex(),
		"logs":              receipt.Logs,
	}, nil
}

func (c *RPCClient) GetGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}
	return gasPrice, nil
}

func (c *RPCClient) GetCode(ctx context.Context, address string) ([]byte, error) {
	addr := common.HexToAddress(address)
	code, err := c.client.CodeAt(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get code: %w", err)
	}
	return code, nil
}

func (c *RPCClient) IsHealthy(ctx context.Context) bool {
	_, err := c.client.BlockNumber(ctx)
	return err == nil
}

func (c *RPCClient) Close() error {
	c.client.Close()
	return nil
}

func (c *RPCClient) Client() *ethclient.Client {
	return c.client
}
