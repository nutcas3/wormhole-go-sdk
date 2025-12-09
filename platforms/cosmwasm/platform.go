package cosmwasm

import (
	wormhole "github.com/wormhole-foundation/wormhole-go-sdk"
	"github.com/wormhole-foundation/wormhole-go-sdk/types"
)

type Platform struct{}

func New() *Platform {
	return &Platform{}
}

func (p *Platform) Initialize(wh *wormhole.Wormhole) error {
	chains := []types.Chain{
		types.Terra,
		types.Terra2,
		types.Injective,
		types.Osmosis,
		types.Cosmoshub,
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
	return types.PlatformCosmWasm
}
