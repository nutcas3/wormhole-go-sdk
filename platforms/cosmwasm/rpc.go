package cosmwasm

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
	Data    string `json:"data,omitempty"`
}

type BlockInfo struct {
	Height string `json:"height"`
	Time   string `json:"time"`
	Hash   string `json:"hash"`
}

type AccountInfo struct {
	Address       string `json:"address"`
	AccountNumber string `json:"account_number"`
	Sequence      string `json:"sequence"`
	Coins         []Coin `json:"coins"`
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type TxResponse struct {
	Height    string                 `json:"height"`
	TxHash    string                 `json:"txhash"`
	Code      uint32                 `json:"code"`
	RawLog    string                 `json:"raw_log"`
	Logs      []interface{}          `json:"logs,omitempty"`
	GasWanted string                 `json:"gas_wanted"`
	GasUsed   string                 `json:"gas_used"`
	Tx        map[string]interface{} `json:"tx,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

type SmartQueryResponse struct {
	Data json.RawMessage `json:"data"`
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
		Block struct {
			Header struct {
				Height string `json:"height"`
			} `json:"header"`
		} `json:"block"`
	}

	err := c.callRPC(ctx, "block", []interface{}{}, &result)
	if err != nil {
		return 0, fmt.Errorf("failed to get block height: %w", err)
	}

	var height uint64
	fmt.Sscanf(result.Block.Header.Height, "%d", &height)
	return height, nil
}

func (c *RPCClient) GetBalance(ctx context.Context, address string) (string, error) {
	endpoint := fmt.Sprintf("/cosmos/bank/v1beta1/balances/%s", address)

	var result struct {
		Balances []Coin `json:"balances"`
	}

	err := c.callREST(ctx, "GET", endpoint, nil, &result)
	if err != nil {
		return "", fmt.Errorf("failed to get balance: %w", err)
	}

	if len(result.Balances) == 0 {
		return "0", nil
	}

	return result.Balances[0].Amount, nil
}

func (c *RPCClient) GetAccount(ctx context.Context, address string) (*AccountInfo, error) {
	endpoint := fmt.Sprintf("/cosmos/auth/v1beta1/accounts/%s", address)

	var result struct {
		Account AccountInfo `json:"account"`
	}

	err := c.callREST(ctx, "GET", endpoint, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &result.Account, nil
}

func (c *RPCClient) GetTransaction(ctx context.Context, txHash string) (*TxResponse, error) {
	endpoint := fmt.Sprintf("/cosmos/tx/v1beta1/txs/%s", txHash)

	var result struct {
		TxResponse TxResponse `json:"tx_response"`
	}

	err := c.callREST(ctx, "GET", endpoint, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &result.TxResponse, nil
}

func (c *RPCClient) BroadcastTx(ctx context.Context, txBytes []byte, mode string) (*TxResponse, error) {
	endpoint := "/cosmos/tx/v1beta1/txs"

	body := map[string]interface{}{
		"tx_bytes": txBytes,
		"mode":     mode, // BROADCAST_MODE_SYNC, BROADCAST_MODE_ASYNC, BROADCAST_MODE_BLOCK
	}

	var result struct {
		TxResponse TxResponse `json:"tx_response"`
	}

	err := c.callREST(ctx, "POST", endpoint, body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	return &result.TxResponse, nil
}

func (c *RPCClient) QuerySmartContract(ctx context.Context, contractAddr string, queryMsg interface{}) (json.RawMessage, error) {
	queryBytes, err := json.Marshal(queryMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	endpoint := fmt.Sprintf("/cosmwasm/wasm/v1/contract/%s/smart/%s", contractAddr, string(queryBytes))

	var result SmartQueryResponse
	err = c.callREST(ctx, "GET", endpoint, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart contract: %w", err)
	}

	return result.Data, nil
}

func (c *RPCClient) GetContractInfo(ctx context.Context, contractAddr string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/cosmwasm/wasm/v1/contract/%s", contractAddr)

	var result map[string]interface{}
	err := c.callREST(ctx, "GET", endpoint, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract info: %w", err)
	}

	return result, nil
}

func (c *RPCClient) GetContractState(ctx context.Context, contractAddr string, key []byte) ([]byte, error) {
	endpoint := fmt.Sprintf("/cosmwasm/wasm/v1/contract/%s/raw/%x", contractAddr, key)

	var result struct {
		Data string `json:"data"`
	}

	err := c.callREST(ctx, "GET", endpoint, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract state: %w", err)
	}

	return []byte(result.Data), nil
}

func (c *RPCClient) SimulateTx(ctx context.Context, txBytes []byte) (uint64, error) {
	endpoint := "/cosmos/tx/v1beta1/simulate"

	body := map[string]interface{}{
		"tx_bytes": txBytes,
	}

	var result struct {
		GasInfo struct {
			GasUsed string `json:"gas_used"`
		} `json:"gas_info"`
	}

	err := c.callREST(ctx, "POST", endpoint, body, &result)
	if err != nil {
		return 0, fmt.Errorf("failed to simulate transaction: %w", err)
	}

	var gasUsed uint64
	fmt.Sscanf(result.GasInfo.GasUsed, "%d", &gasUsed)
	return gasUsed, nil
}

func (c *RPCClient) IsHealthy(ctx context.Context) bool {
	var result interface{}
	err := c.callRPC(ctx, "health", []interface{}{}, &result)
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

func (c *RPCClient) callREST(ctx context.Context, method, endpoint string, body, result interface{}) error {
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
