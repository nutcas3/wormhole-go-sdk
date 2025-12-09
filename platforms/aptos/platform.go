package aptos

import (
	wormhole "github.com/wormhole-foundation/wormhole-go-sdk"
	"github.com/wormhole-foundation/wormhole-go-sdk/types"
)

type Platform struct{}

func New() *Platform {
	return &Platform{}
}

func (p *Platform) Initialize(wh *wormhole.Wormhole) error {
	if err := wh.AddChain(types.Aptos); err != nil {
		return err
	}
	return nil
}

func (p *Platform) Name() types.Platform {
	return types.PlatformAptos
}
