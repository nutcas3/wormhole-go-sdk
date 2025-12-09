// Package wormhole provides a comprehensive Go SDK for interacting with the Wormhole
// cross-chain messaging protocol.
//
// The Wormhole SDK enables developers to build cross-chain applications by providing
// a unified interface for:
//   - Cross-chain messaging
//   - Token transfers between different blockchains
//   - Integration with multiple blockchain platforms (EVM, Solana, Aptos, Sui, etc.)
//   - Protocol-specific functionality (Core Bridge, Token Bridge, CCTP)
//
// # Quick Start
//
// Initialize the SDK with your desired network and platforms:
//
//	import (
//	    wormhole "github.com/wormhole-foundation/wormhole-go-sdk"
//	    "github.com/wormhole-foundation/wormhole-go-sdk/platforms/evm"
//	    "github.com/wormhole-foundation/wormhole-go-sdk/platforms/solana"
//	    "github.com/wormhole-foundation/wormhole-go-sdk/types"
//	)
//
//	wh, err := wormhole.New(
//	    types.Testnet,
//	    evm.New(),
//	    solana.New(),
//	)
//
// # Chain Information
//
// Get information about supported chains:
//
//	chain, err := wh.GetChain(types.Ethereum)
//	fmt.Printf("Chain ID: %d\n", chain.ChainID())
//	fmt.Printf("RPC: %s\n", chain.Config().RPC)
//
// # Addresses and Tokens
//
// Create chain-specific addresses and token identifiers:
//
//	// Create addresses
//	sender := wormhole.ChainAddress(types.Ethereum, "0xYourAddress")
//	receiver := wormhole.ChainAddress(types.Solana, "YourSolanaAddress")
//
//	// Create token IDs
//	usdc := wormhole.TokenID(types.Ethereum, "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
//	nativeETH := wormhole.TokenID(types.Ethereum, "native")
//
// # Token Transfers
//
// Perform cross-chain token transfers:
//
//	import "github.com/wormhole-foundation/wormhole-go-sdk/transfer"
//
//	xfer := transfer.NewTokenTransfer(
//	    wh,
//	    token,
//	    amount,
//	    source,
//	    destination,
//	    false, // manual transfer
//	    nil,
//	    "",
//	)
//
//	// Get quote
//	quote, err := xfer.QuoteTransfer(ctx)
//
//	// Initiate transfer
//	txHashes, err := xfer.InitiateTransfer(ctx, signer)
//
//	// Wait for attestation
//	attestIds, err := xfer.FetchAttestation(ctx, timeout)
//
//	// Complete transfer
//	destTxHashes, err := xfer.CompleteTransfer(ctx, destSigner)
//
// # Supported Platforms
//
// The SDK supports multiple blockchain platforms:
//   - EVM: Ethereum, BSC, Polygon, Avalanche, Arbitrum, Optimism, Base, etc.
//   - Solana: Native Solana integration
//   - Aptos: Aptos Move-based chain
//   - Sui: Sui Move-based chain
//   - Algorand: Algorand blockchain
//   - CosmWasm: Terra, Injective, Osmosis, Cosmos Hub
//
// # Protocols
//
// Access protocol-specific functionality:
//
//	import (
//	    "github.com/wormhole-foundation/wormhole-go-sdk/protocols/core"
//	    "github.com/wormhole-foundation/wormhole-go-sdk/protocols/tokenbridge"
//	    "github.com/wormhole-foundation/wormhole-go-sdk/protocols/cctp"
//	)
//
// For more information, see the README and examples directory.
package wormhole
