package aptos

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

type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      int             `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type AccountResource struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type TransactionInfo struct {
	Hash                string `json:"hash"`
	Version             string `json:"version"`
	Success             bool   `json:"success"`
	VMStatus            string `json:"vm_status"`
	GasUsed             string `json:"gas_used"`
	SequenceNumber      string `json:"sequence_number"`
	Sender              string `json:"sender"`
	MaxGasAmount        string `json:"max_gas_amount"`
	GasUnitPrice        string `json:"gas_unit_price"`
	ExpirationTimestamp string `json:"expiration_timestamp_secs"`
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
	var result struct {
		BlockHeight string `json:"block_height"`
	}

	err := c.call(ctx, "GET", "/", nil, &result)
	if err != nil {
		return 0, fmt.Errorf("failed to get block height: %w", err)
	}

	var height uint64
	fmt.Sscanf(result.BlockHeight, "%d", &height)
	return height, nil
}

func (c *RPCClient) GetBalance(ctx context.Context, address string) (string, error) {
	endpoint := fmt.Sprintf("/v1/accounts/%s/resource/0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>", address)

	var resource AccountResource
	err := c.call(ctx, "GET", endpoint, nil, &resource)
	if err != nil {
		return "", fmt.Errorf("failed to get balance: %w", err)
	}

	var coinData struct {
		Coin struct {
			Value string `json:"value"`
		} `json:"coin"`
	}

	if err := json.Unmarshal(resource.Data, &coinData); err != nil {
		return "", fmt.Errorf("failed to parse balance: %w", err)
	}

	return coinData.Coin.Value, nil
}

func (c *RPCClient) GetAccount(ctx context.Context, address string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/v1/accounts/%s", address)

	var account map[string]interface{}
	err := c.call(ctx, "GET", endpoint, nil, &account)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return account, nil
}

func (c *RPCClient) GetAccountResources(ctx context.Context, address string) ([]AccountResource, error) {
	endpoint := fmt.Sprintf("/v1/accounts/%s/resources", address)

	var resources []AccountResource
	err := c.call(ctx, "GET", endpoint, nil, &resources)
	if err != nil {
		return nil, fmt.Errorf("failed to get account resources: %w", err)
	}

	return resources, nil
}

func (c *RPCClient) GetTransaction(ctx context.Context, txHash string) (*TransactionInfo, error) {
	endpoint := fmt.Sprintf("/v1/transactions/by_hash/%s", txHash)

	var tx TransactionInfo
	err := c.call(ctx, "GET", endpoint, nil, &tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &tx, nil
}

func (c *RPCClient) SubmitTransaction(ctx context.Context, signedTx interface{}) (string, error) {
	var result struct {
		Hash string `json:"hash"`
	}

	err := c.call(ctx, "POST", "/v1/transactions", signedTx, &result)
	if err != nil {
		return "", fmt.Errorf("failed to submit transaction: %w", err)
	}

	return result.Hash, nil
}

func (c *RPCClient) SimulateTransaction(ctx context.Context, tx interface{}) ([]interface{}, error) {
	var result []interface{}

	err := c.call(ctx, "POST", "/v1/transactions/simulate", tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to simulate transaction: %w", err)
	}

	return result, nil
}

func (c *RPCClient) IsHealthy(ctx context.Context) bool {
	err := c.call(ctx, "GET", "/v1/", nil, nil)
	return err == nil
}

func (c *RPCClient) Close() error {
	return nil
}

func (c *RPCClient) call(ctx context.Context, method, endpoint string, body, result interface{}) error {
	url := c.config.URL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
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
