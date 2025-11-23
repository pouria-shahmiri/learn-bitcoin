package main

import (
	"fmt"
	"math/big"
	"time"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/consensus"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/crypto"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/mempool"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/monitoring"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/reorg"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/security"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/utxo"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/validation"
)

func main() {
	fmt.Println("=== Phase 10: Reorg Handling & Hardening ===\n")

	// Initialize logger
	logger := monitoring.NewLogger(monitoring.INFO)
	logger.Info("Starting Phase 10 demonstrations")

	// Run all demos
	demoConsensusRules(logger)
	demoCheckpoints(logger)
	demoReorgDetection(logger)
	demoReorgHandling(logger)
	demoMetrics(logger)
	demoRateLimiting(logger)
	demoDoSProtection(logger)
	demoFuzzTesting(logger)

	logger.Info("Phase 10 demonstrations completed successfully!")
}

func demoConsensusRules(logger *monitoring.Logger) {
	fmt.Println("\n--- Demo 1: Consensus Rules ---")
	logger.Info("Testing consensus rules")

	// Create consensus rules for different networks
	mainnet := consensus.NewMainnetRules()
	testnet := consensus.NewTestnetRules()
	regtest := consensus.NewRegtestRules()

	fmt.Printf("Mainnet max block size: %d bytes\n", mainnet.MaxBlockSize)
	fmt.Printf("Testnet coinbase maturity: %d blocks\n", testnet.CoinbaseMaturity)
	fmt.Printf("Regtest halving interval: %d blocks\n", regtest.SubsidyHalvingInterval)

	// Test block subsidy calculation
	heights := []uint64{0, 210000, 420000, 630000}
	fmt.Println("\nBlock subsidies at different heights:")
	for _, height := range heights {
		subsidy := mainnet.GetBlockSubsidy(height)
		btc := float64(subsidy) / 100000000.0
		fmt.Printf("  Height %d: %.8f BTC\n", height, btc)
	}

	// Test BIP activation
	fmt.Println("\nBIP activations at height 500000:")
	height := uint64(500000)
	fmt.Printf("  BIP34 active: %v\n", mainnet.IsBIP34Active(height))
	fmt.Printf("  BIP65 active: %v\n", mainnet.IsBIP65Active(height))
	fmt.Printf("  BIP66 active: %v\n", mainnet.IsBIP66Active(height))
	fmt.Printf("  SegWit active: %v\n", mainnet.IsSegWitActive(height))

	// Test time validation
	currentTime := time.Now()
	blockTime := uint32(currentTime.Unix())
	medianTimePast := uint32(currentTime.Add(-10 * time.Minute).Unix())

	err := mainnet.ValidateBlockTime(blockTime, medianTimePast, currentTime)
	if err != nil {
		logger.Errorf("Block time validation failed: %v", err)
	} else {
		fmt.Println("\n✓ Block time validation passed")
	}

	// Test future block time rejection
	futureBlockTime := uint32(currentTime.Add(3 * time.Hour).Unix())
	err = mainnet.ValidateBlockTime(futureBlockTime, medianTimePast, currentTime)
	if err != nil {
		fmt.Printf("✓ Future block correctly rejected: %v\n", err)
	}
}

func demoCheckpoints(logger *monitoring.Logger) {
	fmt.Println("\n--- Demo 2: Checkpoint Verification ---")
	logger.Info("Testing checkpoint verification")

	verifier := consensus.NewCheckpointVerifier(true)

	// Test checkpoint verification
	testHeight := uint64(11111)
	testHash := types.Hash{} // Would be actual hash in production

	err := verifier.VerifyCheckpoint(testHeight, testHash)
	if err != nil {
		fmt.Printf("Checkpoint verification: %v\n", err)
	}

	// Get last checkpoint
	lastCP := verifier.GetLastCheckpoint(100000)
	if lastCP != nil {
		fmt.Printf("Last checkpoint before height 100000: %d\n", lastCP.Height)
	}

	// Check if height is a checkpoint
	isCP := verifier.IsCheckpoint(11111)
	fmt.Printf("Height 11111 is checkpoint: %v\n", isCP)

	fmt.Println("✓ Checkpoint verification completed")
}

func demoReorgDetection(logger *monitoring.Logger) {
	fmt.Println("\n--- Demo 3: Reorg Detection ---")
	logger.Info("Testing reorg detection")

	// Create blockchain storage
	blockchain, err := storage.NewBlockchainStorage("data/phase10_reorg_test.db")
	if err != nil {
		logger.Errorf("Failed to create blockchain: %v", err)
		return
	}
	defer blockchain.Close()

	// Create genesis block
	genesis := createGenesisBlock()
	isEmpty, _ := blockchain.IsEmpty()
	if isEmpty {
		blockchain.SaveBlock(genesis, 0)
		logger.Info("Created genesis block")
	}

	// Create detector
	detector := reorg.NewReorgDetector(blockchain)

	// Create competing chain
	block1 := createBlock(genesis, 1, 0x1d00ffff)
	block2 := createBlock(block1, 2, 0x1d00ffff)
	competingBlocks := []*types.Block{block1, block2}

	// Detect reorg
	needsReorg, chainInfo, err := detector.DetectReorg(competingBlocks)
	if err != nil {
		logger.Errorf("Reorg detection failed: %v", err)
		return
	}

	if needsReorg {
		fmt.Printf("✓ Reorg detected!\n")
		fmt.Printf("  Fork height: %d\n", chainInfo.ForkHeight)
		fmt.Printf("  New chain height: %d\n", chainInfo.Height)
		fmt.Printf("  Total work: %s\n", chainInfo.TotalWork.String())
	} else {
		fmt.Println("✓ No reorg needed (current chain has more work)")
	}
}

func demoReorgHandling(logger *monitoring.Logger) {
	fmt.Println("\n--- Demo 4: Reorg Handling ---")
	logger.Info("Testing reorg handling")

	// Create components
	blockchain, _ := storage.NewBlockchainStorage("data/phase10_reorg_handler.db")
	defer blockchain.Close()

	utxoSet := utxo.NewUTXOSet()
	mp := mempool.NewMempool(10*1024*1024, 1000, 3600) // 10MB, 1000 sat/byte min, 1 hour max age

	// Create genesis if needed
	isEmpty, _ := blockchain.IsEmpty()
	if isEmpty {
		genesis := createGenesisBlock()
		blockchain.SaveBlock(genesis, 0)

		// Apply genesis to UTXO set
		validator := validation.NewBlockValidator(utxoSet)
		validator.ApplyBlock(genesis, 0)
	}

	// Create reorg handler
	handler := reorg.NewReorgHandler(blockchain, utxoSet, mp)

	// Simulate a reorg scenario
	fmt.Println("\nSimulating chain reorganization...")

	// Create competing blocks
	genesis, _, _ := blockchain.GetBestBlock()
	block1 := createBlock(genesis, 1, 0x1d00ffff)
	block2 := createBlock(block1, 2, 0x1d00ffff)
	newBlocks := []*types.Block{block1, block2}

	// Handle reorg
	err := handler.HandleReorg(newBlocks)
	if err != nil {
		fmt.Printf("Reorg handling: %v\n", err)
	} else {
		fmt.Println("✓ Reorg handled successfully")
	}

	// Check final state
	_, height, _ := blockchain.GetBestBlock()
	fmt.Printf("Final chain height: %d\n", height)
}

func demoMetrics(logger *monitoring.Logger) {
	fmt.Println("\n--- Demo 5: Metrics & Monitoring ---")
	logger.Info("Testing metrics collection")

	metrics := monitoring.NewMetrics()

	// Simulate block processing
	fmt.Println("\nSimulating block processing...")
	for i := 0; i < 10; i++ {
		start := time.Now()
		time.Sleep(10 * time.Millisecond) // Simulate work
		metrics.RecordBlockProcessed(time.Since(start))
	}

	// Simulate transaction processing
	for i := 0; i < 100; i++ {
		start := time.Now()
		time.Sleep(1 * time.Millisecond) // Simulate work
		metrics.RecordTxProcessed(time.Since(start))
	}

	// Simulate peer connections
	metrics.IncrementPeerCount(true)  // inbound
	metrics.IncrementPeerCount(false) // outbound
	metrics.IncrementPeerCount(true)

	// Simulate network activity
	metrics.RecordBytesReceived(1024 * 1024) // 1 MB
	metrics.RecordBytesSent(512 * 1024)      // 512 KB

	// Set mempool metrics
	metrics.SetMempoolSize(150, 75000)

	// Set UTXO metrics
	metrics.SetUTXOSetSize(1000000)
	for i := 0; i < 100; i++ {
		if i%3 == 0 {
			metrics.RecordUTXOCacheMiss()
		} else {
			metrics.RecordUTXOCacheHit()
		}
	}

	// Record reorg
	metrics.RecordReorg(5)

	// Print summary
	fmt.Println("\nMetrics Summary:")
	summary := metrics.Summary()
	for key, value := range summary {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println("\n✓ Metrics collection completed")
}

func demoRateLimiting(logger *monitoring.Logger) {
	fmt.Println("\n--- Demo 6: Rate Limiting ---")
	logger.Info("Testing rate limiting")

	// Create rate limiter (10 requests per second, burst of 20)
	limiter := security.NewRateLimiter(10, 20)

	// Test rate limiting
	allowed := 0
	denied := 0

	fmt.Println("\nTesting rate limiter (100 requests)...")
	for i := 0; i < 100; i++ {
		if limiter.Allow() {
			allowed++
		} else {
			denied++
		}
		time.Sleep(5 * time.Millisecond)
	}

	fmt.Printf("  Allowed: %d\n", allowed)
	fmt.Printf("  Denied: %d\n", denied)

	// Test connection rate limiting
	connLimiter := security.NewConnectionRateLimiter(100, 10, 20)

	fmt.Println("\nTesting connection rate limiter...")
	testIPs := []string{"192.168.1.1", "192.168.1.2", "192.168.1.1"}

	for _, ip := range testIPs {
		if connLimiter.AllowConnection(ip) {
			fmt.Printf("  ✓ Connection from %s allowed\n", ip)
		} else {
			fmt.Printf("  ✗ Connection from %s denied\n", ip)
		}
	}

	fmt.Println("\n✓ Rate limiting tests completed")
}

func demoDoSProtection(logger *monitoring.Logger) {
	fmt.Println("\n--- Demo 7: DoS Protection ---")
	logger.Info("Testing DoS protection")

	protection := security.NewDoSProtection()

	// Test connection allowance
	testAddr := &mockAddr{addr: "192.168.1.100:8333"}

	fmt.Println("\nTesting connection protection...")
	for i := 0; i < 5; i++ {
		err := protection.AllowConnection(testAddr)
		if err != nil {
			fmt.Printf("  Connection %d: denied - %v\n", i+1, err)
		} else {
			fmt.Printf("  Connection %d: allowed\n", i+1)
		}
	}

	// Test bandwidth limiting
	fmt.Println("\nTesting bandwidth protection...")
	peer := "192.168.1.100:8333"

	sizes := []int{1024, 2048, 5000000} // Last one should trigger limit
	for _, size := range sizes {
		err := protection.AllowBytes(peer, size)
		if err != nil {
			fmt.Printf("  %d bytes: denied - %v\n", size, err)
		} else {
			fmt.Printf("  %d bytes: allowed\n", size)
		}
	}

	// Test IP banning
	fmt.Println("\nTesting IP banning...")
	maliciousIP := "10.0.0.1"
	protection.BanIP(maliciousIP)

	bannedIPs := protection.GetBannedIPs()
	fmt.Printf("  Banned IPs: %v\n", bannedIPs)

	protection.UnbanIP(maliciousIP)
	fmt.Println("  ✓ IP unbanned")

	fmt.Println("\n✓ DoS protection tests completed")
}

func demoFuzzTesting(logger *monitoring.Logger) {
	fmt.Println("\n--- Demo 8: Fuzz Testing ---")
	logger.Info("Testing fuzz capabilities")

	fuzzer := security.NewFuzzTester(10, 1000)

	// Generate random data
	fmt.Println("\nGenerating random fuzz data...")
	for i := 0; i < 5; i++ {
		data, err := fuzzer.GenerateRandomBytes()
		if err != nil {
			logger.Errorf("Failed to generate fuzz data: %v", err)
			continue
		}
		fmt.Printf("  Generated %d bytes of random data\n", len(data))
	}

	// Test mutation
	fmt.Println("\nTesting data mutation...")
	original := []byte("Hello, Bitcoin!")
	for i := 0; i < 3; i++ {
		mutated := fuzzer.MutateBytes(original)
		fmt.Printf("  Mutation %d: %d bytes\n", i+1, len(mutated))
	}

	// Test input validation
	fmt.Println("\nTesting input validation...")
	validator := security.NewInputValidator(1024 * 1024) // 1 MB max

	testData := make([]byte, 500)
	err := validator.ValidateMessageSize(testData)
	if err != nil {
		fmt.Printf("  Validation failed: %v\n", err)
	} else {
		fmt.Printf("  ✓ Message size valid (%d bytes)\n", len(testData))
	}

	// Test string sanitization
	dirtyString := "Hello\x00World\x01!"
	clean := validator.SanitizeString(dirtyString)
	fmt.Printf("  Sanitized string: %q\n", clean)

	// Test port validation
	ports := []int{8333, 0, 70000}
	for _, port := range ports {
		err := validator.ValidatePort(port)
		if err != nil {
			fmt.Printf("  Port %d: invalid - %v\n", port, err)
		} else {
			fmt.Printf("  Port %d: ✓ valid\n", port)
		}
	}

	fmt.Println("\n✓ Fuzz testing completed")
}

// Helper functions

func createGenesisBlock() *types.Block {
	// Create coinbase transaction
	coinbase := types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      types.Hash{},
				OutputIndex:     0xffffffff,
				SignatureScript: []byte("Genesis Block"),
				Sequence:        0xffffffff,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        50 * 100000000,           // 50 BTC
				PubKeyScript: []byte{0x76, 0xa9, 0x14}, // Dummy script
			},
		},
		LockTime: 0,
	}

	// Calculate merkle root
	txHash, _ := serialization.HashTransaction(&coinbase)
	merkleRoot := crypto.DoubleSHA256(txHash[:])

	return &types.Block{
		Header: types.BlockHeader{
			Version:       1,
			PrevBlockHash: types.Hash{},
			MerkleRoot:    merkleRoot,
			Timestamp:     uint32(time.Now().Unix()),
			Bits:          0x1d00ffff,
			Nonce:         0,
		},
		Transactions: []types.Transaction{coinbase},
	}
}

func createBlock(prevBlock *types.Block, height uint64, bits uint32) *types.Block {
	prevHash, _ := serialization.HashBlockHeader(&prevBlock.Header)

	// Create coinbase
	coinbase := types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      types.Hash{},
				OutputIndex:     0xffffffff,
				SignatureScript: []byte(fmt.Sprintf("Block %d", height)),
				Sequence:        0xffffffff,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        50 * 100000000,
				PubKeyScript: []byte{0x76, 0xa9, 0x14},
			},
		},
		LockTime: 0,
	}

	txHash, _ := serialization.HashTransaction(&coinbase)
	merkleRoot := crypto.DoubleSHA256(txHash[:])

	block := &types.Block{
		Header: types.BlockHeader{
			Version:       1,
			PrevBlockHash: prevHash,
			MerkleRoot:    merkleRoot,
			Timestamp:     uint32(time.Now().Unix()),
			Bits:          bits,
			Nonce:         0,
		},
		Transactions: []types.Transaction{coinbase},
	}

	// Simple mining - just find any valid nonce
	// In production, use the actual miner
	for nonce := uint32(0); nonce < 1000000; nonce++ {
		block.Header.Nonce = nonce
		blockHash, _ := serialization.HashBlockHeader(&block.Header)
		// Check if hash meets difficulty (simplified)
		if blockHash[0] == 0 || blockHash[1] == 0 {
			break
		}
	}

	return block
}

// Mock address for testing
type mockAddr struct {
	addr string
}

func (m *mockAddr) Network() string {
	return "tcp"
}

func (m *mockAddr) String() string {
	return m.addr
}

// Dummy function to satisfy unused import
func init() {
	_ = big.NewInt(0)
}
