# Wormhole Go SDK

A comprehensive Go SDK for interacting with the Wormhole cross-chain messaging protocol. This SDK provides a clean, idiomatic Go interface for building cross-chain applications.

## Features

- **Cross-Chain Messaging**: Send messages across multiple blockchain networks
- **Token Transfers**: Transfer tokens between different chains
- **Multiple Platforms**: Support for EVM, Solana, Aptos, Sui, Algorand, and CosmWasm
- **Type-Safe**: Strongly typed interfaces for all operations
- **Modular Design**: Use only the components you need
- **Protocol Support**: Core Bridge, Token Bridge, and CCTP integration

## Supported Platforms

### Platform-Specific Packages
- **EVM**: Ethereum, BSC, Polygon, Avalanche, Arbitrum, Optimism, Base, and more
- **Solana**: Native Solana integration
- **Algorand**: Algorand blockchain support
- **Aptos**: Aptos Move-based chain
- **CosmWasm**: Terra, Injective, Osmosis, Cosmos Hub
- **Sui**: Sui Move-based chain

### Protocol-Specific Packages
- **Core Protocol**: Basic message passing
- **Token Bridge (WTT)**: Cross-chain token transfers
- **CCTP**: Circle's Cross-Chain Transfer Protocol for native USDC

## Installation

```bash
go get github.com/wormhole-foundation/wormhole-go-sdk
```

## Quick Start

### Initialize the SDK

```go
package main

import (
    "context"
    "fmt"
    "log"

    wormhole "github.com/wormhole-foundation/wormhole-go-sdk"
    "github.com/wormhole-foundation/wormhole-go-sdk/platforms/evm"
    "github.com/wormhole-foundation/wormhole-go-sdk/platforms/solana"
    "github.com/wormhole-foundation/wormhole-go-sdk/types"
)

func main() {
    // Initialize with desired platforms
    wh, err := wormhole.New(
        types.Testnet,
        evm.New(),
        solana.New(),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Get chain information
    chain, _ := wh.GetChain(types.Ethereum)
    fmt.Printf("Chain ID: %d\n", chain.ChainID())
    fmt.Printf("RPC: %s\n", chain.Config().RPC)
}
```

### Fetch Chain Information

```go
// Get Solana chain context
solanaChain, err := wh.GetChain(types.Solana)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Chain ID: %d\n", solanaChain.ChainID())
fmt.Printf("RPC: %s\n", solanaChain.Config().RPC)
fmt.Printf("Platform: %s\n", solanaChain.Platform())
```

### Create Addresses and Token IDs

```go
// Create chain addresses
senderAddr := wormhole.ChainAddress(types.Ethereum, "0xYourAddress")
receiverAddr := wormhole.ChainAddress(types.Solana, "YourSolanaAddress")

// Create token IDs
usdcToken := wormhole.TokenID(types.Ethereum, "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
nativeETH := wormhole.TokenID(types.Ethereum, "native")

// Get canonical address
canonicalAddr := wormhole.CanonicalAddress(senderAddr)
```

## Token Transfers

### Manual Token Transfer

```go
package main

import (
    "context"
    "time"

    wormhole "github.com/wormhole-foundation/wormhole-go-sdk"
    "github.com/wormhole-foundation/wormhole-go-sdk/platforms/evm"
    "github.com/wormhole-foundation/wormhole-go-sdk/transfer"
    "github.com/wormhole-foundation/wormhole-go-sdk/types"
)

func main() {
    ctx := context.Background()
    
    // Initialize SDK
    wh, _ := wormhole.New(types.Testnet, evm.New())
    
    // Define transfer parameters
    token := wormhole.TokenID(types.Ethereum, "0xTokenAddress")
    amount := types.NewAmount("1000000", 6) // 1 USDC
    source := wormhole.ChainAddress(types.Ethereum, "0xSourceAddress")
    dest := wormhole.ChainAddress(types.Solana, "DestAddress")
    
    // Create transfer
    xfer := transfer.NewTokenTransfer(wh, token, amount, source, dest, false, nil, "")
    
    // Get quote
    quote, _ := xfer.QuoteTransfer(ctx)
    
    // 1. Initiate transfer (requires signer)
    // srcTxHashes, _ := xfer.InitiateTransfer(ctx, signer)
    
    // 2. Wait for attestation
    // attestIds, _ := xfer.FetchAttestation(ctx, 60*time.Second)
    
    // 3. Complete transfer
    // destTxHashes, _ := xfer.CompleteTransfer(ctx, destSigner)
}
```

### Automatic Token Transfer

For automatic transfers, set the `automatic` parameter to `true`:

```go
xfer := transfer.NewTokenTransfer(
    wh,
    token,
    amount,
    source,
    dest,
    true,  // Automatic transfer
    nil,
    "",
)

// Only need to initiate - completion is automatic
srcTxHashes, _ := xfer.InitiateTransfer(ctx, signer)
```

## Architecture

### Core Types

- **Network**: Mainnet, Testnet, or Devnet
- **Chain**: Specific blockchain (Ethereum, Solana, etc.)
- **Platform**: Blockchain platform type (EVM, Solana, etc.)
- **ChainAddress**: Address on a specific chain
- **TokenID**: Token identifier (chain + address)
- **UniversalAddress**: 32-byte address format used across all chains

### Chain Context

The `ChainContext` provides a unified interface for interacting with chains:

```go
chain, _ := wh.GetChain(types.Ethereum)

// Get platform-specific RPC client
rpcClient := chain.GetRPCClient()

// Get protocol clients
coreBridge, _ := chain.GetCoreBridge()
tokenBridge, _ := chain.GetTokenBridge()
```

### Signers

The SDK supports two types of signers:

- **SignOnlySigner**: Signs transactions without broadcasting
- **SignAndSendSigner**: Signs and broadcasts transactions

```go
// Create an EVM signer
signer, err := evm.NewSigner(types.Ethereum, "YOUR_PRIVATE_KEY")

// Create a Solana signer
solanaSigner := solana.NewSigner([]byte("YOUR_PRIVATE_KEY"))
```

## Protocols

### Core Bridge

```go
import "github.com/wormhole-foundation/wormhole-go-sdk/protocols/core"

// Publish a message
sequence, err := coreBridge.PublishMessage(ctx, payload, nonce, consistencyLevel)

// Get message fee
fee, err := coreBridge.GetMessageFee(ctx)
```

### Token Bridge

```go
import "github.com/wormhole-foundation/wormhole-go-sdk/protocols/tokenbridge"

// Transfer tokens
txHash, err := tokenBridge.Transfer(ctx, token, amount, recipient, signer)

// Complete transfer with VAA
txHash, err := tokenBridge.CompleteTransfer(ctx, vaa, signer)

// Attest token
txHash, err := tokenBridge.AttestToken(ctx, token, signer)
```

### CCTP (Circle)

```go
import "github.com/wormhole-foundation/wormhole-go-sdk/protocols/cctp"

// Transfer native USDC
txHash, err := cctp.Transfer(ctx, amount, recipient, signer)

// Complete transfer
txHash, err := cctp.CompleteTransfer(ctx, message, attestation, signer)
```

## Examples

See the [examples](./examples) directory for complete working examples:

- **basic**: Basic SDK initialization and chain queries
- **transfer**: Token transfer workflow

Run examples:

```bash
cd examples/basic
go run main.go

cd examples/transfer
go run main.go
```

## Network Configuration

The SDK comes with pre-configured endpoints for all networks:

- **Mainnet**: Production network with real assets
- **Testnet**: Test network for development
- **Devnet**: Local development network

You can customize RPC endpoints and contract addresses as needed.

## Development

### Project Structure

```
wormhole-go-sdk/
├── types/              # Core type definitions
├── config/             # Network and chain configurations
├── context/            # Chain context implementation
├── platforms/          # Platform-specific implementations
│   ├── evm/
│   ├── solana/
│   ├── aptos/
│   ├── sui/
│   ├── algorand/
│   └── cosmwasm/
├── protocols/          # Protocol implementations
│   ├── core/          # Core bridge
│   ├── tokenbridge/   # Token bridge
│   └── cctp/          # Circle CCTP
├── transfer/          # Transfer abstractions
└── examples/          # Example applications
```

### Building

```bash
go build ./...
```

### Testing

```bash
go test ./...
```

## Comparison with TypeScript SDK

This Go SDK mirrors the architecture of the official Wormhole TypeScript SDK:

| TypeScript | Go |
|------------|-----|
| `@wormhole-foundation/sdk` | `github.com/wormhole-foundation/wormhole-go-sdk` |
| `@wormhole-foundation/sdk-evm` | `platforms/evm` |
| `@wormhole-foundation/sdk-solana` | `platforms/solana` |
| `@wormhole-foundation/sdk-definitions` | `types` |
| `@wormhole-foundation/sdk-connect` | `context` |

## Resources

- [Wormhole Documentation](https://wormhole.com/docs/)
- [TypeScript SDK Reference](https://wormhole.com/docs/tools/typescript-sdk/sdk-reference/)
- [Wormhole GitHub](https://github.com/wormhole-foundation)

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Disclaimer

This SDK is provided as-is. Always test thoroughly on testnets before using on mainnet with real assets.
