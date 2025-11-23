package mempool

import (
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// FeeEstimator estimates transaction fees based on mempool state
type FeeEstimator struct {
	mempool *Mempool
}

// NewFeeEstimator creates a new fee estimator
func NewFeeEstimator(mempool *Mempool) *FeeEstimator {
	return &FeeEstimator{
		mempool: mempool,
	}
}

// CalculateTransactionSize calculates the size of a transaction in bytes
func CalculateTransactionSize(tx *types.Transaction) int64 {
	size := int64(0)

	// Version (4 bytes)
	size += 4

	// Input count (varint, assume 1 byte for simplicity)
	size += 1

	// Inputs
	for _, input := range tx.Inputs {
		// Previous transaction hash (32 bytes)
		size += 32
		// Output index (4 bytes)
		size += 4
		// Script length (varint, assume 1 byte)
		size += 1
		// Script
		size += int64(len(input.SignatureScript))
		// Sequence (4 bytes)
		size += 4
	}

	// Output count (varint, assume 1 byte)
	size += 1

	// Outputs
	for _, output := range tx.Outputs {
		// Value (8 bytes)
		size += 8
		// Script length (varint, assume 1 byte)
		size += 1
		// Script
		size += int64(len(output.PubKeyScript))
	}

	// Locktime (4 bytes)
	size += 4

	return size
}

// CalculateTransactionFee calculates the fee for a transaction
func CalculateTransactionFee(tx *types.Transaction, inputValues []int64) (int64, error) {
	if len(inputValues) != len(tx.Inputs) {
		return 0, fmt.Errorf("input values count mismatch")
	}

	// Sum input values
	totalInput := int64(0)
	for _, value := range inputValues {
		totalInput += value
	}

	// Sum output values
	totalOutput := int64(0)
	for _, output := range tx.Outputs {
		totalOutput += output.Value
	}

	// Fee = inputs - outputs
	fee := totalInput - totalOutput

	if fee < 0 {
		return 0, fmt.Errorf("negative fee: inputs=%d, outputs=%d", totalInput, totalOutput)
	}

	return fee, nil
}

// CalculateFeeRate calculates the fee rate (satoshis per byte)
func CalculateFeeRate(fee int64, size int64) int64 {
	if size == 0 {
		return 0
	}
	return fee / size
}

// EstimateFee estimates the required fee for a transaction to be included
// within a target number of blocks
func (fe *FeeEstimator) EstimateFee(targetBlocks int, txSize int64) int64 {
	fe.mempool.mu.RLock()
	defer fe.mempool.mu.RUnlock()

	if len(fe.mempool.entries) == 0 {
		// No transactions in mempool, use minimum fee rate
		return fe.mempool.minFeeRate * txSize
	}

	// Get all fee rates
	feeRates := make([]int64, 0, len(fe.mempool.entries))
	for _, entry := range fe.mempool.entries {
		feeRates = append(feeRates, entry.FeeRate)
	}

	// Sort fee rates (descending)
	for i := 0; i < len(feeRates)-1; i++ {
		for j := i + 1; j < len(feeRates); j++ {
			if feeRates[i] < feeRates[j] {
				feeRates[i], feeRates[j] = feeRates[j], feeRates[i]
			}
		}
	}

	// Estimate based on target blocks
	// For simplicity, use percentile based on target
	var percentile int
	switch {
	case targetBlocks <= 1:
		percentile = 10 // Top 10% for next block
	case targetBlocks <= 3:
		percentile = 25 // Top 25% for 3 blocks
	case targetBlocks <= 6:
		percentile = 50 // Median for 6 blocks
	default:
		percentile = 75 // 75th percentile for longer
	}

	index := (len(feeRates) * percentile) / 100
	if index >= len(feeRates) {
		index = len(feeRates) - 1
	}

	estimatedFeeRate := feeRates[index]

	// Ensure minimum fee rate
	if estimatedFeeRate < fe.mempool.minFeeRate {
		estimatedFeeRate = fe.mempool.minFeeRate
	}

	return estimatedFeeRate * txSize
}

// GetFeeStatistics returns fee statistics from the mempool
func (fe *FeeEstimator) GetFeeStatistics() *FeeStatistics {
	fe.mempool.mu.RLock()
	defer fe.mempool.mu.RUnlock()

	stats := &FeeStatistics{
		TxCount: len(fe.mempool.entries),
	}

	if stats.TxCount == 0 {
		return stats
	}

	// Collect fee rates
	feeRates := make([]int64, 0, len(fe.mempool.entries))
	totalFee := int64(0)
	totalSize := int64(0)

	for _, entry := range fe.mempool.entries {
		feeRates = append(feeRates, entry.FeeRate)
		totalFee += entry.Fee
		totalSize += entry.Size
	}

	// Sort fee rates
	for i := 0; i < len(feeRates)-1; i++ {
		for j := i + 1; j < len(feeRates); j++ {
			if feeRates[i] > feeRates[j] {
				feeRates[i], feeRates[j] = feeRates[j], feeRates[i]
			}
		}
	}

	// Calculate statistics
	stats.MinFeeRate = feeRates[0]
	stats.MaxFeeRate = feeRates[len(feeRates)-1]
	stats.MedianFeeRate = feeRates[len(feeRates)/2]
	stats.AverageFeeRate = totalFee / totalSize

	// Percentiles
	stats.P25FeeRate = feeRates[len(feeRates)/4]
	stats.P75FeeRate = feeRates[(len(feeRates)*3)/4]
	stats.P90FeeRate = feeRates[(len(feeRates)*9)/10]

	stats.TotalFees = totalFee
	stats.TotalSize = totalSize

	return stats
}

// FeeStatistics contains fee statistics from the mempool
type FeeStatistics struct {
	TxCount        int
	MinFeeRate     int64
	MaxFeeRate     int64
	MedianFeeRate  int64
	AverageFeeRate int64
	P25FeeRate     int64
	P75FeeRate     int64
	P90FeeRate     int64
	TotalFees      int64
	TotalSize      int64
}

// CalculateAncestorFee calculates the total fee including all ancestors
func (fe *FeeEstimator) CalculateAncestorFee(txHash types.Hash) (int64, error) {
	fe.mempool.mu.RLock()
	defer fe.mempool.mu.RUnlock()

	entry, exists := fe.mempool.entries[txHash]
	if !exists {
		return 0, fmt.Errorf("transaction not in mempool")
	}

	return entry.AncestorFee, nil
}

// CalculateAncestorFeeRate calculates the fee rate including all ancestors
func (fe *FeeEstimator) CalculateAncestorFeeRate(txHash types.Hash) (int64, error) {
	fe.mempool.mu.RLock()
	defer fe.mempool.mu.RUnlock()

	entry, exists := fe.mempool.entries[txHash]
	if !exists {
		return 0, fmt.Errorf("transaction not in mempool")
	}

	if entry.AncestorSize == 0 {
		return 0, nil
	}

	return entry.AncestorFee / entry.AncestorSize, nil
}

// GetDescendants returns all descendant transactions
func (fe *FeeEstimator) GetDescendants(txHash types.Hash) ([]types.Hash, error) {
	fe.mempool.mu.RLock()
	defer fe.mempool.mu.RUnlock()

	_, exists := fe.mempool.entries[txHash]
	if !exists {
		return nil, fmt.Errorf("transaction not in mempool")
	}

	descendants := make([]types.Hash, 0)
	visited := make(map[types.Hash]bool)

	var collectDescendants func(types.Hash)
	collectDescendants = func(hash types.Hash) {
		if visited[hash] {
			return
		}
		visited[hash] = true

		if e, ok := fe.mempool.entries[hash]; ok {
			for _, childHash := range e.Children {
				descendants = append(descendants, childHash)
				collectDescendants(childHash)
			}
		}
	}

	collectDescendants(txHash)

	return descendants, nil
}
