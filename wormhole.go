package wormhole

import (
	"fmt"

	"github.com/wormhole-foundation/wormhole-go-sdk/config"
	"github.com/wormhole-foundation/wormhole-go-sdk/context"
	"github.com/wormhole-foundation/wormhole-go-sdk/types"
)

// Wormhole is the main SDK instance
type Wormhole struct {
	network types.Network
	config  *config.NetworkConfig
	chains  map[types.Chain]*context.ChainContext
}

// New creates a new Wormhole SDK instance
func New(network types.Network, platforms ...Platform) (*Wormhole, error) {
	if !network.IsValid() {
		return nil, fmt.Errorf("invalid network: %s", network)
	}

	wh := &Wormhole{
		network: network,
		config:  config.GetDefaultConfig(network),
		chains:  make(map[types.Chain]*context.ChainContext),
	}

	// Initialize platforms
	for _, platform := range platforms {
		if err := platform.Initialize(wh); err != nil {
			return nil, fmt.Errorf("failed to initialize platform: %w", err)
		}
	}

	return wh, nil
}

// Network returns the current network
func (w *Wormhole) Network() types.Network {
	return w.network
}

// GetChain returns the chain context for a specific chain
func (w *Wormhole) GetChain(chain types.Chain) (*context.ChainContext, error) {
	ctx, ok := w.chains[chain]
	if !ok {
		return nil, fmt.Errorf("chain %s not initialized", chain)
	}
	return ctx, nil
}

// AddChain adds a chain context to the Wormhole instance
func (w *Wormhole) AddChain(chain types.Chain) error {
	cfg, ok := w.config.Chains[chain]
	if !ok {
		return fmt.Errorf("chain %s not found in config", chain)
	}

	ctx, err := context.NewChainContext(cfg)
	if err != nil {
		return fmt.Errorf("failed to create chain context: %w", err)
	}

	w.chains[chain] = ctx
	return nil
}

// GetChains returns all initialized chains
func (w *Wormhole) GetChains() []types.Chain {
	chains := make([]types.Chain, 0, len(w.chains))
	for chain := range w.chains {
		chains = append(chains, chain)
	}
	return chains
}

// ChainAddress creates a new ChainAddress
func ChainAddress(chain types.Chain, address string) types.ChainAddress {
	return types.NewChainAddress(chain, address)
}

// TokenID creates a new TokenID
func TokenID(chain types.Chain, address string) types.TokenID {
	return types.NewTokenID(chain, address)
}

// CanonicalAddress converts a ChainAddress to its canonical string format
func CanonicalAddress(addr types.ChainAddress) string {
	return addr.Address
}

// Platform is an interface that platforms must implement
type Platform interface {
	Initialize(wh *Wormhole) error
	Name() types.Platform
}
