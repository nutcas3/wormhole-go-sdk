package context

import (
	"fmt"

	"github.com/wormhole-foundation/wormhole-go-sdk/config"
	"github.com/wormhole-foundation/wormhole-go-sdk/types"
)

type ChainContext struct {
	config    config.ChainConfig
	rpcClient any // Platform-specific RPC client
	protocols map[string]interface{}
}

func NewChainContext(cfg config.ChainConfig) (*ChainContext, error) {
	ctx := &ChainContext{
		config:    cfg,
		protocols: make(map[string]interface{}),
	}

	// Initialize RPC client based on platform
	if err := ctx.initRPCClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize RPC client: %w", err)
	}

	return ctx, nil
}

func (c *ChainContext) Chain() types.Chain {
	return c.config.Chain
}

func (c *ChainContext) ChainID() types.ChainID {
	return c.config.ChainID
}

func (c *ChainContext) Platform() types.Platform {
	return c.config.Platform
}

func (c *ChainContext) Config() config.ChainConfig {
	return c.config
}

func (c *ChainContext) GetRPCClient() any {
	return c.rpcClient
}

func (c *ChainContext) initRPCClient() error {
	switch c.config.Platform {
	case types.PlatformEVM:
		// Initialize EVM RPC client
		// This will be implemented in the EVM platform package
		return nil
	case types.PlatformSolana:
		// Initialize Solana RPC client
		// This will be implemented in the Solana platform package
		return nil
	default:
		return fmt.Errorf("unsupported platform: %s", c.config.Platform)
	}
}

func (c *ChainContext) GetCoreBridge() (any, error) {
	if client, ok := c.protocols["core"]; ok {
		return client, nil
	}
	return nil, fmt.Errorf("core bridge not initialized for chain %s", c.config.Chain)
}

func (c *ChainContext) GetTokenBridge() (any, error) {
	if client, ok := c.protocols["tokenbridge"]; ok {
		return client, nil
	}
	return nil, fmt.Errorf("token bridge not initialized for chain %s", c.config.Chain)
}

func (c *ChainContext) SetProtocol(name string, client any) {
	c.protocols[name] = client
}
