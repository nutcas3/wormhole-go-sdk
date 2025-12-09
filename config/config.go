package config

import "github.com/wormhole-foundation/wormhole-go-sdk/types"

// ChainConfig holds configuration for a specific chain
type ChainConfig struct {
	Chain       types.Chain
	ChainID     types.ChainID
	Platform    types.Platform
	Network     types.Network
	RPC         string
	CoreBridge  string // Core bridge contract address
	TokenBridge string // Token bridge contract address
}

// NetworkConfig holds all chain configurations for a network
type NetworkConfig struct {
	Network types.Network
	Chains  map[types.Chain]ChainConfig
}

// GetDefaultConfig returns the default configuration for a network
func GetDefaultConfig(network types.Network) *NetworkConfig {
	config := &NetworkConfig{
		Network: network,
		Chains:  make(map[types.Chain]ChainConfig),
	}

	switch network {
	case types.Mainnet:
		config.Chains = getMainnetChains()
	case types.Testnet:
		config.Chains = getTestnetChains()
	case types.Devnet:
		config.Chains = getDevnetChains()
	}

	return config
}

func getMainnetChains() map[types.Chain]ChainConfig {
	return map[types.Chain]ChainConfig{
		types.Ethereum: {
			Chain:       types.Ethereum,
			ChainID:     types.ChainIDEthereum,
			Platform:    types.PlatformEVM,
			Network:     types.Mainnet,
			RPC:         "https://ethereum.publicnode.com",
			CoreBridge:  "0x98f3c9e6E3fAce36bAAd05FE09d375Ef1464288B",
			TokenBridge: "0x3ee18B2214AFF97000D974cf647E7C347E8fa585",
		},
		types.Solana: {
			Chain:       types.Solana,
			ChainID:     types.ChainIDSolana,
			Platform:    types.PlatformSolana,
			Network:     types.Mainnet,
			RPC:         "https://api.mainnet-beta.solana.com",
			CoreBridge:  "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth",
			TokenBridge: "wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb",
		},
		types.BSC: {
			Chain:       types.BSC,
			ChainID:     types.ChainIDBSC,
			Platform:    types.PlatformEVM,
			Network:     types.Mainnet,
			RPC:         "https://bsc-dataseed.binance.org",
			CoreBridge:  "0x98f3c9e6E3fAce36bAAd05FE09d375Ef1464288B",
			TokenBridge: "0xB6F6D86a8f9879A9c87f643768d9efc38c1Da6E7",
		},
	}
}

func getTestnetChains() map[types.Chain]ChainConfig {
	return map[types.Chain]ChainConfig{
		types.Sepolia: {
			Chain:       types.Sepolia,
			ChainID:     types.ChainIDSepolia,
			Platform:    types.PlatformEVM,
			Network:     types.Testnet,
			RPC:         "https://ethereum-sepolia.publicnode.com",
			CoreBridge:  "0x4a8bc80Ed5a4067f1CCf107057b8270E0cC11A78",
			TokenBridge: "0xDB5492265f6038831E89f495670FF909aDe94bd9",
		},
		types.Solana: {
			Chain:       types.Solana,
			ChainID:     types.ChainIDSolana,
			Platform:    types.PlatformSolana,
			Network:     types.Testnet,
			RPC:         "https://api.devnet.solana.com",
			CoreBridge:  "3u8hJUVTA4jH1wYAyUur7FFZVQ8H635K3tSHHF4ssjQ5",
			TokenBridge: "DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe",
		},
	}
}

func getDevnetChains() map[types.Chain]ChainConfig {
	return map[types.Chain]ChainConfig{
		types.Ethereum: {
			Chain:       types.Ethereum,
			ChainID:     types.ChainIDEthereum,
			Platform:    types.PlatformEVM,
			Network:     types.Devnet,
			RPC:         "http://localhost:8545",
			CoreBridge:  "0xC89Ce4735882C9F0f0FE26686c53074E09B0D550",
			TokenBridge: "0x0290FB167208Af455bB137780163b7B7a9a10C16",
		},
		types.Solana: {
			Chain:       types.Solana,
			ChainID:     types.ChainIDSolana,
			Platform:    types.PlatformSolana,
			Network:     types.Devnet,
			RPC:         "http://localhost:8899",
			CoreBridge:  "Bridge1p5gheXUvJ6jGWGeCsgPKgnE3YgdGKRVCMY9o",
			TokenBridge: "B6RHG3mfcckmrYN1UhmJzyS1XX3fZKbkeUcpJe9Sy3FE",
		},
	}
}
