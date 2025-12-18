package tokenbridge

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gagliardetto/solana-go"
	solatoken "github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	wormholetypes "github.com/wormhole-foundation/wormhole-go-sdk/types"
)

// TokenBridge interface for cross-chain token transfers
type TokenBridge interface {
	Transfer(ctx context.Context, token wormholetypes.TokenID, amount wormholetypes.Amount, recipient wormholetypes.ChainAddress, signer wormholetypes.Signer) (wormholetypes.TxHash, error)
	CompleteTransfer(ctx context.Context, vaa []byte, signer wormholetypes.Signer) (wormholetypes.TxHash, error)
	AttestToken(ctx context.Context, token wormholetypes.TokenID, signer wormholetypes.Signer) (wormholetypes.TxHash, error)
	CreateWrapped(ctx context.Context, vaa []byte, signer wormholetypes.Signer) (wormholetypes.TxHash, error)
	GetWrappedAsset(ctx context.Context, token wormholetypes.TokenID) (string, error)
	IsWrappedAsset(ctx context.Context, token wormholetypes.TokenID) (bool, error)
}

// TransferParams contains parameters for token transfers
type TransferParams struct {
	Token      wormholetypes.TokenID
	Amount     wormholetypes.Amount
	Recipient  wormholetypes.ChainAddress
	RelayerFee string
	Payload    []byte
	ArbiterFee string
}

// EVMTokenBridge implements TokenBridge for EVM chains
type EVMTokenBridge struct {
	chain           wormholetypes.Chain
	contractAddress common.Address
	client          *ethclient.Client
	abi             abi.ABI
}

// NewEVMTokenBridge creates a new EVM token bridge client
func NewEVMTokenBridge(chain wormholetypes.Chain, contractAddress string, rpcClient interface{}) (*EVMTokenBridge, error) {
	client, ok := rpcClient.(*ethclient.Client)
	if !ok {
		return nil, fmt.Errorf("rpcClient must be *ethclient.Client")
	}

	// Parse the token bridge ABI
	bridgeABI, err := abi.JSON(getTokenBridgeABI())
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return &EVMTokenBridge{
		chain:           chain,
		contractAddress: common.HexToAddress(contractAddress),
		client:          client,
		abi:             bridgeABI,
	}, nil
}

// Transfer initiates a token transfer on EVM
func (t *EVMTokenBridge) Transfer(ctx context.Context, token wormholetypes.TokenID, amount wormholetypes.Amount, recipient wormholetypes.ChainAddress, signer wormholetypes.Signer) (wormholetypes.TxHash, error) {
	// Validate inputs
	if signer.Chain() != t.chain {
		return "", fmt.Errorf("signer chain %s does not match bridge chain %s", signer.Chain(), t.chain)
	}

	if token.Chain != t.chain {
		return "", fmt.Errorf("token chain %s does not match bridge chain %s", token.Chain, t.chain)
	}

	// Convert recipient to universal address format
	recipientAddr, err := recipient.ToUniversalAddress()
	if err != nil {
		return "", fmt.Errorf("failed to convert recipient address: %w", err)
	}

	// Get destination chain ID
	destChainID := recipient.Chain.GetChainID()

	// Parse amount
	amountBig := new(big.Int)
	amountBig.SetString(amount.Value, 10)

	// Prepare transfer parameters
	tokenAddr := common.HexToAddress(token.Address)

	// First, approve the token bridge to spend tokens
	auth, err := t.getTransactor(ctx, signer)
	if err != nil {
		return "", fmt.Errorf("failed to create transactor: %w", err)
	}

	// Approve token spending
	tokenABI, _ := abi.JSON(getERC20ABI())
	tokenContract := bind.NewBoundContract(tokenAddr, tokenABI, t.client, t.client, t.client)

	approveTx, err := tokenContract.Transact(auth, "approve", t.contractAddress, amountBig)
	if err != nil {
		return "", fmt.Errorf("failed to approve tokens: %w", err)
	}

	// Wait for approval
	receipt, err := bind.WaitMined(ctx, t.client, approveTx)
	if err != nil {
		return "", fmt.Errorf("approval transaction failed: %w", err)
	}

	if receipt.Status != 1 {
		return "", fmt.Errorf("approval transaction reverted")
	}

	// Now execute the transfer
	auth, err = t.getTransactor(ctx, signer)
	if err != nil {
		return "", fmt.Errorf("failed to create transactor: %w", err)
	}

	// Call transferTokens function
	bridgeContract := bind.NewBoundContract(t.contractAddress, t.abi, t.client, t.client, t.client)

	tx, err := bridgeContract.Transact(auth, "transferTokens",
		tokenAddr,
		amountBig,
		uint16(destChainID),
		recipientAddr,
		big.NewInt(0), // arbiter fee
		uint32(0),     // nonce
	)
	if err != nil {
		return "", fmt.Errorf("failed to transfer tokens: %w", err)
	}

	return wormholetypes.TxHash(tx.Hash().Hex()), nil
}

// CompleteTransfer completes a token transfer using a VAA
func (t *EVMTokenBridge) CompleteTransfer(ctx context.Context, vaa []byte, signer wormholetypes.Signer) (wormholetypes.TxHash, error) {
	if len(vaa) == 0 {
		return "", fmt.Errorf("VAA cannot be empty")
	}

	if signer.Chain() != t.chain {
		return "", fmt.Errorf("signer chain %s does not match bridge chain %s", signer.Chain(), t.chain)
	}

	auth, err := t.getTransactor(ctx, signer)
	if err != nil {
		return "", fmt.Errorf("failed to create transactor: %w", err)
	}

	bridgeContract := bind.NewBoundContract(t.contractAddress, t.abi, t.client, t.client, t.client)

	tx, err := bridgeContract.Transact(auth, "completeTransfer", vaa)
	if err != nil {
		return "", fmt.Errorf("failed to complete transfer: %w", err)
	}

	return wormholetypes.TxHash(tx.Hash().Hex()), nil
}

// AttestToken attests a token for cross-chain transfer
func (t *EVMTokenBridge) AttestToken(ctx context.Context, token wormholetypes.TokenID, signer wormholetypes.Signer) (wormholetypes.TxHash, error) {
	if signer.Chain() != t.chain {
		return "", fmt.Errorf("signer chain %s does not match bridge chain %s", signer.Chain(), t.chain)
	}

	if token.Chain != t.chain {
		return "", fmt.Errorf("token chain %s does not match bridge chain %s", token.Chain, t.chain)
	}

	auth, err := t.getTransactor(ctx, signer)
	if err != nil {
		return "", fmt.Errorf("failed to create transactor: %w", err)
	}

	tokenAddr := common.HexToAddress(token.Address)
	bridgeContract := bind.NewBoundContract(t.contractAddress, t.abi, t.client, t.client, t.client)

	tx, err := bridgeContract.Transact(auth, "attestToken", tokenAddr, uint32(0))
	if err != nil {
		return "", fmt.Errorf("failed to attest token: %w", err)
	}

	return wormholetypes.TxHash(tx.Hash().Hex()), nil
}

// CreateWrapped creates a wrapped token on the target chain
func (t *EVMTokenBridge) CreateWrapped(ctx context.Context, vaa []byte, signer wormholetypes.Signer) (wormholetypes.TxHash, error) {
	if len(vaa) == 0 {
		return "", fmt.Errorf("VAA cannot be empty")
	}

	if signer.Chain() != t.chain {
		return "", fmt.Errorf("signer chain %s does not match bridge chain %s", signer.Chain(), t.chain)
	}

	auth, err := t.getTransactor(ctx, signer)
	if err != nil {
		return "", fmt.Errorf("failed to create transactor: %w", err)
	}

	bridgeContract := bind.NewBoundContract(t.contractAddress, t.abi, t.client, t.client, t.client)

	tx, err := bridgeContract.Transact(auth, "createWrapped", vaa)
	if err != nil {
		return "", fmt.Errorf("failed to create wrapped token: %w", err)
	}

	return wormholetypes.TxHash(tx.Hash().Hex()), nil
}

// GetWrappedAsset returns the wrapped asset address for a token
func (t *EVMTokenBridge) GetWrappedAsset(ctx context.Context, token wormholetypes.TokenID) (string, error) {
	if token.Chain == t.chain {
		return "", fmt.Errorf("token is native to this chain, not wrapped")
	}

	tokenAddr, err := wormholetypes.NewChainAddress(token.Chain, token.Address).ToUniversalAddress()
	if err != nil {
		return "", fmt.Errorf("failed to convert token address: %w", err)
	}

	chainID := token.Chain.GetChainID()

	bridgeContract := bind.NewBoundContract(t.contractAddress, t.abi, t.client, t.client, t.client)

	var result []interface{}
	err = bridgeContract.Call(nil, &result, "wrappedAsset", uint16(chainID), tokenAddr)
	if err != nil {
		return "", fmt.Errorf("failed to get wrapped asset: %w", err)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("no result returned")
	}

	wrappedAddr, ok := result[0].(common.Address)
	if !ok {
		return "", fmt.Errorf("unexpected result type")
	}

	return wrappedAddr.Hex(), nil
}

// IsWrappedAsset checks if a token is a wrapped asset
func (t *EVMTokenBridge) IsWrappedAsset(ctx context.Context, token wormholetypes.TokenID) (bool, error) {
	if token.Chain != t.chain {
		return false, fmt.Errorf("token is not on this chain")
	}

	tokenAddr := common.HexToAddress(token.Address)
	bridgeContract := bind.NewBoundContract(t.contractAddress, t.abi, t.client, t.client, t.client)

	var result []interface{}
	err := bridgeContract.Call(nil, &result, "isWrappedAsset", tokenAddr)
	if err != nil {
		return false, fmt.Errorf("failed to check if wrapped: %w", err)
	}

	if len(result) == 0 {
		return false, fmt.Errorf("no result returned")
	}

	isWrapped, ok := result[0].(bool)
	if !ok {
		return false, fmt.Errorf("unexpected result type")
	}

	return isWrapped, nil
}

func (t *EVMTokenBridge) getTransactor(ctx context.Context, signer wormholetypes.Signer) (*bind.TransactOpts, error) {
	// TODO: Implement transactor creation based on Signer interface
	return nil, fmt.Errorf("transactor creation not implemented")
}

// SolanaTokenBridge implements TokenBridge for Solana
type SolanaTokenBridge struct {
	chain     wormholetypes.Chain
	programID solana.PublicKey
	client    *rpc.Client
}

// NewSolanaTokenBridge creates a new Solana token bridge client
func NewSolanaTokenBridge(chain wormholetypes.Chain, programID string, rpcClient interface{}) (*SolanaTokenBridge, error) {
	client, ok := rpcClient.(*rpc.Client)
	if !ok {
		return nil, fmt.Errorf("rpcClient must be *rpc.Client")
	}

	pubkey, err := solana.PublicKeyFromBase58(programID)
	if err != nil {
		return nil, fmt.Errorf("invalid program ID: %w", err)
	}

	return &SolanaTokenBridge{
		chain:     chain,
		programID: pubkey,
		client:    client,
	}, nil
}

// Transfer initiates a token transfer on Solana
func (t *SolanaTokenBridge) Transfer(ctx context.Context, token wormholetypes.TokenID, amount wormholetypes.Amount, recipient wormholetypes.ChainAddress, signer wormholetypes.Signer) (wormholetypes.TxHash, error) {
	if signer.Chain() != t.chain {
		return "", fmt.Errorf("signer chain %s does not match bridge chain %s", signer.Chain(), t.chain)
	}

	if token.Chain != t.chain {
		return "", fmt.Errorf("token chain %s does not match bridge chain %s", token.Chain, t.chain)
	}

	recipientAddr, err := recipient.ToUniversalAddress()
	if err != nil {
		return "", fmt.Errorf("failed to convert recipient address: %w", err)
	}

	destChainID := recipient.Chain.GetChainID()

	// Parse addresses
	signerPubkey, err := solana.PublicKeyFromBase58(signer.Address())
	if err != nil {
		return "", fmt.Errorf("invalid signer address: %w", err)
	}

	tokenMint, err := solana.PublicKeyFromBase58(token.Address)
	if err != nil {
		return "", fmt.Errorf("invalid token address: %w", err)
	}

	// Derive token account
	tokenAccount, _, err := solana.FindAssociatedTokenAddress(signerPubkey, tokenMint)
	if err != nil {
		return "", fmt.Errorf("failed to derive token account: %w", err)
	}

	// Derive bridge authority PDA
	bridgeAuthority, _, err := solana.FindProgramAddress(
		[][]byte{[]byte("authority_signer")},
		t.programID,
	)
	if err != nil {
		return "", fmt.Errorf("failed to derive bridge authority: %w", err)
	}

	// Parse amount
	amountU64, err := parseAmount(amount.Value)
	if err != nil {
		return "", fmt.Errorf("failed to parse amount: %w", err)
	}

	// Build transaction
	recent, err := t.client.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	// Create approve instruction
	approveIx := solatoken.NewApproveInstruction(
		amountU64,
		tokenAccount,
		bridgeAuthority,
		signerPubkey,
		[]solana.PublicKey{},
	).Build()

	// Create transfer instruction (simplified - actual implementation would be more complex)
	transferData := encodeTransferInstruction(amountU64, recipientAddr, uint16(destChainID))
	transferIx := solana.NewInstruction(
		t.programID,
		solana.AccountMetaSlice{
			solana.Meta(signerPubkey).WRITE().SIGNER(),
			solana.Meta(tokenAccount).WRITE(),
			solana.Meta(tokenMint),
			solana.Meta(bridgeAuthority),
		},
		transferData,
	)

	tx, err := solana.NewTransaction(
		[]solana.Instruction{approveIx, transferIx},
		recent.Value.Blockhash,
		solana.TransactionPayer(signerPubkey),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create transaction: %w", err)
	}

	// Sign transaction
	signedTx, err := t.signTransaction(tx, signer)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	sig, err := t.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return wormholetypes.TxHash(sig.String()), nil
}

// CompleteTransfer completes a token transfer using a VAA
func (t *SolanaTokenBridge) CompleteTransfer(ctx context.Context, vaa []byte, signer wormholetypes.Signer) (wormholetypes.TxHash, error) {
	if len(vaa) == 0 {
		return "", fmt.Errorf("VAA cannot be empty")
	}

	if signer.Chain() != t.chain {
		return "", fmt.Errorf("signer chain %s does not match bridge chain %s", signer.Chain(), t.chain)
	}

	signerPubkey, err := solana.PublicKeyFromBase58(signer.Address())
	if err != nil {
		return "", fmt.Errorf("invalid signer address: %w", err)
	}

	// Derive claim account from VAA hash
	vaaHash := solana.Hash(vaa)
	claimAccount, _, err := solana.FindProgramAddress(
		[][]byte{[]byte("claim"), vaaHash[:]},
		t.programID,
	)
	if err != nil {
		return "", fmt.Errorf("failed to derive claim account: %w", err)
	}

	// Check if already redeemed
	accountInfo, err := t.client.GetAccountInfo(ctx, claimAccount)
	if err == nil && accountInfo != nil && accountInfo.Value != nil {
		return "", fmt.Errorf("transfer already redeemed")
	}

	// Build transaction
	recent, err := t.client.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	// Post VAA instruction
	postVAAData := append([]byte{0x01}, vaa...)
	postVAAIx := solana.NewInstruction(
		t.programID,
		solana.AccountMetaSlice{
			solana.Meta(signerPubkey).WRITE().SIGNER(),
		},
		postVAAData,
	)

	// Complete transfer instruction
	completeData := append([]byte{0x02}, vaa...)
	completeIx := solana.NewInstruction(
		t.programID,
		solana.AccountMetaSlice{
			solana.Meta(signerPubkey).WRITE().SIGNER(),
			solana.Meta(claimAccount).WRITE(),
		},
		completeData,
	)

	tx, err := solana.NewTransaction(
		[]solana.Instruction{postVAAIx, completeIx},
		recent.Value.Blockhash,
		solana.TransactionPayer(signerPubkey),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create transaction: %w", err)
	}

	signedTx, err := t.signTransaction(tx, signer)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	sig, err := t.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return wormholetypes.TxHash(sig.String()), nil
}

// AttestToken attests a token for cross-chain transfer
func (t *SolanaTokenBridge) AttestToken(ctx context.Context, token wormholetypes.TokenID, signer wormholetypes.Signer) (wormholetypes.TxHash, error) {
	if signer.Chain() != t.chain {
		return "", fmt.Errorf("signer chain %s does not match bridge chain %s", signer.Chain(), t.chain)
	}

	if token.Chain != t.chain {
		return "", fmt.Errorf("token chain %s does not match bridge chain %s", token.Chain, t.chain)
	}

	signerPubkey, err := solana.PublicKeyFromBase58(signer.Address())
	if err != nil {
		return "", fmt.Errorf("invalid signer address: %w", err)
	}

	tokenMint, err := solana.PublicKeyFromBase58(token.Address)
	if err != nil {
		return "", fmt.Errorf("invalid token address: %w", err)
	}

	// Get token mint info to retrieve decimals
	mintInfo, err := t.client.GetAccountInfo(ctx, tokenMint)
	if err != nil {
		return "", fmt.Errorf("failed to get mint info: %w", err)
	}
	if mintInfo == nil || mintInfo.Value == nil {
		return "", fmt.Errorf("token mint not found")
	}

	// Build transaction
	recent, err := t.client.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	// Attest token instruction
	attestData := encodeAttestInstruction(tokenMint)
	attestIx := solana.NewInstruction(
		t.programID,
		solana.AccountMetaSlice{
			solana.Meta(signerPubkey).WRITE().SIGNER(),
			solana.Meta(tokenMint),
		},
		attestData,
	)

	tx, err := solana.NewTransaction(
		[]solana.Instruction{attestIx},
		recent.Value.Blockhash,
		solana.TransactionPayer(signerPubkey),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create transaction: %w", err)
	}

	signedTx, err := t.signTransaction(tx, signer)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	sig, err := t.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return wormholetypes.TxHash(sig.String()), nil
}

// CreateWrapped creates a wrapped token on Solana
func (t *SolanaTokenBridge) CreateWrapped(ctx context.Context, vaa []byte, signer wormholetypes.Signer) (wormholetypes.TxHash, error) {
	if len(vaa) == 0 {
		return "", fmt.Errorf("VAA cannot be empty")
	}

	if signer.Chain() != t.chain {
		return "", fmt.Errorf("signer chain %s does not match bridge chain %s", signer.Chain(), t.chain)
	}

	signerPubkey, err := solana.PublicKeyFromBase58(signer.Address())
	if err != nil {
		return "", fmt.Errorf("invalid signer address: %w", err)
	}

	// Parse VAA to get token info (simplified)
	tokenChain, tokenAddress := parseAttestationVAA(vaa)

	// Derive wrapped mint address
	wrappedMint, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("wrapped"),
			uint16ToBytes(tokenChain),
			tokenAddress[:],
		},
		t.programID,
	)
	if err != nil {
		return "", fmt.Errorf("failed to derive wrapped mint: %w", err)
	}

	// Check if already exists
	accountInfo, err := t.client.GetAccountInfo(ctx, wrappedMint)
	if err == nil && accountInfo != nil && accountInfo.Value != nil {
		return "", fmt.Errorf("wrapped token already exists: %s", wrappedMint.String())
	}

	// Build transaction
	recent, err := t.client.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	// Post VAA instruction
	postVAAData := append([]byte{0x01}, vaa...)
	postVAAIx := solana.NewInstruction(
		t.programID,
		solana.AccountMetaSlice{
			solana.Meta(signerPubkey).WRITE().SIGNER(),
		},
		postVAAData,
	)

	// Create wrapped instruction
	createData := append([]byte{0x03}, vaa...)
	createIx := solana.NewInstruction(
		t.programID,
		solana.AccountMetaSlice{
			solana.Meta(signerPubkey).WRITE().SIGNER(),
			solana.Meta(wrappedMint).WRITE(),
		},
		createData,
	)

	tx, err := solana.NewTransaction(
		[]solana.Instruction{postVAAIx, createIx},
		recent.Value.Blockhash,
		solana.TransactionPayer(signerPubkey),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create transaction: %w", err)
	}

	signedTx, err := t.signTransaction(tx, signer)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	sig, err := t.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return wormholetypes.TxHash(sig.String()), nil
}

// GetWrappedAsset returns the wrapped asset address for a token
func (t *SolanaTokenBridge) GetWrappedAsset(ctx context.Context, token wormholetypes.TokenID) (string, error) {
	if token.Chain == t.chain {
		return "", fmt.Errorf("token is native to this chain, not wrapped")
	}

	tokenAddr, err := wormholetypes.NewChainAddress(token.Chain, token.Address).ToUniversalAddress()
	if err != nil {
		return "", fmt.Errorf("failed to convert token address: %w", err)
	}

	chainID := token.Chain.GetChainID()

	// Derive wrapped mint PDA
	wrappedMint, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("wrapped"),
			uint16ToBytes(uint16(chainID)),
			tokenAddr[:],
		},
		t.programID,
	)
	if err != nil {
		return "", fmt.Errorf("failed to derive wrapped mint: %w", err)
	}

	// Check if account exists
	accountInfo, err := t.client.GetAccountInfo(ctx, wrappedMint)
	if err != nil || accountInfo == nil || accountInfo.Value == nil {
		return "", fmt.Errorf("wrapped asset does not exist")
	}

	return wrappedMint.String(), nil
}

// IsWrappedAsset checks if a token is a wrapped asset
func (t *SolanaTokenBridge) IsWrappedAsset(ctx context.Context, token wormholetypes.TokenID) (bool, error) {
	if token.Chain != t.chain {
		return false, fmt.Errorf("token is not on this chain")
	}

	tokenMint, err := solana.PublicKeyFromBase58(token.Address)
	if err != nil {
		return false, fmt.Errorf("invalid token address: %w", err)
	}

	// Get mint account
	mintInfo, err := t.client.GetAccountInfo(ctx, tokenMint)
	if err != nil || mintInfo == nil || mintInfo.Value == nil {
		return false, fmt.Errorf("token mint not found")
	}

	// Derive bridge authority
	bridgeAuthority, _, err := solana.FindProgramAddress(
		[][]byte{[]byte("mint_authority")},
		t.programID,
	)
	if err != nil {
		return false, fmt.Errorf("failed to derive bridge authority: %w", err)
	}

	// Check if metadata account exists
	metadataAccount, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("meta"),
			tokenMint[:],
		},
		t.programID,
	)
	if err != nil {
		return false, nil
	}

	metaInfo, err := t.client.GetAccountInfo(ctx, metadataAccount)
	if err != nil || metaInfo == nil || metaInfo.Value == nil {
		return false, nil
	}

	_ = bridgeAuthority
	return true, nil
}

func (t *SolanaTokenBridge) signTransaction(tx *solana.Transaction, signer wormholetypes.Signer) (*solana.Transaction, error) {
	// TODO: Implement transaction signing based on Signer interface
	return tx, nil
}

// Helper functions

func getTokenBridgeABI() *strings.Reader {
	// Simplified ABI for token bridge
	abi := `[
		{"name":"transferTokens","type":"function","inputs":[{"name":"token","type":"address"},{"name":"amount","type":"uint256"},{"name":"recipientChain","type":"uint16"},{"name":"recipient","type":"bytes32"},{"name":"arbiterFee","type":"uint256"},{"name":"nonce","type":"uint32"}]},
		{"name":"completeTransfer","type":"function","inputs":[{"name":"encodedVm","type":"bytes"}]},
		{"name":"attestToken","type":"function","inputs":[{"name":"tokenAddress","type":"address"},{"name":"nonce","type":"uint32"}]},
		{"name":"createWrapped","type":"function","inputs":[{"name":"encodedVm","type":"bytes"}]},
		{"name":"wrappedAsset","type":"function","inputs":[{"name":"tokenChainId","type":"uint16"},{"name":"tokenAddress","type":"bytes32"}],"outputs":[{"name":"","type":"address"}]},
		{"name":"isWrappedAsset","type":"function","inputs":[{"name":"token","type":"address"}],"outputs":[{"name":"","type":"bool"}]}
	]`
	return strings.NewReader(abi)
}

func getERC20ABI() *strings.Reader {
	abi := `[
		{"name":"approve","type":"function","inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[{"name":"","type":"bool"}]}
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

func encodeTransferInstruction(amount uint64, recipient [32]byte, targetChain uint16) []byte {
	// Instruction discriminator for transfer
	data := []byte{0x01}
	// Add amount (8 bytes, little-endian)
	amountBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		amountBytes[i] = byte(amount >> (i * 8))
	}
	data = append(data, amountBytes...)
	// Add recipient (32 bytes)
	data = append(data, recipient[:]...)
	// Add target chain (2 bytes, little-endian)
	data = append(data, byte(targetChain), byte(targetChain>>8))
	return data
}

func encodeAttestInstruction(mint solana.PublicKey) []byte {
	// Instruction discriminator for attest
	data := []byte{0x02}
	// Add nonce (4 bytes)
	data = append(data, 0, 0, 0, 0)
	return data
}

func parseAttestationVAA(vaa []byte) (uint16, [32]byte) {
	// Simplified VAA parsing - in reality this would be much more complex
	var tokenChain uint16
	var tokenAddress [32]byte

	if len(vaa) > 100 {
		tokenChain = uint16(vaa[50])<<8 | uint16(vaa[51])
		copy(tokenAddress[:], vaa[52:84])
	}

	return tokenChain, tokenAddress
}

func uint16ToBytes(n uint16) []byte {
	return []byte{byte(n >> 8), byte(n)}
}
