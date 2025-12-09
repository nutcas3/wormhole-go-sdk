package solana

import (
	wormhole "github.com/wormhole-foundation/wormhole-go-sdk"
	"github.com/wormhole-foundation/wormhole-go-sdk/types"
)

type Platform struct{}

func New() *Platform {
	return &Platform{}
}

func (p *Platform) Initialize(wh *wormhole.Wormhole) error {
	if err := wh.AddChain(types.Solana); err != nil {
		return err
	}
	return nil
}

func (p *Platform) Name() types.Platform {
	return types.PlatformSolana
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
	chainAddr := types.NewChainAddress(types.Solana, a.address)
	return chainAddr.ToUniversalAddress()
}

func (a *Address) Platform() types.Platform {
	return types.PlatformSolana
}

type Signer struct {
	address    string
	privateKey []byte
}

func NewSigner(privateKey []byte) *Signer {
	return &Signer{
		privateKey: privateKey,
		address:    "", // Would be derived from private key
	}
}

func (s *Signer) Chain() types.Chain {
	return types.Solana
}

func (s *Signer) Address() string {
	return s.address
}
