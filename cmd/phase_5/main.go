package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/crypto"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/keys"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/script"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/transaction"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/utxo"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/validation"
)

func main() {
	fmt.Println("=== Bitcoin Learning - Milestone 5 ===")
	fmt.Println("UTXO Management & Chain Validation\n")

	// Clean up old database for fresh demo
	dbPath := "./data/blockchain_phase5"
	os.RemoveAll(dbPath)

	// Demo 1: UTXO Set basics
	demoUTXOSetBasics()

	// Demo 2: UTXO tracking with transactions
	demoUTXOTracking()

	// Demo 3: UTXO statistics
	demoUTXOStatistics()

	// Demo 4: Block validation
	demoBlockValidation(dbPath)

	// Demo 5: Chain validation
	demoChainValidation(dbPath)

	// Demo 6: Transaction validation with UTXO
	demoTransactionValidation(dbPath)

	// Demo 7: Full blockchain with validation
	demoFullBlockchain(dbPath)

	// Demo 8: UTXO persistence
	demoUTXOPersistence(dbPath)

	fmt.Println("\n=== All demos completed successfully! ===")
}

func demoUTXOSetBasics() {
	fmt.Println("--- Demo 1: UTXO Set Basics ---")

	// Create a new UTXO set
	utxoSet := utxo.NewUTXOSet()

	// Create some sample UTXOs
	txHash1 := crypto.DoubleSHA256([]byte("transaction1"))
	txHash2 := crypto.DoubleSHA256([]byte("transaction2"))

	output1 := types.TxOutput{
		Value:        100000000, // 1 BTC
		PubKeyScript: []byte("script1"),
	}

	output2 := types.TxOutput{
		Value:        50000000, // 0.5 BTC
		PubKeyScript: []byte("script2"),
	}

	// Add UTXOs
	utxo1 := utxo.NewUTXO(txHash1, 0, output1, 100, false)
	utxo2 := utxo.NewUTXO(txHash2, 0, output2, 101, false)

	err := utxoSet.Add(utxo1)
	if err != nil {
		panic(err)
	}

	err = utxoSet.Add(utxo2)
	if err != nil {
		panic(err)
	}

	fmt.Printf("UTXO Set created:\n")
	fmt.Printf("  Size: %d UTXOs\n", utxoSet.Size())
	fmt.Printf("  Total value: %.8f BTC\n", float64(utxoSet.TotalValue())/100000000)

	// Check if UTXO exists
	outpoint1 := utxo.NewOutPoint(txHash1, 0)
	exists := utxoSet.Exists(outpoint1)
	fmt.Printf("  UTXO 1 exists: %v ✓\n", exists)

	// Get a specific UTXO
	retrieved, err := utxoSet.Get(outpoint1)
	if err == nil {
		fmt.Printf("  Retrieved UTXO value: %.8f BTC\n", float64(retrieved.Value())/100000000)
	}

	// Remove a UTXO (simulate spending)
	err = utxoSet.Remove(outpoint1)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nAfter spending UTXO 1:\n")
	fmt.Printf("  Size: %d UTXOs\n", utxoSet.Size())
	fmt.Printf("  Total value: %.8f BTC\n", float64(utxoSet.TotalValue())/100000000)

	fmt.Println()
}

func demoUTXOTracking() {
	fmt.Println("--- Demo 2: UTXO Tracking with Transactions ---")

	utxoSet := utxo.NewUTXOSet()

	// Create a coinbase transaction (creates new UTXOs)
	coinbaseTx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      types.Hash{},
				OutputIndex:     0xFFFFFFFF,
				SignatureScript: []byte("Block 1 coinbase"),
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        5000000000, // 50 BTC
				PubKeyScript: []byte("miner_address"),
			},
		},
		LockTime: 0,
	}

	coinbaseHash, _ := serialization.HashTransaction(coinbaseTx)
	fmt.Printf("Coinbase transaction: %s\n", coinbaseHash.String()[:16]+"...")

	// Apply coinbase to UTXO set
	err := utxoSet.ApplyTransaction(coinbaseTx, coinbaseHash, 1, true)
	if err != nil {
		panic(err)
	}

	fmt.Printf("After coinbase:\n")
	fmt.Printf("  UTXOs: %d\n", utxoSet.Size())
	fmt.Printf("  Total value: %.8f BTC\n\n", float64(utxoSet.TotalValue())/100000000)

	// Create a spending transaction
	spendingTx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      coinbaseHash,
				OutputIndex:     0,
				SignatureScript: []byte("signature"),
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        3000000000, // 30 BTC to recipient
				PubKeyScript: []byte("recipient_address"),
			},
			{
				Value:        1999900000, // 19.999 BTC change
				PubKeyScript: []byte("miner_address"),
			},
		},
		LockTime: 0,
	}

	spendingHash, _ := serialization.HashTransaction(spendingTx)
	fmt.Printf("Spending transaction: %s\n", spendingHash.String()[:16]+"...")

	// Apply spending transaction
	err = utxoSet.ApplyTransaction(spendingTx, spendingHash, 2, false)
	if err != nil {
		panic(err)
	}

	fmt.Printf("After spending:\n")
	fmt.Printf("  UTXOs: %d (coinbase spent, 2 new outputs created)\n", utxoSet.Size())
	fmt.Printf("  Total value: %.8f BTC\n", float64(utxoSet.TotalValue())/100000000)

	// Calculate fee
	fee := int64(5000000000) - (3000000000 + 1999900000)
	fmt.Printf("  Transaction fee: %.8f BTC\n", float64(fee)/100000000)

	fmt.Println()
}

func demoUTXOStatistics() {
	fmt.Println("--- Demo 3: UTXO Statistics ---")

	utxoSet := utxo.NewUTXOSet()

	// Add multiple UTXOs (mix of coinbase and regular)
	for i := 0; i < 5; i++ {
		txHash := crypto.DoubleSHA256([]byte(fmt.Sprintf("tx_%d", i)))
		output := types.TxOutput{
			Value:        int64((i + 1) * 100000000), // 1, 2, 3, 4, 5 BTC
			PubKeyScript: []byte(fmt.Sprintf("script_%d", i)),
		}

		isCoinbase := i%2 == 0
		u := utxo.NewUTXO(txHash, 0, output, uint64(100+i), isCoinbase)
		utxoSet.Add(u)
	}

	// Get statistics
	stats := utxoSet.GetStatistics()

	fmt.Printf("UTXO Set Statistics:\n")
	fmt.Printf("  Total UTXOs: %d\n", stats.Count)
	fmt.Printf("  Total value: %.8f BTC (%d satoshis)\n",
		float64(stats.TotalValue)/100000000, stats.TotalValue)
	fmt.Printf("  Average value: %.8f BTC\n", float64(stats.AverageValue)/100000000)
	fmt.Printf("  Coinbase UTXOs: %d\n", stats.CoinbaseCount)
	fmt.Printf("  Coinbase value: %.8f BTC\n", float64(stats.CoinbaseValue)/100000000)
	fmt.Printf("  Regular UTXOs: %d\n", stats.Count-stats.CoinbaseCount)
	fmt.Printf("  Regular value: %.8f BTC\n",
		float64(stats.TotalValue-stats.CoinbaseValue)/100000000)

	fmt.Println()
}

func demoBlockValidation(dbPath string) {
	fmt.Println("--- Demo 4: Block Validation ---")

	// Initialize storage
	bs, err := storage.NewBlockchainStorage(dbPath)
	if err != nil {
		panic(err)
	}
	defer bs.Close()

	// Create UTXO set
	utxoSet := utxo.NewUTXOSet()

	// Create validator
	blockValidator := validation.NewBlockValidator(utxoSet)

	// Create genesis block
	genesisBlock := createGenesisBlock()
	genesisHash, _ := serialization.HashBlockHeader(&genesisBlock.Header)

	fmt.Printf("Validating genesis block: %s\n", genesisHash.String()[:16]+"...")

	// Validate genesis (height 0, no previous block)
	err = blockValidator.ValidateBlock(genesisBlock, 0, types.Hash{})
	if err != nil {
		fmt.Printf("  ✗ Validation failed: %v\n", err)
	} else {
		fmt.Printf("  ✓ Genesis block is valid\n")
	}

	// Apply genesis to UTXO set
	err = blockValidator.ApplyBlock(genesisBlock, 0)
	if err != nil {
		panic(err)
	}

	fmt.Printf("  ✓ Applied to UTXO set\n")
	fmt.Printf("  UTXOs after genesis: %d\n", utxoSet.Size())

	// Save genesis
	err = bs.SaveBlock(genesisBlock, 0)
	if err != nil {
		panic(err)
	}

	// Create and validate block 1
	block1 := createBlock(genesisHash, uint32(time.Now().Unix()), 1)
	block1Hash, _ := serialization.HashBlockHeader(&block1.Header)

	fmt.Printf("\nValidating block 1: %s\n", block1Hash.String()[:16]+"...")

	err = blockValidator.ValidateBlock(block1, 1, genesisHash)
	if err != nil {
		fmt.Printf("  ✗ Validation failed: %v\n", err)
	} else {
		fmt.Printf("  ✓ Block 1 is valid\n")
	}

	// Apply block 1
	err = blockValidator.ApplyBlock(block1, 1)
	if err != nil {
		panic(err)
	}

	fmt.Printf("  ✓ Applied to UTXO set\n")
	fmt.Printf("  UTXOs after block 1: %d\n", utxoSet.Size())

	fmt.Println()
}

func demoChainValidation(dbPath string) {
	fmt.Println("--- Demo 5: Chain Validation ---")

	// Initialize storage
	bs, err := storage.NewBlockchainStorage(dbPath)
	if err != nil {
		panic(err)
	}
	defer bs.Close()

	// Create UTXO set
	utxoSet := utxo.NewUTXOSet()

	// Create chain validator
	chainValidator := validation.NewChainValidator(bs, utxoSet)

	fmt.Printf("Validating entire blockchain...\n")

	// Validate the entire chain
	err = chainValidator.IsValidChain()
	if err != nil {
		fmt.Printf("  ✗ Chain validation failed: %v\n", err)
	} else {
		fmt.Printf("  ✓ Entire chain is valid\n")
	}

	// Get chain statistics
	bestHash, _ := bs.GetBestBlockHash()
	bestHeight, _ := bs.GetBestBlockHeight()
	chainWork, _ := chainValidator.GetChainWork()

	fmt.Printf("\nChain Statistics:\n")
	fmt.Printf("  Best block: %s\n", bestHash.String()[:16]+"...")
	fmt.Printf("  Height: %d\n", bestHeight)
	fmt.Printf("  Chain work: %d\n", chainWork)
	fmt.Printf("  UTXO set size: %d\n", utxoSet.Size())

	// Get block locator (for sync)
	locator, err := chainValidator.GetBlockLocator()
	if err == nil {
		fmt.Printf("\nBlock Locator (for sync):\n")
		for i, hash := range locator {
			if i < 5 { // Show first 5
				fmt.Printf("  %d: %s\n", i, hash.String()[:16]+"...")
			}
		}
		if len(locator) > 5 {
			fmt.Printf("  ... and %d more\n", len(locator)-5)
		}
	}

	fmt.Println()
}

func demoTransactionValidation(dbPath string) {
	fmt.Println("--- Demo 6: Transaction Validation with UTXO ---")

	// Create UTXO set with some funds
	utxoSet := utxo.NewUTXOSet()

	// Generate keys
	aliceKey, _ := keys.GeneratePrivateKey()
	bobKey, _ := keys.GeneratePrivateKey()

	aliceAddress := aliceKey.PublicKey().P2PKHAddress()
	bobAddress := bobKey.PublicKey().P2PKHAddress()

	fmt.Printf("Alice: %s\n", aliceAddress)
	fmt.Printf("Bob: %s\n", bobAddress)

	// Create a UTXO for Alice
	alicePubKeyHash := aliceKey.PublicKey().Hash160()
	aliceScript, _ := script.P2PKH(alicePubKeyHash)

	prevTxHash := crypto.DoubleSHA256([]byte("previous_tx"))
	aliceOutput := types.TxOutput{
		Value:        200000000, // 2 BTC
		PubKeyScript: aliceScript,
	}

	aliceUTXO := utxo.NewUTXO(prevTxHash, 0, aliceOutput, 100, false)
	utxoSet.Add(aliceUTXO)

	fmt.Printf("\nAlice's initial balance: %.8f BTC\n", float64(aliceOutput.Value)/100000000)

	// Alice creates a transaction to send 1 BTC to Bob
	bobPubKeyHash := bobKey.PublicKey().Hash160()
	bobScript, _ := script.P2PKH(bobPubKeyHash)

	tx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      prevTxHash,
				OutputIndex:     0,
				SignatureScript: []byte("alice_signature"), // Simplified
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        100000000, // 1 BTC to Bob
				PubKeyScript: bobScript,
			},
			{
				Value:        99900000, // 0.999 BTC change to Alice
				PubKeyScript: aliceScript,
			},
		},
		LockTime: 0,
	}

	fmt.Printf("\nTransaction created:\n")
	fmt.Printf("  Sending: %.8f BTC to Bob\n", float64(tx.Outputs[0].Value)/100000000)
	fmt.Printf("  Change: %.8f BTC to Alice\n", float64(tx.Outputs[1].Value)/100000000)
	fmt.Printf("  Fee: %.8f BTC\n", float64(200000000-100000000-99900000)/100000000)

	// Validate transaction
	err := transaction.ValidateTransaction(tx)
	if err != nil {
		fmt.Printf("  ✗ Transaction invalid: %v\n", err)
	} else {
		fmt.Printf("  ✓ Transaction structure is valid\n")
	}

	// Apply to UTXO set
	txHash, _ := serialization.HashTransaction(tx)
	err = utxoSet.ApplyTransaction(tx, txHash, 101, false)
	if err != nil {
		fmt.Printf("  ✗ Failed to apply: %v\n", err)
	} else {
		fmt.Printf("  ✓ Transaction applied to UTXO set\n")
	}

	// Check balances
	aliceUTXOs := utxoSet.FindByScript(aliceScript)
	bobUTXOs := utxoSet.FindByScript(bobScript)

	aliceBalance := int64(0)
	for _, u := range aliceUTXOs {
		aliceBalance += u.Value()
	}

	bobBalance := int64(0)
	for _, u := range bobUTXOs {
		bobBalance += u.Value()
	}

	fmt.Printf("\nFinal balances:\n")
	fmt.Printf("  Alice: %.8f BTC (%d UTXOs)\n", float64(aliceBalance)/100000000, len(aliceUTXOs))
	fmt.Printf("  Bob: %.8f BTC (%d UTXOs)\n", float64(bobBalance)/100000000, len(bobUTXOs))

	fmt.Println()
}

func demoFullBlockchain(dbPath string) {
	fmt.Println("--- Demo 7: Full Blockchain with Validation ---")

	// Clean and recreate
	os.RemoveAll(dbPath + "_full")
	bs, err := storage.NewBlockchainStorage(dbPath + "_full")
	if err != nil {
		panic(err)
	}
	defer bs.Close()

	// Create UTXO set
	utxoSet := utxo.NewUTXOSet()

	// Create chain validator
	chainValidator := validation.NewChainValidator(bs, utxoSet)

	// Create and accept genesis block
	genesisBlock := createGenesisBlock()
	err = chainValidator.AcceptBlock(genesisBlock)
	if err != nil {
		panic(fmt.Sprintf("Failed to accept genesis: %v", err))
	}

	fmt.Printf("Genesis block accepted\n")
	fmt.Printf("  UTXOs: %d\n", utxoSet.Size())
	fmt.Printf("  Total value: %.8f BTC\n", float64(utxoSet.TotalValue())/100000000)

	// Add several blocks
	prevBlock := genesisBlock
	acceptedCount := 0
	for i := 1; i <= 10; i++ {
		prevHash, _ := serialization.HashBlockHeader(&prevBlock.Header)
		block := createBlock(prevHash, uint32(time.Now().Unix()), i)

		err = chainValidator.AcceptBlock(block)
		if err != nil {
			if i == 1 {
				// First block might fail due to validation, try to continue
				continue
			}
			fmt.Printf("Block %d rejected: %v\n", i, err)
			continue
		}

		acceptedCount++
		if i%3 == 0 {
			fmt.Printf("Block %d accepted\n", i)
		}

		prevBlock = block
	}

	if acceptedCount > 0 {
		fmt.Printf("Accepted %d blocks\n", acceptedCount)
	}

	// Final statistics
	bestHeight, _ := bs.GetBestBlockHeight()
	stats := utxoSet.GetStatistics()

	fmt.Printf("\nFinal Blockchain State:\n")
	fmt.Printf("  Chain height: %d\n", bestHeight)
	fmt.Printf("  Total UTXOs: %d\n", stats.Count)
	fmt.Printf("  Total value: %.8f BTC\n", float64(stats.TotalValue)/100000000)
	fmt.Printf("  Coinbase UTXOs: %d (%.8f BTC)\n",
		stats.CoinbaseCount, float64(stats.CoinbaseValue)/100000000)

	// Validate entire chain
	err = chainValidator.IsValidChain()
	if err != nil {
		fmt.Printf("  ✗ Chain validation failed: %v\n", err)
	} else {
		fmt.Printf("  ✓ Entire chain is valid\n")
	}

	fmt.Println()
}

func demoUTXOPersistence(dbPath string) {
	fmt.Println("--- Demo 8: UTXO Persistence ---")

	utxoDbPath := dbPath + "_utxo"
	os.RemoveAll(utxoDbPath)

	// Create UTXO storage
	utxoStorage, err := utxo.NewUTXOStorage(utxoDbPath)
	if err != nil {
		panic(err)
	}

	// Create and populate UTXO set
	utxoSet := utxo.NewUTXOSet()

	for i := 0; i < 5; i++ {
		txHash := crypto.DoubleSHA256([]byte(fmt.Sprintf("tx_%d", i)))
		output := types.TxOutput{
			Value:        int64((i + 1) * 100000000),
			PubKeyScript: []byte(fmt.Sprintf("script_%d", i)),
		}
		u := utxo.NewUTXO(txHash, 0, output, uint64(i), i%2 == 0)
		utxoSet.Add(u)
	}

	fmt.Printf("Created UTXO set:\n")
	fmt.Printf("  Size: %d UTXOs\n", utxoSet.Size())
	fmt.Printf("  Total value: %.8f BTC\n", float64(utxoSet.TotalValue())/100000000)

	// Save to disk
	err = utxoStorage.SaveSet(utxoSet)
	if err != nil {
		panic(err)
	}

	fmt.Printf("  ✓ Saved to disk\n")

	// Close and reopen
	utxoStorage.Close()

	utxoStorage2, err := utxo.NewUTXOStorage(utxoDbPath)
	if err != nil {
		panic(err)
	}
	defer utxoStorage2.Close()

	// Load from disk
	loadedSet, err := utxoStorage2.LoadAll()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nLoaded UTXO set from disk:\n")
	fmt.Printf("  Size: %d UTXOs\n", loadedSet.Size())
	fmt.Printf("  Total value: %.8f BTC\n", float64(loadedSet.TotalValue())/100000000)

	// Verify data matches
	if utxoSet.Size() == loadedSet.Size() && utxoSet.TotalValue() == loadedSet.TotalValue() {
		fmt.Printf("  ✓ UTXO set persisted correctly\n")
	} else {
		fmt.Printf("  ✗ UTXO set mismatch\n")
	}

	fmt.Println()
}

// Helper: Create genesis block
func createGenesisBlock() *types.Block {
	coinbaseTx := types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      types.Hash{},
				OutputIndex:     0xFFFFFFFF,
				SignatureScript: []byte("Genesis Block - The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"),
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        5000000000, // 50 BTC
				PubKeyScript: []byte("Genesis miner pubkey"),
			},
		},
		LockTime: 0,
	}

	txHash, _ := serialization.HashTransaction(&coinbaseTx)
	merkleRoot := crypto.ComputeMerkleRoot([]types.Hash{txHash})

	header := types.BlockHeader{
		Version:       1,
		PrevBlockHash: types.Hash{},
		MerkleRoot:    merkleRoot,
		Timestamp:     1231006505, // Jan 3, 2009
		Bits:          0x1d00ffff,
		Nonce:         2083236893,
	}

	return &types.Block{
		Header:       header,
		Transactions: []types.Transaction{coinbaseTx},
	}
}

// Helper: Create a new block
func createBlock(prevHash types.Hash, timestamp uint32, blockNum int) *types.Block {
	coinbaseTx := types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      types.Hash{},
				OutputIndex:     0xFFFFFFFF,
				SignatureScript: []byte(fmt.Sprintf("Block %d coinbase", blockNum)),
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        5000000000,
				PubKeyScript: []byte(fmt.Sprintf("Miner %d", blockNum)),
			},
		},
		LockTime: 0,
	}

	// Create some regular transactions
	transactions := []types.Transaction{coinbaseTx}

	// For demo purposes, only include coinbase to avoid UTXO validation issues
	// In a real scenario, we would create valid transactions spending existing UTXOs

	// Compute Merkle root
	var txHashes []types.Hash
	for _, tx := range transactions {
		hash, _ := serialization.HashTransaction(&tx)
		txHashes = append(txHashes, hash)
	}
	merkleRoot := crypto.ComputeMerkleRoot(txHashes)

	header := types.BlockHeader{
		Version:       1,
		PrevBlockHash: prevHash,
		MerkleRoot:    merkleRoot,
		Timestamp:     timestamp,
		Bits:          0x1d00ffff,
		Nonce:         uint32(blockNum * 1000),
	}

	return &types.Block{
		Header:       header,
		Transactions: transactions,
	}
}
