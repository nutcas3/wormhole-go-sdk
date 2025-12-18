# Contributing to Wormhole Go SDK

Thank you for your interest in contributing to the Wormhole Go SDK! This document provides guidelines and instructions for contributing.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/nutcas3/wormhole-go-sdk.git
   cd wormhole-go-sdk
   ```
3. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Setup

### Prerequisites

- Go 1.25 or later
- Git

### Install Dependencies

```bash
go mod download
```

### Build the Project

```bash
go build ./...
```

### Run Tests

```bash
go test ./...
```

### Run Examples

```bash
cd examples/basic
go run main.go
```

## Code Style

### Go Formatting

- Use `gofmt` to format your code
- Run `go fmt ./...` before committing

### Naming Conventions

- Use camelCase for unexported identifiers
- Use PascalCase for exported identifiers
- Use descriptive names (avoid single-letter variables except in loops)

### Documentation

- Add godoc comments for all exported types, functions, and methods
- Comments should start with the name of the item being documented
- Provide examples in documentation where helpful

Example:
```go
// ChainAddress represents an address on a specific blockchain.
// It combines a chain identifier with a chain-specific address string.
type ChainAddress struct {
    Chain   Chain
    Address string
}
```

### Error Handling

- Always return errors, never panic in library code
- Wrap errors with context using `fmt.Errorf`
- Use descriptive error messages

Example:
```go
if err != nil {
    return fmt.Errorf("failed to initialize chain %s: %w", chain, err)
}
```

## Project Structure

```
wormhole-go-sdk/
├── types/              # Core type definitions
├── config/             # Configuration management
├── context/            # Chain context
├── platforms/          # Platform implementations
│   ├── evm/
│   ├── solana/
│   └── ...
├── protocols/          # Protocol implementations
│   ├── core/
│   ├── tokenbridge/
│   └── cctp/
├── transfer/          # Transfer abstractions
├── examples/          # Example applications
└── ...
```

## Adding New Features

### Adding a New Platform

1. Create a new directory under `platforms/`
2. Implement the `Platform` interface
3. Add platform-specific types (Address, Signer)
4. Update documentation

### Adding a New Protocol

1. Create a new directory under `protocols/`
2. Define the protocol interface
3. Implement platform-specific versions
4. Add tests
5. Update documentation

### Adding Examples

1. Create a new directory under `examples/`
2. Write a complete, runnable example
3. Add comments explaining each step
4. Update the main README

## Testing

### Unit Tests

- Write unit tests for all new functionality
- Place tests in `*_test.go` files
- Use table-driven tests where appropriate
- Aim for high test coverage

Example:
```go
func TestChainGetPlatform(t *testing.T) {
    tests := []struct {
        chain    Chain
        expected Platform
    }{
        {Ethereum, PlatformEVM},
        {Solana, PlatformSolana},
        {Aptos, PlatformAptos},
    }

    for _, tt := range tests {
        t.Run(string(tt.chain), func(t *testing.T) {
            if got := tt.chain.GetPlatform(); got != tt.expected {
                t.Errorf("GetPlatform() = %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### Integration Tests

- Integration tests should be marked with build tags
- Use testnet for integration tests
- Never use mainnet for automated tests

```go
//go:build integration
// +build integration

package integration_test
```

## Pull Request Process

1. **Update Documentation**: Ensure README and godoc comments are updated
2. **Add Tests**: Include tests for new functionality
3. **Run Tests**: Ensure all tests pass
4. **Format Code**: Run `go fmt ./...`
5. **Commit Messages**: Use clear, descriptive commit messages
6. **Create PR**: Submit a pull request with a clear description

### Commit Message Format

```
<type>: <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Example:
```
feat: add Sui platform support

- Implement Sui platform interface
- Add Sui address and signer types
- Update documentation with Sui examples

Closes #123
```

## Code Review

All contributions will be reviewed by maintainers. Please:

- Respond to feedback promptly
- Be open to suggestions
- Keep discussions professional and constructive

## Areas for Contribution

### High Priority

- Complete RPC client implementations for all platforms
- Add comprehensive test coverage
- Implement missing protocol methods
- Add more examples

### Medium Priority

- Performance optimizations
- Additional platform support
- Enhanced error messages
- Monitoring and metrics

### Documentation

- Improve godoc comments
- Add more examples
- Create tutorials
- Translate documentation

## Questions?

If you have questions:

1. Check existing issues and discussions
2. Review the documentation
3. Open a new issue with your question

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Code of Conduct

Please be respectful and professional in all interactions. We aim to maintain a welcoming and inclusive community.

Thank you for contributing to the Wormhole Go SDK!
