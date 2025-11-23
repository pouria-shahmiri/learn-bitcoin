package validation

import (
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/crypto"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/transaction"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/utxo"
)

// BlockValidator validates blocks
type BlockValidator struct {
	utxoSet *utxo.UTXOSet
}

// NewBlockValidator creates a new block validator
func NewBlockValidator(utxoSet *utxo.UTXOSet) *BlockValidator {
	return &BlockValidator{
		utxoSet: utxoSet,
	}
}

// ValidateBlock performs full block validation
func (bv *BlockValidator) ValidateBlock(block *types.Block, height uint64, prevBlockHash types.Hash) error {
	// 1. Validate block header
	if err := bv.validateBlockHeader(&block.Header, prevBlockHash); err != nil {
		return fmt.Errorf("invalid block header: %w", err)
	}

	// 2. Check block size
	blockSize, err := serialization.SerializeBlock(block)
	if err != nil {
		return err
	}
	if len(blockSize) > MaxBlockSize {
		return fmt.Errorf("block too large: %d > %d", len(blockSize), MaxBlockSize)
	}

	// 3. Validate transactions
	if len(block.Transactions) == 0 {
		return fmt.Errorf("block has no transactions")
	}

	// 4. First transaction must be coinbase
	if !transaction.IsCoinbase(&block.Transactions[0]) {
		return fmt.Errorf("first transaction is not coinbase")
	}

	// 5. Only first transaction can be coinbase
	for i := 1; i < len(block.Transactions); i++ {
		if transaction.IsCoinbase(&block.Transactions[i]) {
			return fmt.Errorf("coinbase transaction at index %d (must be first)", i)
		}
	}

	// 6. Validate coinbase
	if err := transaction.ValidateCoinbase(&block.Transactions[0], height); err != nil {
		return fmt.Errorf("invalid coinbase: %w", err)
	}

	// 7. Validate all transactions
	totalFees := int64(0)
	for i, tx := range block.Transactions {
		if i == 0 {
			continue // Skip coinbase
		}

		// Basic validation
		if err := transaction.ValidateTransaction(&tx); err != nil {
			return fmt.Errorf("transaction %d invalid: %w", i, err)
		}

		// Check inputs against UTXO set
		fee, err := bv.validateTransactionInputs(&tx)
		if err != nil {
			return fmt.Errorf("transaction %d inputs invalid: %w", i, err)
		}

		totalFees += fee
	}

	// 8. Validate coinbase reward
	coinbaseValue := block.Transactions[0].Outputs[0].Value
	if err := ValidateBlockReward(coinbaseValue, totalFees, height); err != nil {
		return fmt.Errorf("invalid block reward: %w", err)
	}

	// 9. Verify merkle root
	var txHashes []types.Hash
	for _, tx := range block.Transactions {
		txHash, err := serialization.HashTransaction(&tx)
		if err != nil {
			return err
		}
		txHashes = append(txHashes, txHash)
	}

	calculatedMerkleRoot := crypto.ComputeMerkleRoot(txHashes)
	if calculatedMerkleRoot != block.Header.MerkleRoot {
		return fmt.Errorf("merkle root mismatch: expected %s, got %s",
			block.Header.MerkleRoot, calculatedMerkleRoot)
	}

	// 10. Check for duplicate transactions
	seen := make(map[types.Hash]bool)
	for _, tx := range block.Transactions {
		txHash, _ := serialization.HashTransaction(&tx)
		if seen[txHash] {
			return fmt.Errorf("duplicate transaction: %s", txHash)
		}
		seen[txHash] = true
	}

	return nil
}

// validateBlockHeader validates block header
func (bv *BlockValidator) validateBlockHeader(header *types.BlockHeader, prevBlockHash types.Hash) error {
	// 1. Check previous block hash
	if header.PrevBlockHash != prevBlockHash {
		return fmt.Errorf("previous block hash mismatch")
	}

	// 2. Check proof of work
	blockHash, err := serialization.HashBlockHeader(header)
	if err != nil {
		return err
	}

	if !IsValidProofOfWork(blockHash[:], header.Bits) {
		return fmt.Errorf("insufficient proof of work")
	}

	// 3. Check timestamp (simplified - should not be too far in future)
	// In real Bitcoin, this is more complex

	return nil
}

// validateTransactionInputs validates transaction inputs against UTXO set
func (bv *BlockValidator) validateTransactionInputs(tx *types.Transaction) (int64, error) {
	totalIn := int64(0)

	for i, input := range tx.Inputs {
		// Get the UTXO being spent
		outpoint := utxo.NewOutPoint(input.PrevTxHash, input.OutputIndex)

		spentUTXO, err := bv.utxoSet.Get(outpoint)
		if err != nil {
			return 0, fmt.Errorf("input %d: UTXO not found: %s", i, outpoint)
		}

		// Check if UTXO is mature (for coinbase)
		// Note: We'd need current height for this - simplified for now

		// Validate script
		if err := bv.validateInputScript(&input, &spentUTXO.Output, tx, i); err != nil {
			return 0, fmt.Errorf("input %d: script validation failed: %w", i, err)
		}

		totalIn += spentUTXO.Value()
	}

	// Calculate total outputs
	totalOut := int64(0)
	for _, output := range tx.Outputs {
		totalOut += output.Value

		// Check money range
		if err := CheckMoneyRange(output.Value); err != nil {
			return 0, err
		}
	}

	// Calculate fee
	fee := totalIn - totalOut
	if fee < 0 {
		return 0, fmt.Errorf("outputs exceed inputs")
	}

	return fee, nil
}

// validateInputScript validates input script against output script
func (bv *BlockValidator) validateInputScript(input *types.TxInput, prevOutput *types.TxOutput, tx *types.Transaction, inputIdx int) error {
	// For now, we'll do a simplified validation
	// Full implementation would execute the script

	// Combine unlocking and locking scripts
	_ = append(input.SignatureScript, prevOutput.PubKeyScript...)

	// TODO: Execute script with transaction context
	// This would use the script engine from Milestone 4

	return nil
}

// ApplyBlock applies a validated block to the UTXO set
func (bv *BlockValidator) ApplyBlock(block *types.Block, height uint64) error {
	// Apply each transaction
	for i, tx := range block.Transactions {
		txHash, err := serialization.HashTransaction(&tx)
		if err != nil {
			return err
		}

		isCoinbase := (i == 0)

		if err := bv.utxoSet.ApplyTransaction(&tx, txHash, height, isCoinbase); err != nil {
			return fmt.Errorf("failed to apply transaction %d: %w", i, err)
		}
	}

	return nil
}

// RevertBlock reverts a block from the UTXO set
func (bv *BlockValidator) RevertBlock(block *types.Block) error {
	// Revert transactions in reverse order
	for i := len(block.Transactions) - 1; i >= 0; i-- {
		tx := block.Transactions[i]
		txHash, err := serialization.HashTransaction(&tx)
		if err != nil {
			return err
		}

		if err := bv.utxoSet.RevertTransaction(&tx, txHash); err != nil {
			return fmt.Errorf("failed to revert transaction %d: %w", i, err)
		}
	}

	return nil
}
