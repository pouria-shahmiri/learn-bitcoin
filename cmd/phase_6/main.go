package main

import (
	"fmt"
	"time"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/crypto"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/keys"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/mempool"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/script"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

func main() {
	fmt.Println("=== Bitcoin Learning - Milestone 6 ===")
	fmt.Println("Mempool & Fee Policy\n")

	// Demo 1: Basic mempool operations
	demoBasicMempool()

	// Demo 2: Fee calculation and estimation
	demoFeeCalculation()

	// Demo 3: Transaction dependencies
	demoTransactionDependencies()

	// Demo 4: Replace-By-Fee (RBF)
	demoReplaceByFee()

	// Demo 5: Mempool policies
	demoMempoolPolicies()

	// Demo 6: Transaction selection for blocks
	demoTransactionSelection()

	// Demo 7: High load scenario
	demoHighLoad()

	// Demo 8: Transaction expiration
	demoTransactionExpiration()

	fmt.Println("\n=== All demos completed successfully! ===")
}

func demoBasicMempool() {
	fmt.Println("--- Demo 1: Basic Mempool Operations ---")

	// Create mempool with 10 MB limit, 1 sat/byte min fee, 24 hour expiration
	mp := mempool.NewMempool(10*1024*1024, 1, 24*3600)

	fmt.Printf("Mempool created:\n")
	fmt.Printf("  Max size: 10 MB\n")
	fmt.Printf("  Min fee rate: 1 sat/byte\n")
	fmt.Printf("  Max age: 24 hours\n\n")

	// Create some transactions
	for i := 0; i < 5; i++ {
		tx := createSampleTransaction(i, 1)
		fee := int64((i + 1) * 10000) // Increasing fees

		err := mp.Add(tx, fee, 100)
		if err != nil {
			fmt.Printf("  ✗ Failed to add tx %d: %v\n", i, err)
		} else {
			txHash, _ := serialization.HashTransaction(tx)
			fmt.Printf("  ✓ Added tx %d: %s... (fee: %d sats)\n", i, txHash.String()[:16], fee)
		}
	}

	fmt.Printf("\nMempool state:\n")
	fmt.Printf("  Transactions: %d\n", mp.Size())
	fmt.Printf("  Memory usage: %d bytes\n", mp.GetMemoryUsage())

	// Get a specific transaction
	txHash, _ := serialization.HashTransaction(createSampleTransaction(0, 1))
	entry, err := mp.Get(txHash)
	if err == nil {
		fmt.Printf("\nRetrieved transaction:\n")
		fmt.Printf("  Hash: %s...\n", entry.TxHash.String()[:16])
		fmt.Printf("  Size: %d bytes\n", entry.Size)
		fmt.Printf("  Fee: %d sats\n", entry.Fee)
		fmt.Printf("  Fee rate: %d sat/byte\n", entry.FeeRate)
	}

	fmt.Println()
}

func demoFeeCalculation() {
	fmt.Println("--- Demo 2: Fee Calculation and Estimation ---")

	mp := mempool.NewMempool(10*1024*1024, 1, 24*3600)
	estimator := mempool.NewFeeEstimator(mp)

	// Add transactions with varying fee rates
	feeRates := []int64{1, 5, 10, 20, 50, 100}
	fmt.Println("Adding transactions with different fee rates:")

	for i, feeRate := range feeRates {
		tx := createSampleTransaction(i, 1)
		size := mempool.CalculateTransactionSize(tx)
		fee := feeRate * size

		err := mp.Add(tx, fee, 100)
		if err == nil {
			fmt.Printf("  ✓ Tx %d: %d sat/byte (fee: %d sats)\n", i, feeRate, fee)
		}
	}

	// Get fee statistics
	stats := estimator.GetFeeStatistics()
	fmt.Printf("\nFee Statistics:\n")
	fmt.Printf("  Total transactions: %d\n", stats.TxCount)
	fmt.Printf("  Min fee rate: %d sat/byte\n", stats.MinFeeRate)
	fmt.Printf("  Max fee rate: %d sat/byte\n", stats.MaxFeeRate)
	fmt.Printf("  Median fee rate: %d sat/byte\n", stats.MedianFeeRate)
	fmt.Printf("  Average fee rate: %d sat/byte\n", stats.AverageFeeRate)
	fmt.Printf("  25th percentile: %d sat/byte\n", stats.P25FeeRate)
	fmt.Printf("  75th percentile: %d sat/byte\n", stats.P75FeeRate)
	fmt.Printf("  90th percentile: %d sat/byte\n", stats.P90FeeRate)

	// Estimate fees for different confirmation targets
	fmt.Printf("\nFee Estimation (for 250-byte transaction):\n")
	targets := []int{1, 3, 6, 12}
	for _, target := range targets {
		estimatedFee := estimator.EstimateFee(target, 250)
		estimatedRate := estimatedFee / 250
		fmt.Printf("  %d blocks: %d sats (%d sat/byte)\n", target, estimatedFee, estimatedRate)
	}

	fmt.Println()
}

func demoTransactionDependencies() {
	fmt.Println("--- Demo 3: Transaction Dependencies ---")

	mp := mempool.NewMempool(10*1024*1024, 1, 24*3600)

	// Create parent transaction
	parentTx := createSampleTransaction(0, 1)
	parentHash, _ := serialization.HashTransaction(parentTx)
	parentFee := int64(50000)

	err := mp.Add(parentTx, parentFee, 100)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Parent transaction added:\n")
	fmt.Printf("  Hash: %s...\n", parentHash.String()[:16])
	fmt.Printf("  Fee: %d sats\n", parentFee)

	// Create child transaction that spends parent's output
	childTx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      parentHash,
				OutputIndex:     0,
				SignatureScript: []byte("child_signature"),
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        90000000,
				PubKeyScript: []byte("child_output"),
			},
		},
		LockTime: 0,
	}

	childFee := int64(30000)
	err = mp.Add(childTx, childFee, 100)
	if err != nil {
		panic(err)
	}

	childHash, _ := serialization.HashTransaction(childTx)
	fmt.Printf("\nChild transaction added:\n")
	fmt.Printf("  Hash: %s...\n", childHash.String()[:16])
	fmt.Printf("  Fee: %d sats\n", childFee)

	// Check parent-child relationship
	childEntry, _ := mp.Get(childHash)
	fmt.Printf("\nDependency information:\n")
	fmt.Printf("  Child has %d parent(s)\n", len(childEntry.Parents))
	fmt.Printf("  Ancestor fee: %d sats\n", childEntry.AncestorFee)
	fmt.Printf("  Ancestor size: %d bytes\n", childEntry.AncestorSize)

	parentEntry, _ := mp.Get(parentHash)
	fmt.Printf("  Parent has %d child(ren)\n", len(parentEntry.Children))

	// Calculate package fee rate
	packageFeeRate := childEntry.AncestorFee / childEntry.AncestorSize
	fmt.Printf("  Package fee rate: %d sat/byte\n", packageFeeRate)

	fmt.Println()
}

func demoReplaceByFee() {
	fmt.Println("--- Demo 4: Replace-By-Fee (RBF) ---")

	mp := mempool.NewMempool(10*1024*1024, 1, 24*3600)

	// Create original transaction with RBF signal (sequence < 0xfffffffe)
	originalTx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      crypto.DoubleSHA256([]byte("utxo_1")),
				OutputIndex:     0,
				SignatureScript: []byte("signature_v1"),
				Sequence:        0xFFFFFFFD, // RBF signal
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        95000000,
				PubKeyScript: []byte("recipient"),
			},
		},
		LockTime: 0,
	}

	originalFee := int64(5000)
	err := mp.Add(originalTx, originalFee, 100)
	if err != nil {
		panic(err)
	}

	originalHash, _ := serialization.HashTransaction(originalTx)
	originalSize := mempool.CalculateTransactionSize(originalTx)
	originalFeeRate := originalFee / originalSize

	fmt.Printf("Original transaction:\n")
	fmt.Printf("  Hash: %s...\n", originalHash.String()[:16])
	fmt.Printf("  Fee: %d sats (%d sat/byte)\n", originalFee, originalFeeRate)
	fmt.Printf("  Sequence: 0x%x (RBF enabled)\n", originalTx.Inputs[0].Sequence)

	// Create replacement transaction with higher fee
	replacementTx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      crypto.DoubleSHA256([]byte("utxo_1")), // Same input
				OutputIndex:     0,
				SignatureScript: []byte("signature_v2"),
				Sequence:        0xFFFFFFFD,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        94000000, // Lower output (higher fee)
				PubKeyScript: []byte("recipient"),
			},
		},
		LockTime: 0,
	}

	replacementFee := int64(15000) // Higher fee
	err = mp.Add(replacementTx, replacementFee, 100)

	replacementHash, _ := serialization.HashTransaction(replacementTx)
	replacementSize := mempool.CalculateTransactionSize(replacementTx)
	replacementFeeRate := replacementFee / replacementSize

	if err == nil {
		fmt.Printf("\n✓ Replacement successful:\n")
		fmt.Printf("  New hash: %s...\n", replacementHash.String()[:16])
		fmt.Printf("  New fee: %d sats (%d sat/byte)\n", replacementFee, replacementFeeRate)
		fmt.Printf("  Fee increase: %d sats\n", replacementFee-originalFee)

		// Verify original is removed
		if !mp.Exists(originalHash) {
			fmt.Printf("  ✓ Original transaction removed\n")
		}
	} else {
		fmt.Printf("\n✗ Replacement failed: %v\n", err)
	}

	fmt.Println()
}

func demoMempoolPolicies() {
	fmt.Println("--- Demo 5: Mempool Policies ---")

	mp := mempool.NewMempool(10*1024*1024, 10, 24*3600) // 10 sat/byte minimum
	policy := mempool.DefaultPolicy()
	policy.MinFeeRate = 10
	validator := mempool.NewPolicyValidator(policy, mp)

	fmt.Println("Policy configuration:")
	policyInfo := validator.GetPolicyInfo()
	for key, value := range policyInfo {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Test 1: Transaction with low fee (should fail)
	fmt.Println("\nTest 1: Low fee transaction")
	lowFeeTx := createSampleTransaction(0, 1)
	lowFee := int64(100) // Very low fee
	err := validator.ValidateTransaction(lowFeeTx, lowFee)
	if err != nil {
		fmt.Printf("  ✗ Rejected: %v\n", err)
	} else {
		fmt.Printf("  ✓ Accepted\n")
	}

	// Test 2: Transaction with dust output (should fail)
	fmt.Println("\nTest 2: Dust output transaction")
	dustTx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      crypto.DoubleSHA256([]byte("input")),
				OutputIndex:     0,
				SignatureScript: []byte("sig"),
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        100, // Dust (< 546 sats)
				PubKeyScript: []byte("output"),
			},
		},
		LockTime: 0,
	}
	dustFee := int64(10000)
	err = validator.ValidateTransaction(dustTx, dustFee)
	if err != nil {
		fmt.Printf("  ✗ Rejected: %v\n", err)
	} else {
		fmt.Printf("  ✓ Accepted\n")
	}

	// Test 3: Valid transaction (should pass)
	fmt.Println("\nTest 3: Valid transaction")
	validTx := createSampleTransaction(1, 1)
	validSize := mempool.CalculateTransactionSize(validTx)
	validFee := policy.MinFeeRate * validSize
	err = validator.ValidateTransaction(validTx, validFee)
	if err != nil {
		fmt.Printf("  ✗ Rejected: %v\n", err)
	} else {
		fmt.Printf("  ✓ Accepted (fee: %d sats, %d sat/byte)\n", validFee, policy.MinFeeRate)
	}

	fmt.Println()
}

func demoTransactionSelection() {
	fmt.Println("--- Demo 6: Transaction Selection for Blocks ---")

	mp := mempool.NewMempool(10*1024*1024, 1, 24*3600)

	// Add transactions with different fee rates
	fmt.Println("Adding transactions to mempool:")
	feeRates := []int64{100, 50, 75, 25, 150, 10, 200, 5, 125, 80}

	for i, feeRate := range feeRates {
		tx := createSampleTransaction(i, 1)
		size := mempool.CalculateTransactionSize(tx)
		fee := feeRate * size

		mp.Add(tx, fee, 100)
		fmt.Printf("  Tx %d: %d sat/byte\n", i, feeRate)
	}

	// Create priority queue
	pq := mempool.NewPriorityQueue(mp)

	// Get top transactions
	fmt.Printf("\nTop 5 transactions by fee rate:\n")
	topTxs := pq.GetTopTransactions(5)
	for i, entry := range topTxs {
		fmt.Printf("  %d. Fee rate: %d sat/byte, Fee: %d sats\n",
			i+1, entry.FeeRate, entry.Fee)
	}

	// Select transactions for a block (1 MB limit)
	maxBlockSize := int64(1024 * 1024)
	fmt.Printf("\nSelecting transactions for block (max %d bytes):\n", maxBlockSize)

	selected, err := pq.SelectTransactions(maxBlockSize)
	if err != nil {
		panic(err)
	}

	totalSize := int64(0)
	totalFees := int64(0)

	for i, tx := range selected {
		size := mempool.CalculateTransactionSize(tx)
		totalSize += size

		// Find fee
		for _, entry := range mp.GetAllTransactions() {
			if entry.Tx == tx {
				totalFees += entry.Fee
				fmt.Printf("  %d. Size: %d bytes, Fee: %d sats (%d sat/byte)\n",
					i+1, size, entry.Fee, entry.FeeRate)
				break
			}
		}
	}

	fmt.Printf("\nBlock summary:\n")
	fmt.Printf("  Transactions: %d\n", len(selected))
	fmt.Printf("  Total size: %d bytes (%.2f%% full)\n",
		totalSize, float64(totalSize)/float64(maxBlockSize)*100)
	fmt.Printf("  Total fees: %d sats (%.8f BTC)\n",
		totalFees, float64(totalFees)/100000000)

	fmt.Println()
}

func demoHighLoad() {
	fmt.Println("--- Demo 7: High Load Scenario ---")

	// Create mempool with 1 MB limit
	mp := mempool.NewMempool(1024*1024, 1, 24*3600)

	fmt.Printf("Mempool limit: 1 MB\n")
	fmt.Println("Adding 100 transactions...")

	added := 0
	evicted := 0

	for i := 0; i < 100; i++ {
		tx := createSampleTransaction(i, 2) // Larger transactions
		size := mempool.CalculateTransactionSize(tx)

		// Random fee rate between 1-100 sat/byte
		feeRate := int64((i % 100) + 1)
		fee := feeRate * size

		err := mp.Add(tx, fee, 100)
		if err != nil {
			if i > 50 { // Only count evictions after mempool is partially full
				evicted++
			}
		} else {
			added++
		}

		if (i+1)%20 == 0 {
			fmt.Printf("  Progress: %d txs, mempool: %d txs, %.2f KB\n",
				i+1, mp.Size(), float64(mp.GetMemoryUsage())/1024)
		}
	}

	fmt.Printf("\nResults:\n")
	fmt.Printf("  Added: %d transactions\n", added)
	fmt.Printf("  Evicted: %d transactions\n", evicted)
	fmt.Printf("  Final mempool size: %d transactions\n", mp.Size())
	fmt.Printf("  Memory usage: %.2f KB / 1024 KB\n",
		float64(mp.GetMemoryUsage())/1024)

	// Get fee statistics after high load
	estimator := mempool.NewFeeEstimator(mp)
	stats := estimator.GetFeeStatistics()

	fmt.Printf("\nFee statistics after high load:\n")
	fmt.Printf("  Min fee rate: %d sat/byte\n", stats.MinFeeRate)
	fmt.Printf("  Median fee rate: %d sat/byte\n", stats.MedianFeeRate)
	fmt.Printf("  Max fee rate: %d sat/byte\n", stats.MaxFeeRate)

	fmt.Println()
}

func demoTransactionExpiration() {
	fmt.Println("--- Demo 8: Transaction Expiration ---")

	// Create mempool with 5 second expiration for demo
	mp := mempool.NewMempool(10*1024*1024, 1, 5)

	fmt.Println("Adding transactions...")

	// Add some transactions
	for i := 0; i < 5; i++ {
		tx := createSampleTransaction(i, 1)
		fee := int64(10000)
		mp.Add(tx, fee, 100)
	}

	fmt.Printf("  Initial mempool size: %d transactions\n", mp.Size())

	// Wait for expiration
	fmt.Println("\nWaiting 6 seconds for expiration...")
	time.Sleep(6 * time.Second)

	// Expire old transactions
	expired := mp.ExpireTransactions()

	fmt.Printf("  Expired: %d transactions\n", expired)
	fmt.Printf("  Remaining: %d transactions\n", mp.Size())

	// Add new transactions
	fmt.Println("\nAdding new transactions...")
	for i := 5; i < 8; i++ {
		tx := createSampleTransaction(i, 1)
		fee := int64(10000)
		mp.Add(tx, fee, 100)
	}

	fmt.Printf("  Mempool size: %d transactions\n", mp.Size())
	fmt.Printf("  ✓ Old transactions expired, new ones active\n")

	fmt.Println()
}

// Helper: Create a sample transaction
func createSampleTransaction(seed int, outputCount int) *types.Transaction {
	// Generate keys
	key, _ := keys.GeneratePrivateKey()
	pubKeyHash := key.PublicKey().Hash160()
	pubKeyScript, _ := script.P2PKH(pubKeyHash)

	// Create inputs
	inputs := []types.TxInput{
		{
			PrevTxHash:      crypto.DoubleSHA256([]byte(fmt.Sprintf("prev_tx_%d", seed))),
			OutputIndex:     0,
			SignatureScript: []byte(fmt.Sprintf("signature_%d", seed)),
			Sequence:        0xFFFFFFFF,
		},
	}

	// Create outputs
	outputs := make([]types.TxOutput, outputCount)
	for i := 0; i < outputCount; i++ {
		outputs[i] = types.TxOutput{
			Value:        100000000 / int64(outputCount), // Split 1 BTC
			PubKeyScript: pubKeyScript,
		}
	}

	return &types.Transaction{
		Version:  1,
		Inputs:   inputs,
		Outputs:  outputs,
		LockTime: 0,
	}
}
