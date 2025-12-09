package types

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type UniversalAddress [32]byte

func NewUniversalAddress(data []byte) (UniversalAddress, error) {
	var addr UniversalAddress
	if len(data) > 32 {
		return addr, fmt.Errorf("address too long: %d bytes", len(data))
	}
	// Right-pad with zeros if needed
	copy(addr[32-len(data):], data)
	return addr, nil
}

func (a UniversalAddress) String() string {
	return "0x" + hex.EncodeToString(a[:])
}

func (a UniversalAddress) Bytes() []byte {
	return a[:]
}

type ChainAddress struct {
	Chain   Chain
	Address string
}

func NewChainAddress(chain Chain, address string) ChainAddress {
	return ChainAddress{
		Chain:   chain,
		Address: address,
	}
}

func (ca ChainAddress) String() string {
	return fmt.Sprintf("%s:%s", ca.Chain, ca.Address)
}

func (ca ChainAddress) ToUniversalAddress() (UniversalAddress, error) {
	platform := ca.Chain.GetPlatform()

	switch platform {
	case PlatformEVM:
		// Remove 0x prefix if present
		addr := strings.TrimPrefix(ca.Address, "0x")
		data, err := hex.DecodeString(addr)
		if err != nil {
			return UniversalAddress{}, fmt.Errorf("invalid EVM address: %w", err)
		}
		return NewUniversalAddress(data)

	case PlatformSolana:
		// Solana addresses are base58 encoded
		// This would require base58 decoding
		return UniversalAddress{}, fmt.Errorf("solana address conversion not yet implemented")

	default:
		return UniversalAddress{}, fmt.Errorf("unsupported platform: %s", platform)
	}
}

type NativeAddress interface {
	String() string
	ToUniversalAddress() (UniversalAddress, error)
	Platform() Platform
}
