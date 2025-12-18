package algorand

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wormhole-foundation/wormhole-go-sdk/rpc"
	"github.com/wormhole-foundation/wormhole-go-sdk/types"
)

type RPCClient struct {
	httpClient *http.Client
	config     rpc.Config
	chain      types.Chain
}

type AccountInfo struct {
	Address                     string `json:"address"`
	Amount                      uint64 `json:"amount"`
	AmountWithoutPendingRewards uint64 `json:"amount-without-pending-rewards"`
	PendingRewards              uint64 `json:"pending-rewards"`
	Round                       uint64 `json:"round"`
	Status                      string `json:"status"`
}

type NodeStatus struct {
	LastRound                 uint64 `json:"last-round"`
	LastVersion               string `json:"last-version"`
	NextVersion               string `json:"next-version"`
	NextVersionRound          uint64 `json:"next-version-round"`
	NextVersionSupported      bool   `json:"next-version-supported"`
	StoppedAtUnsupportedRound bool   `json:"stopped-at-unsupported-round"`
	TimeSinceLastRound        uint64 `json:"time-since-last-round"`
	CatchupTime               uint64 `json:"catchup-time"`
	HasSyncedSinceStartup     bool   `json:"has-synced-since-startup"`
}

type TransactionParams struct {
	ConsensusVersion string `json:"consensus-version"`
	Fee              uint64 `json:"fee"`
	GenesisHash      string `json:"genesis-hash"`
	GenesisID        string `json:"genesis-id"`
	LastRound        uint64 `json:"last-round"`
	MinFee           uint64 `json:"min-fee"`
}

type PendingTransaction struct {
	PoolError        string `json:"pool-error"`
	TxID             string `json:"txn"`
	ApplicationIndex uint64 `json:"application-index,omitempty"`
	AssetIndex       uint64 `json:"asset-index,omitempty"`
	ConfirmedRound   uint64 `json:"confirmed-round,omitempty"`
	GlobalStateDelta []byte `json:"global-state-delta,omitempty"`
	LocalStateDelta  []byte `json:"local-state-delta,omitempty"`
}

func NewRPCClient(chain types.Chain, config rpc.Config) (*RPCClient, error) {
	return &RPCClient{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
		chain:  chain,
	}, nil
}

func (c *RPCClient) GetBlockHeight(ctx context.Context) (uint64, error) {
	status, err := c.GetStatus(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get block height: %w", err)
	}
	return status.LastRound, nil
}

func (c *RPCClient) GetBalance(ctx context.Context, address string) (string, error) {
	account, err := c.GetAccountInfo(ctx, address)
	if err != nil {
		return "", fmt.Errorf("failed to get balance: %w", err)
	}
	return fmt.Sprintf("%d", account.Amount), nil
}

func (c *RPCClient) GetStatus(ctx context.Context) (*NodeStatus, error) {
	var status NodeStatus
	err := c.call(ctx, "GET", "/v2/status", nil, &status)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	return &status, nil
}

func (c *RPCClient) GetAccountInfo(ctx context.Context, address string) (*AccountInfo, error) {
	var account AccountInfo
	endpoint := fmt.Sprintf("/v2/accounts/%s", address)
	err := c.call(ctx, "GET", endpoint, nil, &account)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}
	return &account, nil
}

func (c *RPCClient) GetTransactionParams(ctx context.Context) (*TransactionParams, error) {
	var params TransactionParams
	err := c.call(ctx, "GET", "/v2/transactions/params", nil, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction params: %w", err)
	}
	return &params, nil
}

func (c *RPCClient) SendRawTransaction(ctx context.Context, txn []byte) (string, error) {
	var result struct {
		TxID string `json:"txId"`
	}

	err := c.call(ctx, "POST", "/v2/transactions", txn, &result)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return result.TxID, nil
}

func (c *RPCClient) GetPendingTransaction(ctx context.Context, txID string) (*PendingTransaction, error) {
	var tx PendingTransaction
	endpoint := fmt.Sprintf("/v2/transactions/pending/%s", txID)
	err := c.call(ctx, "GET", endpoint, nil, &tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending transaction: %w", err)
	}
	return &tx, nil
}

func (c *RPCClient) WaitForConfirmation(ctx context.Context, txID string, maxRounds uint64) (*PendingTransaction, error) {
	status, err := c.GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	startRound := status.LastRound
	for {
		tx, err := c.GetPendingTransaction(ctx, txID)
		if err != nil {
			return nil, err
		}

		if tx.ConfirmedRound > 0 {
			return tx, nil
		}

		status, err = c.GetStatus(ctx)
		if err != nil {
			return nil, err
		}

		if status.LastRound > startRound+maxRounds {
			return nil, fmt.Errorf("transaction not confirmed after %d rounds", maxRounds)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
}

func (c *RPCClient) IsHealthy(ctx context.Context) bool {
	_, err := c.GetStatus(ctx)
	return err == nil
}

func (c *RPCClient) Close() error {
	return nil
}

func (c *RPCClient) call(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	url := c.config.URL + endpoint

	var reqBody io.Reader
	if body != nil {
		switch v := body.(type) {
		case []byte:
			reqBody = bytes.NewReader(v)
		default:
			jsonData, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}
			reqBody = bytes.NewReader(jsonData)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
