package types

type Chain string

const (
	Ethereum  Chain = "Ethereum"
	BSC       Chain = "Bsc"
	Polygon   Chain = "Polygon"
	Avalanche Chain = "Avalanche"
	Fantom    Chain = "Fantom"
	Celo      Chain = "Celo"
	Moonbeam  Chain = "Moonbeam"
	Arbitrum  Chain = "Arbitrum"
	Optimism  Chain = "Optimism"
	Base      Chain = "Base"
	Sepolia   Chain = "Sepolia"

	Solana    Chain = "Solana"
	Terra     Chain = "Terra"
	Terra2    Chain = "Terra2"
	Algorand  Chain = "Algorand"
	Aptos     Chain = "Aptos"
	Sui       Chain = "Sui"
	Injective Chain = "Injective"
	Osmosis   Chain = "Osmosis"
	Cosmoshub Chain = "Cosmoshub"
)

type Platform string

const (
	PlatformEVM      Platform = "Evm"
	PlatformSolana   Platform = "Solana"
	PlatformCosmWasm Platform = "Cosmwasm"
	PlatformAlgorand Platform = "Algorand"
	PlatformAptos    Platform = "Aptos"
	PlatformSui      Platform = "Sui"
)

type ChainID uint16

const (
	ChainIDUnset     ChainID = 0
	ChainIDSolana    ChainID = 1
	ChainIDEthereum  ChainID = 2
	ChainIDTerra     ChainID = 3
	ChainIDBSC       ChainID = 4
	ChainIDPolygon   ChainID = 5
	ChainIDAvalanche ChainID = 6
	ChainIDFantom    ChainID = 10
	ChainIDAlgorand  ChainID = 8
	ChainIDAptos     ChainID = 22
	ChainIDSui       ChainID = 21
	ChainIDArbitrum  ChainID = 23
	ChainIDOptimism  ChainID = 24
	ChainIDBase      ChainID = 30
	ChainIDSepolia   ChainID = 10002
)

func (c Chain) GetPlatform() Platform {
	switch c {
	case Ethereum, BSC, Polygon, Avalanche, Fantom, Celo, Moonbeam, Arbitrum, Optimism, Base, Sepolia:
		return PlatformEVM
	case Solana:
		return PlatformSolana
	case Terra, Terra2, Injective, Osmosis, Cosmoshub:
		return PlatformCosmWasm
	case Algorand:
		return PlatformAlgorand
	case Aptos:
		return PlatformAptos
	case Sui:
		return PlatformSui
	default:
		return ""
	}
}

func (c Chain) GetChainID() ChainID {
	switch c {
	case Solana:
		return ChainIDSolana
	case Ethereum:
		return ChainIDEthereum
	case Terra:
		return ChainIDTerra
	case BSC:
		return ChainIDBSC
	case Polygon:
		return ChainIDPolygon
	case Avalanche:
		return ChainIDAvalanche
	case Fantom:
		return ChainIDFantom
	case Algorand:
		return ChainIDAlgorand
	case Aptos:
		return ChainIDAptos
	case Sui:
		return ChainIDSui
	case Arbitrum:
		return ChainIDArbitrum
	case Optimism:
		return ChainIDOptimism
	case Base:
		return ChainIDBase
	case Sepolia:
		return ChainIDSepolia
	default:
		return ChainIDUnset
	}
}
