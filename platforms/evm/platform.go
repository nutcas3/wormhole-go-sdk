package evm

import (
	"fmt"

	wormhole "github.com/wormhole-foundation/wormhole-go-sdk"
	"github.com/wormhole-foundation/wormhole-go-sdk/types"
)

type Platform struct{}

func New() *Platform {
	return &Platform{}
}

func (p *Platform) Initialize(wh *wormhole.Wormhole) error {
	chains := []types.Chain{
		types.Ethereum,
		types.BSC,
		types.Polygon,
		types.Avalanche,
		types.Arbitrum,
		types.Optimism,
		types.Base,
		types.Sepolia,
	}

	for _, chain := range chains {
		if err := wh.AddChain(chain); err != nil {
			// Chain might not be in config, skip
			continue
		}
	}

	return nil
}

func (p *Platform) Name() types.Platform {
	return types.PlatformEVM
}

type Address struct {
	address string
}

func NewAddress(addr string) *Address {
	return &Address{address: addr}
}

func (a *Address) String() string {
	return a.address
}

func (a *Address) ToUniversalAddress() (types.UniversalAddress, error) {
	chainAddr := types.NewChainAddress(types.Ethereum, a.address)
	return chainAddr.ToUniversalAddress()
}

func (a *Address) Platform() types.Platform {
	return types.PlatformEVM
}

type Signer struct {
	chain      types.Chain
	address    string
	privateKey string
}

// NewSigner creates a new EVM signer
func NewSigner(chain types.Chain, privateKey string) (*Signer, error) {
	if chain.GetPlatform() != types.PlatformEVM {
		return nil, fmt.Errorf("chain %s is not an EVM chain", chain)
	}

	// In a real implementation, derive address from private key
	return &Signer{
		chain:      chain,
		privateKey: privateKey,
		address:    "", // Would be derived from private key
	}, nil
}

// Chain returns the chain
func (s *Signer) Chain() types.Chain {
	return s.chain
}

// Address returns the address
func (s *Signer) Address() string {
	return s.address
}
