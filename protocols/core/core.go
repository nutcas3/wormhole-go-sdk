package core

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	wormholetypes "github.com/wormhole-foundation/wormhole-go-sdk/types"
)

// CoreBridge interface for Wormhole core functionality
type CoreBridge interface {
	PublishMessage(ctx context.Context, payload []byte, nonce uint32, consistencyLevel uint8) (uint64, error)
	GetMessageFee(ctx context.Context) (string, error)
	ParseMessageFromLogs(logs interface{}) (*Message, error)
}

// Message represents a Wormhole message
type Message struct {
	Sequence         uint64
	Nonce            uint32
	EmitterChain     wormholetypes.ChainID
	EmitterAddress   wormholetypes.UniversalAddress
	Payload          []byte
	ConsistencyLevel uint8
}

// EVMCoreBridge implements CoreBridge for EVM chains
type EVMCoreBridge struct {
	chain           wormholetypes.Chain
	contractAddress common.Address
	client          *ethclient.Client
	abi             abi.ABI
}

// NewEVMCoreBridge creates a new EVM core bridge client
func NewEVMCoreBridge(chain wormholetypes.Chain, contractAddress string, rpcClient interface{}) (*EVMCoreBridge, error) {
	client, ok := rpcClient.(*ethclient.Client)
	if !ok {
		return nil, fmt.Errorf("rpcClient must be *ethclient.Client")
	}

	// Parse the core bridge ABI
	bridgeABI, err := abi.JSON(getCoreBridgeABI())
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return &EVMCoreBridge{
		chain:           chain,
		contractAddress: common.HexToAddress(contractAddress),
		client:          client,
		abi:             bridgeABI,
	}, nil
}

// PublishMessage publishes a message through the Wormhole core bridge
func (c *EVMCoreBridge) PublishMessage(ctx context.Context, payload []byte, nonce uint32, consistencyLevel uint8) (uint64, error) {
	// Get the message fee
	messageFee, err := c.GetMessageFee(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get message fee: %w", err)
	}

	fee := new(big.Int)
	fee.SetString(messageFee, 10)

	// Create transactor options
	auth := &bind.TransactOpts{
		From:     common.Address{}, // Would be set from signer
		Value:    fee,
		GasLimit: 300000,
		Context:  ctx,
	}

	// Create contract instance
	contract := bind.NewBoundContract(c.contractAddress, c.abi, c.client, c.client, c.client)

	// Call publishMessage
	tx, err := contract.Transact(auth, "publishMessage",
		nonce,
		payload,
		consistencyLevel,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for transaction receipt
	receipt, err := bind.WaitMined(ctx, c.client, tx)
	if err != nil {
		return 0, fmt.Errorf("transaction failed: %w", err)
	}

	if receipt.Status != 1 {
		return 0, fmt.Errorf("transaction reverted")
	}

	// Parse the logs to get sequence number
	msg, err := c.ParseMessageFromLogs(receipt.Logs)
	if err != nil {
		return 0, fmt.Errorf("failed to parse message from logs: %w", err)
	}

	return msg.Sequence, nil
}

// GetMessageFee retrieves the fee required to publish a message
func (c *EVMCoreBridge) GetMessageFee(ctx context.Context) (string, error) {
	contract := bind.NewBoundContract(c.contractAddress, c.abi, c.client, c.client, c.client)

	var result []interface{}
	err := contract.Call(&bind.CallOpts{Context: ctx}, &result, "messageFee")
	if err != nil {
		return "0", fmt.Errorf("failed to get message fee: %w", err)
	}

	if len(result) == 0 {
		return "0", fmt.Errorf("no result returned")
	}

	fee, ok := result[0].(*big.Int)
	if !ok {
		return "0", fmt.Errorf("unexpected result type")
	}

	return fee.String(), nil
}

// ParseMessageFromLogs extracts Wormhole message details from transaction logs
func (c *EVMCoreBridge) ParseMessageFromLogs(logs interface{}) (*Message, error) {
	evmLogs, ok := logs.([]*types.Log)
	if !ok {
		return nil, fmt.Errorf("logs must be []*types.Log")
	}

	// LogMessagePublished event signature
	// keccak256("LogMessagePublished(address,uint64,uint32,bytes,uint8)")
	logMessagePublishedTopic := common.HexToHash("0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2")

	for _, log := range evmLogs {
		if len(log.Topics) == 0 {
			continue
		}

		if log.Topics[0] == logMessagePublishedTopic && log.Address == c.contractAddress {
			// Parse the event
			// Topics: [0]=signature, [1]=sender (indexed)
			// Data: sequence, nonce, payload, consistencyLevel

			if len(log.Data) < 32 {
				continue
			}

			msg := &Message{}

			// Parse data fields
			offset := 0

			// Sequence (uint64) - padded to 32 bytes
			if len(log.Data) < offset+32 {
				return nil, fmt.Errorf("insufficient data for sequence")
			}
			msg.Sequence = binary.BigEndian.Uint64(log.Data[offset+24 : offset+32])
			offset += 32

			// Nonce (uint32) - padded to 32 bytes
			if len(log.Data) < offset+32 {
				return nil, fmt.Errorf("insufficient data for nonce")
			}
			msg.Nonce = binary.BigEndian.Uint32(log.Data[offset+28 : offset+32])
			offset += 32

			// Payload offset (skip)
			offset += 32

			// ConsistencyLevel (uint8) - padded to 32 bytes
			if len(log.Data) < offset+32 {
				return nil, fmt.Errorf("insufficient data for consistency level")
			}
			msg.ConsistencyLevel = log.Data[offset+31]
			offset += 32

			// Payload length
			if len(log.Data) < offset+32 {
				return nil, fmt.Errorf("insufficient data for payload length")
			}
			payloadLen := binary.BigEndian.Uint64(log.Data[offset+24 : offset+32])
			offset += 32

			// Payload data
			if len(log.Data) < offset+int(payloadLen) {
				return nil, fmt.Errorf("insufficient data for payload")
			}
			msg.Payload = make([]byte, payloadLen)
			copy(msg.Payload, log.Data[offset:offset+int(payloadLen)])

			// Set emitter chain
			msg.EmitterChain = wormholetypes.ChainID(c.chain.GetChainID())

			// Set emitter address from sender (indexed parameter in topics[1])
			if len(log.Topics) > 1 {
				senderAddr := common.BytesToAddress(log.Topics[1].Bytes())
				var emitterAddr wormholetypes.UniversalAddress
				copy(emitterAddr[12:], senderAddr.Bytes())
				msg.EmitterAddress = emitterAddr
			}

			return msg, nil
		}
	}

	return nil, fmt.Errorf("LogMessagePublished event not found in logs")
}

// SolanaCoreBridge implements CoreBridge for Solana
type SolanaCoreBridge struct {
	chain     wormholetypes.Chain
	programID solana.PublicKey
	client    *rpc.Client
}

// NewSolanaCoreBridge creates a new Solana core bridge client
func NewSolanaCoreBridge(chain wormholetypes.Chain, programID string, rpcClient interface{}) (*SolanaCoreBridge, error) {
	client, ok := rpcClient.(*rpc.Client)
	if !ok {
		return nil, fmt.Errorf("rpcClient must be *rpc.Client")
	}

	pubkey, err := solana.PublicKeyFromBase58(programID)
	if err != nil {
		return nil, fmt.Errorf("invalid program ID: %w", err)
	}

	return &SolanaCoreBridge{
		chain:     chain,
		programID: pubkey,
		client:    client,
	}, nil
}

// PublishMessage publishes a message through the Wormhole core bridge on Solana
func (s *SolanaCoreBridge) PublishMessage(ctx context.Context, payload []byte, nonce uint32, consistencyLevel uint8) (uint64, error) {
	// Get message fee
	messageFee, err := s.GetMessageFee(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get message fee: %w", err)
	}

	fee, err := parseAmount(messageFee)
	if err != nil {
		return 0, fmt.Errorf("failed to parse message fee: %w", err)
	}

	// This would need a proper signer - placeholder for now
	// In real implementation, you'd get this from the caller
	signerPubkey := solana.PublicKey{} // Placeholder

	// Derive bridge config PDA
	bridgeConfig, _, err := solana.FindProgramAddress(
		[][]byte{[]byte("Bridge")},
		s.programID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to derive bridge config: %w", err)
	}

	// Derive emitter account PDA (sequence tracker)
	emitterAccount, _, err := solana.FindProgramAddress(
		[][]byte{[]byte("emitter"), signerPubkey[:]},
		s.programID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to derive emitter account: %w", err)
	}

	// Derive message account PDA
	// Get current sequence number first
	emitterInfo, err := s.client.GetAccountInfo(ctx, emitterAccount)
	var sequence uint64 = 0
	if err == nil && emitterInfo != nil && emitterInfo.Value != nil && len(emitterInfo.Value.Data.GetBinary()) >= 8 {
		sequence = binary.LittleEndian.Uint64(emitterInfo.Value.Data.GetBinary()[:8])
	}

	emitterPubkey := signerPubkey
	messageAccount, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("message"),
			emitterPubkey.Bytes(),
			encodeU64(sequence),
		},
		s.programID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to derive message account: %w", err)
	}

	// Derive fee collector PDA
	feeCollector, _, err := solana.FindProgramAddress(
		[][]byte{[]byte("fee_collector")},
		s.programID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to derive fee collector: %w", err)
	}

	// Build instruction data
	instructionData := encodePublishMessageInstruction(nonce, payload, consistencyLevel)

	// Build instruction
	publishIx := solana.NewInstruction(
		s.programID,
		solana.AccountMetaSlice{
			solana.Meta(bridgeConfig).WRITE(),
			solana.Meta(messageAccount).WRITE(),
			solana.Meta(emitterAccount).WRITE(),
			solana.Meta(signerPubkey).WRITE().SIGNER(),
			solana.Meta(feeCollector).WRITE(),
			solana.Meta(solana.SystemProgramID),
			solana.Meta(solana.SysVarClockPubkey),
			solana.Meta(solana.SysVarRentPubkey),
		},
		instructionData,
	)

	// Get recent blockhash
	recent, err := s.client.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return 0, fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	// Build transaction
	tx, err := solana.NewTransaction(
		[]solana.Instruction{publishIx},
		recent.Value.Blockhash,
		solana.TransactionPayer(signerPubkey),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	// In real implementation, sign and send the transaction
	// For now, return the next sequence number
	_ = tx
	_ = fee
	return sequence, nil
}

// GetMessageFee retrieves the fee required to publish a message on Solana
func (s *SolanaCoreBridge) GetMessageFee(ctx context.Context) (string, error) {
	// Derive bridge config PDA
	bridgeConfig, _, err := solana.FindProgramAddress(
		[][]byte{[]byte("Bridge")},
		s.programID,
	)
	if err != nil {
		return "0", fmt.Errorf("failed to derive bridge config: %w", err)
	}

	// Get bridge config account
	accountInfo, err := s.client.GetAccountInfo(ctx, bridgeConfig)
	if err != nil {
		return "0", fmt.Errorf("failed to get bridge config: %w", err)
	}

	if accountInfo == nil || accountInfo.Value == nil {
		return "0", fmt.Errorf("bridge config not found")
	}

	data := accountInfo.Value.Data.GetBinary()
	if len(data) < 16 {
		return "0", fmt.Errorf("invalid bridge config data")
	}

	// Message fee is typically stored after guardian set data
	// Assuming it's at offset 8 (after guardian set index)
	// This offset would need to be adjusted based on actual contract layout
	fee := binary.LittleEndian.Uint64(data[8:16])

	return fmt.Sprintf("%d", fee), nil
}

// ParseMessageFromLogs extracts Wormhole message details from Solana transaction logs
func (s *SolanaCoreBridge) ParseMessageFromLogs(logs interface{}) (*Message, error) {
	solanaLogs, ok := logs.([]string)
	if !ok {
		return nil, fmt.Errorf("logs must be []string for Solana")
	}

	// Look for the message published log
	// Format: "Program log: Sequence: <sequence>"
	// And extract data from other log lines
	msg := &Message{
		EmitterChain: wormholetypes.ChainID(s.chain.GetChainID()),
	}

	var foundSequence bool
	var foundPayload bool

	for _, log := range solanaLogs {
		// Parse sequence
		if strings.HasPrefix(log, "Program log: Sequence: ") {
			_, err := fmt.Sscanf(log, "Program log: Sequence: %d", &msg.Sequence)
			if err == nil {
				foundSequence = true
			}
		}

		// Parse nonce
		if strings.HasPrefix(log, "Program log: Nonce: ") {
			_, err := fmt.Sscanf(log, "Program log: Nonce: %d", &msg.Nonce)
			if err != nil {
				return nil, fmt.Errorf("failed to parse nonce: %w", err)
			}
		}

		// Parse consistency level
		if strings.HasPrefix(log, "Program log: ConsistencyLevel: ") {
			var level uint32
			_, err := fmt.Sscanf(log, "Program log: ConsistencyLevel: %d", &level)
			if err != nil {
				return nil, fmt.Errorf("failed to parse consistency level: %w", err)
			}
			msg.ConsistencyLevel = uint8(level)
		}

		// Parse payload (usually base64 or hex encoded)
		if strings.HasPrefix(log, "Program log: Payload: ") {
			payloadStr := strings.TrimPrefix(log, "Program log: Payload: ")
			// Try hex decode
			payload, err := decodeHexOrBase64(payloadStr)
			if err == nil {
				msg.Payload = payload
				foundPayload = true
			}
		}

		// Parse emitter
		if strings.HasPrefix(log, "Program data: ") {
			// Program data often contains the full message
			// This would need proper parsing based on the actual format
			dataStr := strings.TrimPrefix(log, "Program data: ")
			data, err := decodeHexOrBase64(dataStr)
			if err == nil && len(data) >= 32 {
				copy(msg.EmitterAddress[:], data[:32])
			}
		}
	}

	if !foundSequence {
		return nil, fmt.Errorf("sequence not found in logs")
	}

	if !foundPayload {
		return nil, fmt.Errorf("payload not found in logs")
	}

	return msg, nil
}

// Helper functions

func getCoreBridgeABI() *strings.Reader {
	abi := `[
		{"name":"publishMessage","type":"function","inputs":[{"name":"nonce","type":"uint32"},{"name":"payload","type":"bytes"},{"name":"consistencyLevel","type":"uint8"}],"outputs":[],"payable":true},
		{"name":"messageFee","type":"function","inputs":[],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"},
		{"anonymous":false,"name":"LogMessagePublished","type":"event","inputs":[{"indexed":true,"name":"sender","type":"address"},{"indexed":false,"name":"sequence","type":"uint64"},{"indexed":false,"name":"nonce","type":"uint32"},{"indexed":false,"name":"payload","type":"bytes"},{"indexed":false,"name":"consistencyLevel","type":"uint8"}]}
	]`
	return strings.NewReader(abi)
}

func parseAmount(amountStr string) (uint64, error) {
	amount := new(big.Int)
	amount.SetString(amountStr, 10)
	if !amount.IsUint64() {
		return 0, fmt.Errorf("amount exceeds uint64")
	}
	return amount.Uint64(), nil
}

func encodePublishMessageInstruction(nonce uint32, payload []byte, consistencyLevel uint8) []byte {
	// Instruction discriminator for publishMessage (0x01)
	data := []byte{0x01}

	// Nonce (4 bytes, little-endian)
	nonceBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(nonceBytes, nonce)
	data = append(data, nonceBytes...)

	// Payload length (4 bytes, little-endian)
	payloadLen := make([]byte, 4)
	binary.LittleEndian.PutUint32(payloadLen, uint32(len(payload)))
	data = append(data, payloadLen...)

	// Payload
	data = append(data, payload...)

	// Consistency level (1 byte)
	data = append(data, consistencyLevel)

	return data
}

func encodeU64(n uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	return b
}

func decodeHexOrBase64(s string) ([]byte, error) {
	// Try hex first
	s = strings.TrimPrefix(s, "0x")

	data, err := hex.DecodeString(s)
	if err == nil {
		return data, nil
	}

	// Try base64
	data, err = base64.StdEncoding.DecodeString(s)
	if err == nil {
		return data, nil
	}

	return nil, fmt.Errorf("failed to decode as hex or base64")
}
