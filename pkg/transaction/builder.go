package transaction

import (
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/keys"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/script"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// TxBuilder helps construct transactions
type TxBuilder struct {
	version  int32
	inputs   []types.TxInput
	outputs  []types.TxOutput
	lockTime uint32
}

// NewTxBuilder creates a new transaction builder
func NewTxBuilder() *TxBuilder {
	return &TxBuilder{
		version:  1,
		inputs:   make([]types.TxInput, 0),
		outputs:  make([]types.TxOutput, 0),
		lockTime: 0,
	}
}

// AddInput adds an input to the transaction
func (b *TxBuilder) AddInput(prevTxHash types.Hash, outputIndex uint32) *TxBuilder {
	input := types.TxInput{
		PrevTxHash:      prevTxHash,
		OutputIndex:     outputIndex,
		SignatureScript: nil, // Will be filled when signing
		Sequence:        0xFFFFFFFF,
	}

	b.inputs = append(b.inputs, input)
	return b
}

// AddOutput adds an output to the transaction
func (b *TxBuilder) AddOutput(value int64, scriptPubKey []byte) *TxBuilder {
	output := types.TxOutput{
		Value:        value,
		PubKeyScript: scriptPubKey,
	}

	b.outputs = append(b.outputs, output)
	return b
}

// AddP2PKHOutput adds a Pay-to-PubKey-Hash output
func (b *TxBuilder) AddP2PKHOutput(value int64, address string) (*TxBuilder, error) {
	// Decode address to get pubkey hash
	addr, err := keys.DecodeAddress(address)
	if err != nil {
		return b, fmt.Errorf("invalid address: %w", err)
	}

	// Create P2PKH script
	scriptPubKey, err := script.P2PKH(addr.Hash())
	if err != nil {
		return b, fmt.Errorf("failed to create P2PKH script: %w", err)
	}

	return b.AddOutput(value, scriptPubKey), nil
}

// SetLockTime sets the transaction lock time
func (b *TxBuilder) SetLockTime(lockTime uint32) *TxBuilder {
	b.lockTime = lockTime
	return b
}

// Build creates the unsigned transaction
func (b *TxBuilder) Build() (*types.Transaction, error) {
	if len(b.inputs) == 0 {
		return nil, fmt.Errorf("transaction must have at least one input")
	}

	if len(b.outputs) == 0 {
		return nil, fmt.Errorf("transaction must have at least one output")
	}

	tx := &types.Transaction{
		Version:  b.version,
		Inputs:   b.inputs,
		Outputs:  b.outputs,
		LockTime: b.lockTime,
	}

	return tx, nil
}

// SignInput signs a specific input
func SignInput(tx *types.Transaction, inputIdx int, privKey *keys.PrivateKey, prevScript []byte, hashType SigHashType) error {
	if inputIdx < 0 || inputIdx >= len(tx.Inputs) {
		return fmt.Errorf("invalid input index: %d", inputIdx)
	}

	// Calculate signature hash
	sigHash, err := CalcSignatureHash(tx, inputIdx, prevScript, hashType)
	if err != nil {
		return fmt.Errorf("failed to calculate signature hash: %w", err)
	}

	// Sign the hash
	signature, err := privKey.Sign(sigHash)
	if err != nil {
		return fmt.Errorf("failed to sign: %w", err)
	}

	// Append hash type to signature (DER + hash type byte)
	sigBytes := signature.Serialize()
	sigBytes = append(sigBytes, byte(hashType))

	// Get public key
	pubKey := privKey.PublicKey()
	pubKeyBytes := pubKey.Bytes(true)

	// Create unlocking script
	unlockingScript := script.P2PKHUnlockingScript(sigBytes, pubKeyBytes)

	// Set the signature script
	tx.Inputs[inputIdx].SignatureScript = unlockingScript

	return nil
}

// CreateCoinbase creates a coinbase transaction
func CreateCoinbase(blockHeight uint64, reward int64, address string, extraData []byte) (*types.Transaction, error) {
	// Decode address
	addr, err := keys.DecodeAddress(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Create coinbase input script (height + extra data)
	scriptSig := script.NewBuilder()
	scriptSig.AddInt(int64(blockHeight)) // BIP34: block height
	if len(extraData) > 0 {
		scriptSig.AddData(extraData)
	}

	// Create coinbase input
	input := types.TxInput{
		PrevTxHash:      types.Hash{}, // All zeros
		OutputIndex:     0xFFFFFFFF,
		SignatureScript: scriptSig.Script(),
		Sequence:        0xFFFFFFFF,
	}

	// Create output (reward goes to miner)
	scriptPubKey, err := script.P2PKH(addr.Hash())
	if err != nil {
		return nil, err
	}

	output := types.TxOutput{
		Value:        reward,
		PubKeyScript: scriptPubKey,
	}

	tx := &types.Transaction{
		Version:  1,
		Inputs:   []types.TxInput{input},
		Outputs:  []types.TxOutput{output},
		LockTime: 0,
	}

	return tx, nil
}

// CalculateSize estimates transaction size in bytes
func CalculateSize(numInputs, numOutputs int) int {
	// Version: 4 bytes
	size := 4

	// Input count: 1-9 bytes (VarInt)
	size += 1

	// Each input: ~148 bytes
	// (32 prev hash + 4 output index + ~107 sig script + 4 sequence)
	size += numInputs * 148

	// Output count: 1-9 bytes (VarInt)
	size += 1

	// Each output: ~34 bytes
	// (8 value + ~26 script)
	size += numOutputs * 34

	// Locktime: 4 bytes
	size += 4

	return size
}

// EstimateFee estimates transaction fee based on size and fee rate
func EstimateFee(numInputs, numOutputs int, feePerByte int64) int64 {
	size := CalculateSize(numInputs, numOutputs)
	return int64(size) * feePerByte
}
