package transaction

import (
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/script"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// ValidateTransaction performs basic transaction validation
func ValidateTransaction(tx *types.Transaction) error {
	// Rule 1: Transaction must have at least one input and one output
	if len(tx.Inputs) == 0 {
		return fmt.Errorf("transaction has no inputs")
	}

	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction has no outputs")
	}

	// Rule 2: Check for duplicate inputs (double-spend within same tx)
	seen := make(map[string]bool)
	for _, input := range tx.Inputs {
		key := fmt.Sprintf("%x:%d", input.PrevTxHash, input.OutputIndex)
		if seen[key] {
			return fmt.Errorf("duplicate input: %s", key)
		}
		seen[key] = true
	}

	// Rule 3: All output values must be positive
	for i, output := range tx.Outputs {
		if output.Value < 0 {
			return fmt.Errorf("output %d has negative value: %d", i, output.Value)
		}

		// Rule 4: Check for output value overflow (> 21 million BTC)
		const maxMoney = 21000000 * 100000000 // 21 million BTC in satoshis
		if output.Value > maxMoney {
			return fmt.Errorf("output %d value exceeds maximum: %d", i, output.Value)
		}
	}

	// Rule 5: Sum of outputs must not overflow
	totalOut := int64(0)
	for _, output := range tx.Outputs {
		totalOut += output.Value
		if totalOut < 0 {
			return fmt.Errorf("total output value overflow")
		}
	}

	// Rule 6: Transaction size check (simplified)
	serialized, err := serialization.SerializeTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to serialize transaction: %w", err)
	}

	const maxTxSize = 100000 // 100KB
	if len(serialized) > maxTxSize {
		return fmt.Errorf("transaction too large: %d bytes", len(serialized))
	}

	return nil
}

// ValidateTransactionScripts validates all input scripts against previous outputs
func ValidateTransactionScripts(tx *types.Transaction, prevOutputs []types.TxOutput) error {
	if len(tx.Inputs) != len(prevOutputs) {
		return fmt.Errorf("input count (%d) doesn't match prevOutputs count (%d)",
			len(tx.Inputs), len(prevOutputs))
	}

	for i, input := range tx.Inputs {
		// Get the locking script from previous output
		lockingScript := prevOutputs[i].PubKeyScript

		// Get the unlocking script from this input
		unlockingScript := input.SignatureScript

		// Validate the script
		if err := validateScript(unlockingScript, lockingScript, tx, i); err != nil {
			return fmt.Errorf("input %d script validation failed: %w", i, err)
		}
	}

	return nil
}

// validateScript executes unlocking + locking script
func validateScript(unlocking, locking []byte, tx *types.Transaction, inputIdx int) error {
	// Combine scripts: unlocking + locking
	combined := append(unlocking, locking...)

	// Create engine with transaction context
	engine := script.NewEngine(combined)
	engine.SetTransaction(tx, inputIdx)

	// Execute
	if err := engine.Execute(); err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	return nil
}

// IsCoinbase checks if transaction is a coinbase (mining reward)
func IsCoinbase(tx *types.Transaction) bool {
	// Coinbase has exactly one input
	if len(tx.Inputs) != 1 {
		return false
	}

	// Previous transaction hash is all zeros
	input := tx.Inputs[0]
	for _, b := range input.PrevTxHash {
		if b != 0 {
			return false
		}
	}

	// Previous output index is 0xFFFFFFFF
	return input.OutputIndex == 0xFFFFFFFF
}

// ValidateCoinbase validates a coinbase transaction
func ValidateCoinbase(tx *types.Transaction, blockHeight uint64) error {
	if !IsCoinbase(tx) {
		return fmt.Errorf("not a coinbase transaction")
	}

	// Coinbase scriptSig must be 2-100 bytes
	sigScript := tx.Inputs[0].SignatureScript
	if len(sigScript) < 2 || len(sigScript) > 100 {
		return fmt.Errorf("coinbase scriptSig length invalid: %d", len(sigScript))
	}

	// For BIP34, first bytes should contain block height
	// (We'll implement this check later when we need it)

	// Coinbase must have at least one output
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("coinbase has no outputs")
	}

	return nil
}

// CalculateFee calculates transaction fee
func CalculateFee(tx *types.Transaction, prevOutputs []types.TxOutput) (int64, error) {
	if IsCoinbase(tx) {
		return 0, nil // Coinbase has no fee
	}

	if len(tx.Inputs) != len(prevOutputs) {
		return 0, fmt.Errorf("input/output count mismatch")
	}

	// Sum inputs
	totalIn := int64(0)
	for _, prevOut := range prevOutputs {
		totalIn += prevOut.Value
	}

	// Sum outputs
	totalOut := int64(0)
	for _, output := range tx.Outputs {
		totalOut += output.Value
	}

	// Fee = inputs - outputs
	fee := totalIn - totalOut

	if fee < 0 {
		return 0, fmt.Errorf("negative fee (outputs exceed inputs)")
	}

	return fee, nil
}

// CheckTransactionInputs validates transaction inputs against UTXO set
func CheckTransactionInputs(tx *types.Transaction, getUTXO func(types.Hash, uint32) (*types.TxOutput, error)) error {
	if IsCoinbase(tx) {
		return nil // Coinbase has no previous outputs
	}

	totalIn := int64(0)

	for i, input := range tx.Inputs {
		// Get the previous output being spent
		prevOut, err := getUTXO(input.PrevTxHash, input.OutputIndex)
		if err != nil {
			return fmt.Errorf("input %d: failed to get UTXO: %w", i, err)
		}

		if prevOut == nil {
			return fmt.Errorf("input %d: UTXO not found (already spent or doesn't exist)", i)
		}

		totalIn += prevOut.Value
	}

	// Calculate total outputs
	totalOut := int64(0)
	for _, output := range tx.Outputs {
		totalOut += output.Value
	}

	// Verify inputs >= outputs (with fee)
	if totalIn < totalOut {
		return fmt.Errorf("inputs (%d) less than outputs (%d)", totalIn, totalOut)
	}

	return nil
}
