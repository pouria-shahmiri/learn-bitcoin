package main

import (
	"fmt"
	"time"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/mempool"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/mining"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/validation"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║         Milestone 7: Mining & Proof-of-Work               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Run all demos
	demo1_CoinbaseTransaction()
	demo2_BlockConstruction()
	demo3_Mining()
	demo4_DifficultyComparison()

	fmt.Println("\n✅ All Phase 7 demos completed successfully!")
}

func demo1_CoinbaseTransaction() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Demo 1: Coinbase Transaction")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create coinbase for different block heights
	testCases := []struct {
		height uint64
		fees   int64
		desc   string
	}{
		{0, 0, "Genesis block"},
		{100, 50000000, "Block 100 with fees"},
		{210000, 100000000, "First halving block"},
		{420000, 75000000, "Second halving block"},
	}

	minerAddress := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa" // Satoshi's address

	for _, tc := range testCases {
		fmt.Printf("\n📦 %s (Height: %d)\n", tc.desc, tc.height)

		coinbase, err := mining.CreateCoinbase(tc.height, tc.fees, minerAddress, 0)
		if err != nil {
			panic(err)
		}

		// Get block reward
		reward := validation.GetBlockReward(tc.height)
		total := reward + tc.fees

		fmt.Printf("   Block reward: %.8f BTC\n", float64(reward)/100000000)
		fmt.Printf("   Transaction fees: %.8f BTC\n", float64(tc.fees)/100000000)
		fmt.Printf("   Total payout: %.8f BTC\n", float64(total)/100000000)
		fmt.Printf("   Coinbase value: %.8f BTC\n", float64(coinbase.Outputs[0].Value)/100000000)

		// Verify coinbase
		if err := validation.ValidateBlockReward(coinbase.Outputs[0].Value, tc.fees, tc.height); err != nil {
			panic(err)
		}
		fmt.Printf("   ✅ Coinbase valid\n")

		// Show coinbase details
		fmt.Printf("   Input: PrevTx=%x, Index=%d\n",
			coinbase.Inputs[0].PrevTxHash[:4], coinbase.Inputs[0].OutputIndex)
		fmt.Printf("   ScriptSig length: %d bytes\n", len(coinbase.Inputs[0].SignatureScript))
	}
}

func demo2_BlockConstruction() {
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Demo 2: Block Construction")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create mempool
	// Create mempool (maxSize: 1MB, minFeeRate: 1 sat/byte, maxTxAge: 24 hours)
	mp := mempool.NewMempool(1000000, 1, 86400)
	fmt.Printf("\n📝 Created mempool (max size: 1000 txs)\n")

	// Create block builder
	builder := mining.NewBlockBuilder(mp)

	// Previous block (genesis for this demo)
	prevBlockHash := types.Hash{} // All zeros for genesis
	height := uint64(1)
	minerAddress := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	difficulty := uint32(0x1d00ffff) // Standard difficulty bits

	fmt.Printf("\n🏗️  Building block template...\n")
	fmt.Printf("   Height: %d\n", height)
	fmt.Printf("   Previous block: %x...\n", prevBlockHash[:4])
	fmt.Printf("   Miner: %s\n", minerAddress)

	// Create block template
	template, err := builder.CreateBlockTemplate(prevBlockHash, height, minerAddress, difficulty)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n✅ Block template created:\n")
	fmt.Printf("   Version: %d\n", template.Version)
	fmt.Printf("   Transactions: %d\n", len(template.Transactions))
	fmt.Printf("   Total fees: %.8f BTC\n", float64(template.TotalFees)/100000000)
	fmt.Printf("   Timestamp: %s\n", time.Unix(int64(template.Timestamp), 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("   Difficulty: 0x%08x\n", template.Bits)

	// Build block with nonce 0 (not mined yet)
	block, err := mining.BuildBlock(template, 0)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n📊 Block structure:\n")
	fmt.Printf("   Header size: 80 bytes\n")
	fmt.Printf("   Merkle root: %x...\n", block.Header.MerkleRoot[:4])
	fmt.Printf("   Nonce: %d (not mined)\n", block.Header.Nonce)

	// Serialize and show size
	serialized, err := serialization.SerializeBlock(block)
	if err != nil {
		panic(err)
	}
	fmt.Printf("   Total block size: %d bytes\n", len(serialized))
}

func demo3_Mining() {
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Demo 3: Mining with Proof-of-Work")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create a simple block template
	mp := mempool.NewMempool(1000000, 1, 86400)
	builder := mining.NewBlockBuilder(mp)

	prevBlockHash := types.Hash{}
	height := uint64(1)
	minerAddress := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	difficulty := uint32(0x1d00ffff)

	template, err := builder.CreateBlockTemplate(prevBlockHash, height, minerAddress, difficulty)
	if err != nil {
		panic(err)
	}

	// Mine the block with low difficulty (2 leading zero bytes)
	targetZeros := 2
	fmt.Printf("\n⛏️  Mining block with %d leading zero bytes...\n", targetZeros)

	miner := mining.NewMiner()
	block, err := miner.MineBlock(template, targetZeros)
	if err != nil {
		panic(err)
	}

	// Validate the mined block
	fmt.Printf("\n🔍 Validating mined block...\n")
	if err := mining.ValidateMinedBlock(block, targetZeros); err != nil {
		panic(err)
	}
	fmt.Printf("   ✅ Block validation passed\n")

	// Show block hash
	blockHash, err := serialization.HashBlockHeader(&block.Header)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n📋 Block details:\n")
	fmt.Printf("   Block hash: %x\n", blockHash)
	fmt.Printf("   Leading zeros: %d bytes\n", countLeadingZeros(blockHash[:]))
	fmt.Printf("   Nonce: %d\n", block.Header.Nonce)
	fmt.Printf("   Transactions: %d\n", len(block.Transactions))

	// Show mining stats
	stats := miner.GetStats()
	fmt.Printf("\n📊 Mining statistics:\n")
	fmt.Printf("   Total attempts: %d\n", stats.Attempts)
	fmt.Printf("   Hash rate: %.0f H/s\n", stats.HashRate)
	fmt.Printf("   Time elapsed: %v\n", time.Since(stats.StartTime))
}

func demo4_DifficultyComparison() {
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Demo 4: Difficulty Comparison")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Printf("\n📈 Mining blocks at different difficulties...\n")

	// Create template
	mp := mempool.NewMempool(1000000, 1, 86400)
	builder := mining.NewBlockBuilder(mp)
	prevBlockHash := types.Hash{}
	minerAddress := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

	difficulties := []struct {
		zeros int
		desc  string
	}{
		{1, "Very Easy (1 zero byte)"},
		{2, "Easy (2 zero bytes)"},
		{3, "Medium (3 zero bytes) - This may take a while!"},
	}

	for i, diff := range difficulties {
		if diff.zeros > 2 {
			fmt.Printf("\n⚠️  Skipping %s (would take too long)\n", diff.desc)
			fmt.Printf("   Estimated attempts: ~%d million\n", 1<<(8*diff.zeros)/1000000)
			continue
		}

		fmt.Printf("\n🎯 Difficulty %d: %s\n", i+1, diff.desc)

		template, err := builder.CreateBlockTemplate(
			prevBlockHash,
			uint64(i+1),
			minerAddress,
			0x1d00ffff,
		)
		if err != nil {
			panic(err)
		}

		miner := mining.NewMiner()
		startTime := time.Now()

		block, err := miner.MineBlock(template, diff.zeros)
		if err != nil {
			panic(err)
		}

		elapsed := time.Since(startTime)
		stats := miner.GetStats()

		blockHash, _ := serialization.HashBlockHeader(&block.Header)

		fmt.Printf("   Hash: %x...\n", blockHash[:8])
		fmt.Printf("   Nonce: %d\n", block.Header.Nonce)
		fmt.Printf("   Attempts: %d\n", stats.Attempts)
		fmt.Printf("   Time: %v\n", elapsed)
		fmt.Printf("   Hash rate: %.0f H/s\n", stats.HashRate)
	}

	fmt.Printf("\n💡 Key insights:\n")
	fmt.Printf("   - Each additional zero byte increases difficulty ~256x\n")
	fmt.Printf("   - Bitcoin uses ~20 leading zero bytes!\n")
	fmt.Printf("   - Mining difficulty adjusts every 2016 blocks\n")
	fmt.Printf("   - Target: 10 minutes per block on average\n")
}

// Helper function to count leading zero bytes
func countLeadingZeros(hash []byte) int {
	count := 0
	for _, b := range hash {
		if b == 0 {
			count++
		} else {
			break
		}
	}
	return count
}
