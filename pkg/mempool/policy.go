package mempool

import (
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// Policy defines mempool acceptance policies
type Policy struct {
	MinFeeRate         int64 // Minimum fee rate (satoshis/byte)
	MaxTxSize          int64 // Maximum transaction size
	MaxAncestorCount   int   // Maximum number of ancestors
	MaxAncestorSize    int64 // Maximum total size of ancestors
	MaxDescendantCount int   // Maximum number of descendants
	MaxDescendantSize  int64 // Maximum total size of descendants
	RequireStandard    bool  // Require standard transaction types
	AllowRBF           bool  // Allow Replace-By-Fee
	MaxSigOps          int   // Maximum signature operations
	DustThreshold      int64 // Minimum output value (dust threshold)
}

// DefaultPolicy returns the default mempool policy
func DefaultPolicy() *Policy {
	return &Policy{
		MinFeeRate:         1,      // 1 satoshi per byte
		MaxTxSize:          100000, // 100 KB
		MaxAncestorCount:   25,
		MaxAncestorSize:    101000, // 101 KB
		MaxDescendantCount: 25,
		MaxDescendantSize:  101000, // 101 KB
		RequireStandard:    true,
		AllowRBF:           true,
		MaxSigOps:          4000,
		DustThreshold:      546, // 546 satoshis (standard dust threshold)
	}
}

// PolicyValidator validates transactions against mempool policies
type PolicyValidator struct {
	policy  *Policy
	mempool *Mempool
}

// NewPolicyValidator creates a new policy validator
func NewPolicyValidator(policy *Policy, mempool *Mempool) *PolicyValidator {
	return &PolicyValidator{
		policy:  policy,
		mempool: mempool,
	}
}

// ValidateTransaction validates a transaction against mempool policies
func (pv *PolicyValidator) ValidateTransaction(tx *types.Transaction, fee int64) error {
	// Check transaction size
	size := CalculateTransactionSize(tx)
	if size > pv.policy.MaxTxSize {
		return fmt.Errorf("transaction too large: %d > %d", size, pv.policy.MaxTxSize)
	}

	// Check fee rate
	feeRate := CalculateFeeRate(fee, size)
	if feeRate < pv.policy.MinFeeRate {
		return fmt.Errorf("fee rate too low: %d < %d", feeRate, pv.policy.MinFeeRate)
	}

	// Check for dust outputs
	if err := pv.checkDustOutputs(tx); err != nil {
		return err
	}

	// Check standard transaction
	if pv.policy.RequireStandard {
		if err := pv.checkStandardTransaction(tx); err != nil {
			return err
		}
	}

	// Check signature operations
	if err := pv.checkSigOps(tx); err != nil {
		return err
	}

	return nil
}

// checkDustOutputs checks for dust outputs
func (pv *PolicyValidator) checkDustOutputs(tx *types.Transaction) error {
	for i, output := range tx.Outputs {
		if output.Value < pv.policy.DustThreshold {
			return fmt.Errorf("output %d is dust: %d < %d", i, output.Value, pv.policy.DustThreshold)
		}
	}
	return nil
}

// checkStandardTransaction checks if transaction is standard
func (pv *PolicyValidator) checkStandardTransaction(tx *types.Transaction) error {
	// Check version
	if tx.Version < 1 || tx.Version > 2 {
		return fmt.Errorf("non-standard version: %d", tx.Version)
	}

	// Check inputs
	if len(tx.Inputs) == 0 {
		return fmt.Errorf("no inputs")
	}

	// Check outputs
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("no outputs")
	}

	// Check for null data outputs (OP_RETURN)
	nullDataCount := 0
	for _, output := range tx.Outputs {
		if len(output.PubKeyScript) > 0 && output.PubKeyScript[0] == 0x6a { // OP_RETURN
			nullDataCount++
			if nullDataCount > 1 {
				return fmt.Errorf("multiple OP_RETURN outputs")
			}
			if len(output.PubKeyScript) > 83 {
				return fmt.Errorf("OP_RETURN output too large")
			}
		}
	}

	return nil
}

// checkSigOps checks signature operations count
func (pv *PolicyValidator) checkSigOps(tx *types.Transaction) error {
	// Simplified: count inputs as potential sig ops
	// In reality, we'd parse scripts to count actual sig ops
	sigOps := len(tx.Inputs) * 2 // Assume 2 sig ops per input (conservative)

	if sigOps > pv.policy.MaxSigOps {
		return fmt.Errorf("too many signature operations: %d > %d", sigOps, pv.policy.MaxSigOps)
	}

	return nil
}

// CheckAncestorLimits checks if adding a transaction would violate ancestor limits
func (pv *PolicyValidator) CheckAncestorLimits(tx *types.Transaction) error {
	pv.mempool.mu.RLock()
	defer pv.mempool.mu.RUnlock()

	// Count ancestors
	ancestorCount := 0
	ancestorSize := int64(0)

	for _, input := range tx.Inputs {
		// Skip coinbase
		if input.OutputIndex == 0xFFFFFFFF {
			continue
		}

		// Check if parent is in mempool
		if parent, exists := pv.mempool.entries[input.PrevTxHash]; exists {
			ancestorCount++
			ancestorSize += parent.Size

			// Add parent's ancestors
			ancestorCount += len(parent.Parents)
			ancestorSize += parent.AncestorSize - parent.Size
		}
	}

	// Check limits
	if ancestorCount > pv.policy.MaxAncestorCount {
		return fmt.Errorf("too many ancestors: %d > %d", ancestorCount, pv.policy.MaxAncestorCount)
	}

	if ancestorSize > pv.policy.MaxAncestorSize {
		return fmt.Errorf("ancestor size too large: %d > %d", ancestorSize, pv.policy.MaxAncestorSize)
	}

	return nil
}

// CheckDescendantLimits checks if adding a transaction would violate descendant limits
func (pv *PolicyValidator) CheckDescendantLimits(tx *types.Transaction) error {
	pv.mempool.mu.RLock()
	defer pv.mempool.mu.RUnlock()

	// For each parent in mempool, check if adding this tx would violate limits
	for _, input := range tx.Inputs {
		if parent, exists := pv.mempool.entries[input.PrevTxHash]; exists {
			descendantCount := len(parent.Children) + 1 // +1 for this tx
			descendantSize := CalculateTransactionSize(tx)

			// Count existing descendants
			for _, childHash := range parent.Children {
				if child, ok := pv.mempool.entries[childHash]; ok {
					descendantSize += child.Size
				}
			}

			if descendantCount > pv.policy.MaxDescendantCount {
				return fmt.Errorf("too many descendants: %d > %d", descendantCount, pv.policy.MaxDescendantCount)
			}

			if descendantSize > pv.policy.MaxDescendantSize {
				return fmt.Errorf("descendant size too large: %d > %d", descendantSize, pv.policy.MaxDescendantSize)
			}
		}
	}

	return nil
}

// IsRBFSignaled checks if a transaction signals Replace-By-Fee
func (pv *PolicyValidator) IsRBFSignaled(tx *types.Transaction) bool {
	if !pv.policy.AllowRBF {
		return false
	}

	// Check if any input has sequence < 0xfffffffe (BIP 125)
	for _, input := range tx.Inputs {
		if input.Sequence < 0xfffffffe {
			return true
		}
	}

	return false
}

// ValidateReplacement validates a replacement transaction (RBF)
func (pv *PolicyValidator) ValidateReplacement(newTx *types.Transaction, newFee int64, conflictingTxs []*MempoolEntry) error {
	if !pv.policy.AllowRBF {
		return fmt.Errorf("RBF not allowed")
	}

	// Calculate new transaction fee rate
	newSize := CalculateTransactionSize(newTx)
	newFeeRate := CalculateFeeRate(newFee, newSize)

	// Check each conflicting transaction
	for _, conflicting := range conflictingTxs {
		// Rule 1: New transaction must pay higher absolute fee
		if newFee <= conflicting.Fee {
			return fmt.Errorf("new fee not higher than existing: %d <= %d", newFee, conflicting.Fee)
		}

		// Rule 2: New transaction must have higher fee rate
		if newFeeRate <= conflicting.FeeRate {
			return fmt.Errorf("new fee rate not higher: %d <= %d", newFeeRate, conflicting.FeeRate)
		}

		// Rule 3: Additional fee must cover bandwidth cost
		additionalFee := newFee - conflicting.Fee
		minAdditionalFee := pv.policy.MinFeeRate * conflicting.Size

		if additionalFee < minAdditionalFee {
			return fmt.Errorf("additional fee too low: %d < %d", additionalFee, minAdditionalFee)
		}
	}

	return nil
}

// GetPolicyInfo returns information about the current policy
func (pv *PolicyValidator) GetPolicyInfo() map[string]interface{} {
	return map[string]interface{}{
		"min_fee_rate":         pv.policy.MinFeeRate,
		"max_tx_size":          pv.policy.MaxTxSize,
		"max_ancestor_count":   pv.policy.MaxAncestorCount,
		"max_ancestor_size":    pv.policy.MaxAncestorSize,
		"max_descendant_count": pv.policy.MaxDescendantCount,
		"max_descendant_size":  pv.policy.MaxDescendantSize,
		"require_standard":     pv.policy.RequireStandard,
		"allow_rbf":            pv.policy.AllowRBF,
		"max_sig_ops":          pv.policy.MaxSigOps,
		"dust_threshold":       pv.policy.DustThreshold,
	}
}
