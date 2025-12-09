package core

import (
	"context"
	"fmt"

	"github.com/wormhole-foundation/wormhole-go-sdk/types"
)

type CoreBridge interface {
	PublishMessage(ctx context.Context, payload []byte, nonce uint32, consistencyLevel uint8) (uint64, error)

	GetMessageFee(ctx context.Context) (string, error)

	ParseMessageFromLogs(logs interface{}) (*Message, error)
}

// Message represents a Wormhole message
type Message struct {
	Sequence         uint64
	Nonce            uint32
	EmitterChain     types.ChainID
	EmitterAddress   types.UniversalAddress
	Payload          []byte
	ConsistencyLevel uint8
}

// EVMCoreBridge implements CoreBridge for EVM chains
type EVMCoreBridge struct {
	chain           types.Chain
	contractAddress string
	rpcClient       interface{}
}

// NewEVMCoreBridge creates a new EVM core bridge client
func NewEVMCoreBridge(chain types.Chain, contractAddress string, rpcClient interface{}) *EVMCoreBridge {
	return &EVMCoreBridge{
		chain:           chain,
		contractAddress: contractAddress,
		rpcClient:       rpcClient,
	}
}

func (c *EVMCoreBridge) PublishMessage(ctx context.Context, payload []byte, nonce uint32, consistencyLevel uint8) (uint64, error) {
	// Implementation would interact with the EVM contract
	return 0, fmt.Errorf("not implemented")
}

func (c *EVMCoreBridge) GetMessageFee(ctx context.Context) (string, error) {
	// Implementation would query the contract
	return "0", fmt.Errorf("not implemented")
}

func (c *EVMCoreBridge) ParseMessageFromLogs(logs interface{}) (*Message, error) {
	// Implementation would parse EVM logs
	return nil, fmt.Errorf("not implemented")
}

// SolanaCoreBridge implements CoreBridge for Solana
type SolanaCoreBridge struct {
	chain     types.Chain
	programID string
	rpcClient interface{}
}

// NewSolanaCoreBridge creates a new Solana core bridge client
func NewSolanaCoreBridge(chain types.Chain, programID string, rpcClient interface{}) *SolanaCoreBridge {
	return &SolanaCoreBridge{
		chain:     chain,
		programID: programID,
		rpcClient: rpcClient,
	}
}

func (s *SolanaCoreBridge) PublishMessage(ctx context.Context, payload []byte, nonce uint32, consistencyLevel uint8) (uint64, error) {
	// Implementation would interact with the Solana program
	return 0, fmt.Errorf("not implemented")
}

func (s *SolanaCoreBridge) GetMessageFee(ctx context.Context) (string, error) {
	// Implementation would query the program
	return "0", fmt.Errorf("not implemented")
}

func (s *SolanaCoreBridge) ParseMessageFromLogs(logs interface{}) (*Message, error) {
	// Implementation would parse Solana logs
	return nil, fmt.Errorf("not implemented")
}
