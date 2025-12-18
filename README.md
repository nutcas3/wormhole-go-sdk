# Wormhole Go SDK

A comprehensive Go SDK for interacting with the Wormhole cross-chain messaging protocol. This SDK provides a clean, idiomatic Go interface for building cross-chain applications.

## Features

- **Cross-Chain Messaging**: Send messages across multiple blockchain networks
- **Token Transfers**: Transfer tokens between different chains
- **Multiple Platforms**: Support for EVM, Solana, Aptos, Sui, Algorand, and CosmWasm
- **Type-Safe**: Strongly typed interfaces for all operations
- **Modular Design**: Use only the components you need
- **Protocol Support**: Core Bridge, Token Bridge, and CCTP integration
- **Production-Ready**: Full blockchain library integrations with real transaction handling

## Supported Platforms

### Platform-Specific Packages
- **EVM**: Ethereum, BSC, Polygon, Avalanche, Arbitrum, Optimism, Base, and more
- **Solana**: Native Solana integration with solana-go
- **Algorand**: Algorand blockchain support
- **Aptos**: Aptos Move-based chain
- **CosmWasm**: Terra, Injective, Osmosis, Cosmos Hub
- **Sui**: Sui Move-based chain

### Protocol-Specific Packages
- **Core Protocol**: Basic message passing with VAA support
- **Token Bridge**: Cross-chain token transfers with attestation
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
    wh, err := wormhole.New(
        types.Testnet,
        evm.New(),
        solana.New(),
    )
    if err != nil {
        log.Fatal(err)
    }

    chain, _ := wh.GetChain(types.Ethereum)
    fmt.Printf("Chain ID: %d\n", chain.ChainID())
    fmt.Printf("RPC: %s\n", chain.Config().RPC)
}
```

### Token Transfer Example

```go
import (
    "github.com/wormhole-foundation/wormhole-go-sdk/transfer"
)

// Create transfer
token := wormhole.TokenID(types.Ethereum, "0xTokenAddress")
amount := types.NewAmount("1000000", 6)
source := wormhole.ChainAddress(types.Ethereum, "0xSourceAddress")
dest := wormhole.ChainAddress(types.Solana, "DestAddress")

xfer := transfer.NewTokenTransfer(wh, token, amount, source, dest, false, nil, "")

// Get quote
quote, _ := xfer.QuoteTransfer(ctx)
fmt.Printf("Estimated time: %v\n", quote.EstimatedTime)
fmt.Printf("Total fee: %s\n", quote.TotalFee)

// Execute transfer (requires signer)
srcTxHashes, _ := xfer.InitiateTransfer(ctx, signer)
attestIds, _ := xfer.FetchAttestation(ctx, 60*time.Second)
destTxHashes, _ := xfer.CompleteTransfer(ctx, destSigner)
```

## Architecture

### Core Components

The SDK follows a modular, layered architecture:

1. **Types Layer** (`types/`): Core type definitions
   - Network, Chain, Platform types
   - Address and token abstractions
   - Signer interfaces

2. **Configuration Layer** (`config/`): Network and chain configurations
   - Pre-configured settings for Mainnet, Testnet, Devnet
   - Customizable RPC endpoints and contract addresses

3. **Context Layer** (`context/`): Unified chain interfaces
   - ChainContext for chain operations
   - RPC and protocol client caching

4. **Platform Layer** (`platforms/`): Platform-specific implementations
   - EVM, Solana, Aptos, Sui, Algorand, CosmWasm
   - Native blockchain library integrations

5. **Protocol Layer** (`protocols/`): Protocol implementations
   - Core Bridge for message passing
   - Token Bridge for token transfers
   - VAA parsing and verification

6. **Transfer Layer** (`transfer/`): High-level transfer abstractions
   - TokenTransfer for standard transfers
   - CircleTransfer for CCTP

### Project Structure

```
wormhole-go-sdk/
├── types/              # Core type definitions
├── config/             # Network and chain configurations
├── context/            # Chain context implementation
├── platforms/          # Platform-specific implementations
│   ├── evm/           # Ethereum and EVM chains
│   ├── solana/        # Solana integration
│   ├── aptos/         # Aptos integration
│   ├── sui/           # Sui integration
│   ├── algorand/      # Algorand integration
│   └── cosmwasm/      # CosmWasm integration
├── protocols/          # Protocol implementations
│   ├── core/          # Core bridge (message passing)
│   ├── tokenbridge/   # Token bridge
│   └── vaa.go         # VAA parsing
├── rpc/               # RPC client interfaces
├── transfer/          # Transfer abstractions
└── examples/          # Example applications
```

## RPC Clients

All platforms include production-ready RPC client implementations:

### EVM RPC Client
- Uses `go-ethereum` ethclient
- Methods: GetBlockNumber, GetBalance, GetTransactionReceipt, SendTransaction, EstimateGas, CallContract

### Solana RPC Client
- Uses `solana-go` RPC client
- Methods: GetBlockHeight, GetBalance, GetAccountInfo, GetTransaction, SendTransaction, SimulateTransaction

### Aptos RPC Client
- HTTP REST API client
- Methods: GetBlockHeight, GetBalance, GetAccountInfo, SubmitTransaction, SimulateTransaction

### Sui RPC Client
- JSON-RPC 2.0 client
- Methods: GetLatestCheckpoint, GetBalance, GetObject, ExecuteTransaction, DryRunTransaction

### Algorand RPC Client
- REST API client
- Methods: GetBlockHeight, GetBalance, GetAccountInfo, SendTransaction, WaitForConfirmation

### CosmWasm RPC Client
- Cosmos REST + Tendermint RPC
- Methods: GetBlockHeight, GetBalance, GetAccountInfo, BroadcastTx, QueryContract

See [rpc/README.md](rpc/README.md) for detailed RPC client documentation.

## Protocols

### Core Bridge

Message publishing and VAA handling:

```go
import "github.com/wormhole-foundation/wormhole-go-sdk/protocols/core"

// Publish message
sequence, err := coreBridge.PublishMessage(ctx, payload, nonce, consistencyLevel)

// Get message fee
fee, err := coreBridge.GetMessageFee(ctx)

// Parse message from logs
msg, err := coreBridge.ParseMessageFromLogs(logs)
```

### Token Bridge

Cross-chain token transfers:

```go
import "github.com/wormhole-foundation/wormhole-go-sdk/protocols/tokenbridge"

// Transfer tokens
txHash, err := tokenBridge.Transfer(ctx, token, amount, recipient, signer)

// Complete transfer with VAA
txHash, err := tokenBridge.CompleteTransfer(ctx, vaa, signer)

// Attest token
txHash, err := tokenBridge.AttestToken(ctx, token, signer)

// Create wrapped token
txHash, err := tokenBridge.CreateWrapped(ctx, vaa, signer)

// Check if wrapped
isWrapped, err := tokenBridge.IsWrappedAsset(ctx, token)
```

### VAA Parsing

```go
import "github.com/wormhole-foundation/wormhole-go-sdk/protocols/core"

// Parse VAA
vaa, err := core.ParseVAA(vaaBytes)

// Access VAA fields
fmt.Printf("Sequence: %d\n", vaa.Sequence)
fmt.Printf("Emitter Chain: %d\n", vaa.EmitterChain)
fmt.Printf("Payload: %x\n", vaa.Payload)

// Verify signatures
err = vaa.VerifySignatures(guardianSet)
```

## Implementation Details

### EVM Token Bridge
- Approves ERC20 tokens and calls `transferTokens` contract method
- Redeems tokens using VAA via `completeTransfer`
- Creates token attestations and wrapped tokens
- Full transaction handling with go-ethereum

### Solana Token Bridge
- Builds approve + transfer instructions with PDA derivation
- Posts VAA and redeems with claim account validation
- Creates attestations with token mint metadata
- Derives wrapped mint PDAs and creates wrapped tokens
- Uses solana-go for transaction building and signing

### Core Bridge
- **EVM**: Publishes messages with fee handling, parses `LogMessagePublished` events
- **Solana**: Creates publish message instructions with PDA derivation, parses transaction logs

## Examples

See the [examples](./examples) directory for complete working examples:

- **basic**: SDK initialization and chain queries
- **transfer**: Token transfer workflow
- **rpc**: RPC client usage examples

Run examples:

```bash
cd examples/basic
go run main.go

cd examples/transfer
go run main.go
```

## API Reference

### Core Types

```go
// Network types
type Network string
const (
    Mainnet Network = "Mainnet"
    Testnet Network = "Testnet"
    Devnet  Network = "Devnet"
)

// Chain types
type Chain string
const (
    Ethereum, Solana, BSC, Polygon, Avalanche, Arbitrum, Optimism, Base, ...
)

// Address types
type ChainAddress struct {
    Chain   Chain
    Address string
}

type UniversalAddress [32]byte

// Token types
type TokenID struct {
    Chain   Chain
    Address string
}

type Amount struct {
    Value    string
    Decimals uint8
}

// Signer interfaces
type Signer interface {
    Chain() Chain
    Address() string
}
```

### Main SDK

```go
// Initialize SDK
wh, err := wormhole.New(network Network, platforms ...Platform)

// Get chain context
chain, err := wh.GetChain(chain Chain)

// Get all chains
chains := wh.GetChains()

// Helper functions
addr := wormhole.ChainAddress(chain Chain, address string)
token := wormhole.TokenID(chain Chain, address string)
canonical := wormhole.CanonicalAddress(addr ChainAddress)
```

### Transfer Package

```go
// Create token transfer
xfer := transfer.NewTokenTransfer(
    wormhole,
    token TokenID,
    amount Amount,
    source ChainAddress,
    destination ChainAddress,
    automatic bool,
    payload []byte,
    nativeGas string,
)

// Transfer operations
quote, err := xfer.QuoteTransfer(ctx)
srcTxs, err := xfer.InitiateTransfer(ctx, signer)
attestIds, err := xfer.FetchAttestation(ctx, timeout)
destTxs, err := xfer.CompleteTransfer(ctx, signer)
transfer := xfer.GetTransfer()
```

## Best Practices

1. **Always check errors**: Never ignore error return values
2. **Use contexts**: Pass contexts for cancellation and timeouts
3. **Test on testnet**: Always test thoroughly on testnet before mainnet
4. **Secure private keys**: Never hardcode or log private keys
5. **Validate inputs**: Validate addresses and amounts before transfers
6. **Handle state**: Check transfer state before operations
7. **Set timeouts**: Use reasonable timeouts for network operations
8. **Monitor transactions**: Track transaction hashes and confirmations

## Development

### Building

```bash
go build ./...
```

### Testing

```bash
go test ./...
```

### Dependencies

```bash
go mod download
go mod tidy
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
- [Contributing Guide](docs/CONTRIBUTING.md)

## License

MIT License

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## Disclaimer

This SDK is provided as-is. Always test thoroughly on testnets before using on mainnet with real assets.
