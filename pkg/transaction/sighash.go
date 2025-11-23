package transaction

import (
	"crypto/sha256"
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// SigHashType defines signature hash type
type SigHashType uint32

const (
	// SigHashAll signs all inputs and outputs
	SigHashAll SigHashType = 0x01

	// SigHashNone signs all inputs, no outputs
	SigHashNone SigHashType = 0x02

	// SigHashSingle signs all inputs, one output
	SigHashSingle SigHashType = 0x03

	// SigHashAnyOneCanPay modifier - only sign one input
	SigHashAnyOneCanPay SigHashType = 0x80
)

// CalcSignatureHash computes the signature hash for a transaction input
// This is what gets signed by the private key
func CalcSignatureHash(tx *types.Transaction, inputIdx int, subscript []byte, hashType SigHashType) ([]byte, error) {
	if inputIdx < 0 || inputIdx >= len(tx.Inputs) {
		return nil, fmt.Errorf("invalid input index: %d", inputIdx)
	}

	// Create a copy of the transaction
	txCopy := copyTransaction(tx)

	// Clear all input scripts
	for i := range txCopy.Inputs {
		txCopy.Inputs[i].SignatureScript = nil
	}

	// Set the subscript (previous output's scriptPubKey) for the input being signed
	txCopy.Inputs[inputIdx].SignatureScript = subscript

	// Apply signature hash type modifications
	baseType := hashType & 0x1f

	switch baseType {
	case SigHashAll:
		// Sign all inputs and all outputs (default)
		// Nothing to modify

	case SigHashNone:
		// Sign all inputs, but no outputs
		txCopy.Outputs = nil

		// Set sequence of other inputs to 0
		for i := range txCopy.Inputs {
			if i != inputIdx {
				txCopy.Inputs[i].Sequence = 0
			}
		}

	case SigHashSingle:
		// Sign all inputs and one output (at same index)
		if inputIdx >= len(txCopy.Outputs) {
			return nil, fmt.Errorf("SigHashSingle: input index exceeds output count")
		}

		// Keep only the output at inputIdx
		txCopy.Outputs = txCopy.Outputs[inputIdx : inputIdx+1]

		// Set sequence of other inputs to 0
		for i := range txCopy.Inputs {
			if i != inputIdx {
				txCopy.Inputs[i].Sequence = 0
			}
		}

	default:
		return nil, fmt.Errorf("unsupported signature hash type: %d", hashType)
	}

	// Handle ANYONECANPAY flag
	if hashType&SigHashAnyOneCanPay != 0 {
		// Only include the input being signed
		txCopy.Inputs = []types.TxInput{txCopy.Inputs[inputIdx]}
	}

	// Serialize the modified transaction
	serialized, err := serialization.SerializeTransaction(txCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Append hash type (4 bytes, little-endian)
	hashTypeBytes := make([]byte, 4)
	hashTypeBytes[0] = byte(hashType)
	hashTypeBytes[1] = byte(hashType >> 8)
	hashTypeBytes[2] = byte(hashType >> 16)
	hashTypeBytes[3] = byte(hashType >> 24)

	serialized = append(serialized, hashTypeBytes...)

	// Double SHA-256
	first := sha256.Sum256(serialized)
	second := sha256.Sum256(first[:])

	return second[:], nil
}

// copyTransaction creates a deep copy of a transaction
func copyTransaction(tx *types.Transaction) *types.Transaction {
	txCopy := &types.Transaction{
		Version:  tx.Version,
		LockTime: tx.LockTime,
		Inputs:   make([]types.TxInput, len(tx.Inputs)),
		Outputs:  make([]types.TxOutput, len(tx.Outputs)),
	}

	// Copy inputs
	for i, input := range tx.Inputs {
		txCopy.Inputs[i] = types.TxInput{
			PrevTxHash:  input.PrevTxHash,
			OutputIndex: input.OutputIndex,
			Sequence:    input.Sequence,
		}

		// Deep copy script
		if input.SignatureScript != nil {
			txCopy.Inputs[i].SignatureScript = make([]byte, len(input.SignatureScript))
			copy(txCopy.Inputs[i].SignatureScript, input.SignatureScript)
		}
	}

	// Copy outputs
	for i, output := range tx.Outputs {
		txCopy.Outputs[i] = types.TxOutput{
			Value: output.Value,
		}

		// Deep copy script
		if output.PubKeyScript != nil {
			txCopy.Outputs[i].PubKeyScript = make([]byte, len(output.PubKeyScript))
			copy(txCopy.Outputs[i].PubKeyScript, output.PubKeyScript)
		}
	}

	return txCopy
}

// SignatureHashInfo returns human-readable info about signature hash type
func SignatureHashInfo(hashType SigHashType) string {
	baseType := hashType & 0x1f
	anyoneCanPay := hashType&SigHashAnyOneCanPay != 0

	var info string

	switch baseType {
	case SigHashAll:
		info = "ALL"
	case SigHashNone:
		info = "NONE"
	case SigHashSingle:
		info = "SINGLE"
	default:
		info = fmt.Sprintf("UNKNOWN(%d)", baseType)
	}

	if anyoneCanPay {
		info += "|ANYONECANPAY"
	}

	return info
}
