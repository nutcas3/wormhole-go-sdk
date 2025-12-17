package sui

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

type SuiObjectResponse struct {
	Data  *SuiObjectData `json:"data,omitempty"`
	Error *RPCError      `json:"error,omitempty"`
}

type SuiObjectData struct {
	ObjectID string                 `json:"objectId"`
	Version  string                 `json:"version"`
	Digest   string                 `json:"digest"`
	Type     string                 `json:"type,omitempty"`
	Owner    map[string]interface{} `json:"owner,omitempty"`
	Content  map[string]interface{} `json:"content,omitempty"`
}

type TransactionBlockResponse struct {
	Digest  string                 `json:"digest"`
	Effects map[string]interface{} `json:"effects,omitempty"`
	Events  []interface{}          `json:"events,omitempty"`
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
		SequenceNumber string `json:"sequenceNumber"`
	}

	err := c.callRPC(ctx, "sui_getLatestCheckpointSequenceNumber", []interface{}{}, &result)
	if err != nil {
		return 0, fmt.Errorf("failed to get block height: %w", err)
	}

	var height uint64
	fmt.Sscanf(result.SequenceNumber, "%d", &height)
	return height, nil
}

func (c *RPCClient) GetBalance(ctx context.Context, address string) (string, error) {
	var result struct {
		TotalBalance string `json:"totalBalance"`
	}

	params := []interface{}{address, "0x2::sui::SUI"}
	err := c.callRPC(ctx, "suix_getBalance", params, &result)
	if err != nil {
		return "", fmt.Errorf("failed to get balance: %w", err)
	}

	return result.TotalBalance, nil
}

func (c *RPCClient) GetObject(ctx context.Context, objectID string) (*SuiObjectResponse, error) {
	var result SuiObjectResponse

	params := []interface{}{
		objectID,
		map[string]interface{}{
			"showType":    true,
			"showOwner":   true,
			"showContent": true,
		},
	}

	err := c.callRPC(ctx, "sui_getObject", params, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return &result, nil
}

func (c *RPCClient) GetOwnedObjects(ctx context.Context, address string) ([]SuiObjectResponse, error) {
	var result struct {
		Data []SuiObjectResponse `json:"data"`
	}

	params := []interface{}{address}
	err := c.callRPC(ctx, "suix_getOwnedObjects", params, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get owned objects: %w", err)
	}

	return result.Data, nil
}

func (c *RPCClient) GetTransaction(ctx context.Context, digest string) (*TransactionBlockResponse, error) {
	var result TransactionBlockResponse

	params := []interface{}{
		digest,
		map[string]interface{}{
			"showInput":          true,
			"showEffects":        true,
			"showEvents":         true,
			"showObjectChanges":  true,
			"showBalanceChanges": true,
		},
	}

	err := c.callRPC(ctx, "sui_getTransactionBlock", params, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &result, nil
}

func (c *RPCClient) ExecuteTransaction(ctx context.Context, txBytes string, signatures []string) (*TransactionBlockResponse, error) {
	var result TransactionBlockResponse

	params := []interface{}{
		txBytes,
		signatures,
		map[string]interface{}{
			"showInput":          true,
			"showEffects":        true,
			"showEvents":         true,
			"showObjectChanges":  true,
			"showBalanceChanges": true,
		},
		"WaitForLocalExecution",
	}

	err := c.callRPC(ctx, "sui_executeTransactionBlock", params, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to execute transaction: %w", err)
	}

	return &result, nil
}

func (c *RPCClient) DryRunTransaction(ctx context.Context, txBytes string) (map[string]interface{}, error) {
	var result map[string]interface{}

	params := []interface{}{txBytes}
	err := c.callRPC(ctx, "sui_dryRunTransactionBlock", params, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to dry run transaction: %w", err)
	}

	return result, nil
}

func (c *RPCClient) GetReferenceGasPrice(ctx context.Context) (string, error) {
	var result string

	err := c.callRPC(ctx, "suix_getReferenceGasPrice", []interface{}{}, &result)
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	return result, nil
}

func (c *RPCClient) IsHealthy(ctx context.Context) bool {
	var result interface{}
	err := c.callRPC(ctx, "sui_getLatestCheckpointSequenceNumber", []interface{}{}, &result)
	return err == nil
}

func (c *RPCClient) Close() error {
	return nil
}

func (c *RPCClient) callRPC(ctx context.Context, method string, params []interface{}, result interface{}) error {
	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.URL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var rpcResp JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	if result != nil {
		if err := json.Unmarshal(rpcResp.Result, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}
