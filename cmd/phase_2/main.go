package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/crypto"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

func main() {
	fmt.Println("=== Bitcoin Learning - Milestone 2 ===")
	fmt.Println("Persistent Storage with LevelDB\n")

	// Clean up old database for fresh demo
	dbPath := "./data/blockchain"
	os.RemoveAll(dbPath)

	// Demo 1: Initialize blockchain storage
	demoInitializeStorage(dbPath)

	// Demo 2: Store genesis block
	demoStoreGenesisBlock(dbPath)

	// Demo 3: Store multiple blocks (build a chain)
	demoStoreMultipleBlocks(dbPath)

	// Demo 4: Retrieve blocks by hash and height
	demoRetrieveBlocks(dbPath)

	// Demo 5: Transaction lookup
	demoTransactionLookup(dbPath)

	// Demo 6: Chain statistics
	demoChainStatistics(dbPath)

	// Demo 7: Database persistence (close and reopen)
	demoPersistence(dbPath)

	fmt.Println("\n=== All demos completed successfully! ===")
}

func demoInitializeStorage(dbPath string) {
	fmt.Println("--- Demo 1: Initialize Storage ---")

	// Open database
	bs, err := storage.NewBlockchainStorage(dbPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to open database: %v", err))
	}
	defer bs.Close()

	// Check if blockchain is empty
	isEmpty, err := bs.IsEmpty()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Database path: %s\n", dbPath)
	fmt.Printf("Blockchain is empty: %v\n", isEmpty)
	fmt.Println("✓ Storage initialized successfully\n")
}

func demoStoreGenesisBlock(dbPath string) {
	fmt.Println("--- Demo 2: Store Genesis Block ---")

	bs, err := storage.NewBlockchainStorage(dbPath)
	if err != nil {
		panic(err)
	}
	defer bs.Close()

	// Create genesis block
	genesisBlock := createGenesisBlock()

	// Compute block hash
	blockHash, _ := serialization.HashBlockHeader(&genesisBlock.Header)
	fmt.Printf("Genesis block hash: %s\n", blockHash)

	// Save to database (height 0)
	err = bs.SaveBlock(genesisBlock, 0)
	if err != nil {
		panic(fmt.Sprintf("Failed to save genesis block: %v", err))
	}

	fmt.Println("✓ Genesis block stored successfully")

	// Verify it was stored
	exists, _ := bs.HasBlock(blockHash)
	fmt.Printf("✓ Block exists in database: %v\n", exists)

	// Check chain state
	bestHash, _ := bs.GetBestBlockHash()
	bestHeight, _ := bs.GetBestBlockHeight()
	fmt.Printf("✓ Best block hash: %s\n", bestHash)
	fmt.Printf("✓ Best block height: %d\n\n", bestHeight)
}

func demoStoreMultipleBlocks(dbPath string) {
	fmt.Println("--- Demo 3: Store Multiple Blocks (Build Chain) ---")

	bs, err := storage.NewBlockchainStorage(dbPath)
	if err != nil {
		panic(err)
	}
	defer bs.Close()

	// Get genesis block (previous block)
	prevBlock, prevHeight, err := bs.GetBestBlock()
	if err != nil {
		panic(err)
	}
	prevHash, _ := serialization.HashBlockHeader(&prevBlock.Header)

	// Create and store 5 more blocks
	for i := 1; i <= 5; i++ {
		block := createBlock(prevHash, uint32(time.Now().Unix()), i)
		blockHash, _ := serialization.HashBlockHeader(&block.Header)

		err = bs.SaveBlock(block, prevHeight+1)
		if err != nil {
			panic(fmt.Sprintf("Failed to save block %d: %v", i, err))
		}

		fmt.Printf("Block %d stored: %s\n", i, blockHash.String()[:16]+"...")
		fmt.Printf("  Transactions: %d\n", len(block.Transactions))
		fmt.Printf("  Previous: %s\n", prevHash.String()[:16]+"...")

		// Update for next iteration
		prevBlock = block
		prevHash = blockHash
		prevHeight++
	}

	// Show final chain state
	count, _ := bs.GetBlockCount()
	fmt.Printf("\n✓ Total blocks in chain: %d\n", count)

	bestHash, _ := bs.GetBestBlockHash()
	bestHeight, _ := bs.GetBestBlockHeight()
	fmt.Printf("✓ Chain tip: %s (height %d)\n\n", bestHash.String()[:16]+"...", bestHeight)
}

func demoRetrieveBlocks(dbPath string) {
	fmt.Println("--- Demo 4: Retrieve Blocks ---")

	bs, err := storage.NewBlockchainStorage(dbPath)
	if err != nil {
		panic(err)
	}
	defer bs.Close()

	// Get block by height
	fmt.Println("Retrieving blocks by height:")
	for height := uint64(0); height <= 3; height++ {
		block, err := bs.GetBlockByHeight(height)
		if err != nil {
			fmt.Printf("  Height %d: ERROR - %v\n", height, err)
			continue
		}

		blockHash, _ := serialization.HashBlockHeader(&block.Header)
		fmt.Printf("  Height %d: %s (%d txs)\n", 
			height, 
			blockHash.String()[:16]+"...", 
			len(block.Transactions))
	}

	// Get current best block
	bestBlock, bestHeight, err := bs.GetBestBlock()
	if err != nil {
		panic(err)
	}
	bestHash, _ := serialization.HashBlockHeader(&bestBlock.Header)
	fmt.Printf("\nCurrent best block:\n")
	fmt.Printf("  Height: %d\n", bestHeight)
	fmt.Printf("  Hash: %s\n", bestHash)
	fmt.Printf("  Transactions: %d\n", len(bestBlock.Transactions))
	fmt.Printf("  Timestamp: %s\n\n", time.Unix(int64(bestBlock.Header.Timestamp), 0))
}

func demoTransactionLookup(dbPath string) {
	fmt.Println("--- Demo 5: Transaction Lookup ---")

	bs, err := storage.NewBlockchainStorage(dbPath)
	if err != nil {
		panic(err)
	}
	defer bs.Close()

	// Get a block with transactions
	height := uint64(1)
	block, err := bs.GetBlockByHeight(height)
	if err != nil {
		panic(err)
	}

	if len(block.Transactions) == 0 {
		fmt.Println("No transactions to look up\n")
		return
	}

	// Get first transaction
	tx := block.Transactions[0]
	txHash, _ := serialization.HashTransaction(&tx)

	fmt.Printf("Looking up transaction: %s\n", txHash.String()[:16]+"...")

	// Find which block contains this transaction
	blockHash, txIndex, err := bs.GetTransactionLocation(txHash)
	if err != nil {
		panic(err)
	}

	fmt.Printf("✓ Found in block: %s\n", blockHash.String()[:16]+"...")
	fmt.Printf("✓ Transaction index: %d\n", txIndex)
	fmt.Printf("✓ Block height: %d\n\n", height)
}

func demoChainStatistics(dbPath string) {
	fmt.Println("--- Demo 6: Chain Statistics ---")

	bs, err := storage.NewBlockchainStorage(dbPath)
	if err != nil {
		panic(err)
	}
	defer bs.Close()

	// Get chain info
	count, _ := bs.GetBlockCount()
	bestHash, _ := bs.GetBestBlockHash()
	bestHeight, _ := bs.GetBestBlockHeight()

	fmt.Printf("Blockchain Statistics:\n")
	fmt.Printf("  Total blocks: %d\n", count)
	fmt.Printf("  Current height: %d\n", bestHeight)
	fmt.Printf("  Chain tip: %s\n", bestHash)

	// Count total transactions
	totalTxs := 0
	for height := uint64(0); height <= bestHeight; height++ {
		block, err := bs.GetBlockByHeight(height)
		if err != nil {
			continue
		}
		totalTxs += len(block.Transactions)
	}

	fmt.Printf("  Total transactions: %d\n", totalTxs)
	fmt.Printf("  Average txs per block: %.2f\n\n", float64(totalTxs)/float64(count))
}

func demoPersistence(dbPath string) {
	fmt.Println("--- Demo 7: Database Persistence ---")

	// First session: store data
	fmt.Println("Session 1: Storing data...")
	bs1, err := storage.NewBlockchainStorage(dbPath)
	if err != nil {
		panic(err)
	}

	count1, _ := bs1.GetBlockCount()
	bestHash1, _ := bs1.GetBestBlockHash()
	fmt.Printf("  Blocks before closing: %d\n", count1)
	fmt.Printf("  Best hash: %s\n", bestHash1.String()[:16]+"...")

	// Close database
	bs1.Close()
	fmt.Println("  Database closed\n")

	// Second session: reopen and verify
	fmt.Println("Session 2: Reopening database...")
	bs2, err := storage.NewBlockchainStorage(dbPath)
	if err != nil {
		panic(err)
	}
	defer bs2.Close()

	count2, _ := bs2.GetBlockCount()
	bestHash2, _ := bs2.GetBestBlockHash()
	fmt.Printf("  Blocks after reopening: %d\n", count2)
	fmt.Printf("  Best hash: %s\n", bestHash2.String()[:16]+"...")

	// Verify data persisted
	if count1 == count2 && bestHash1 == bestHash2 {
		fmt.Println("\n✓ Data persisted successfully!")
		fmt.Println("✓ Blockchain survives database restart\n")
	} else {
		fmt.Println("\n✗ Data persistence failed!")
	}
}

// Helper: Create genesis block
func createGenesisBlock() *types.Block {
	// Create coinbase transaction
	coinbaseTx := types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      types.Hash{}, // No previous tx
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

	// Compute Merkle root
	txHash, _ := serialization.HashTransaction(&coinbaseTx)
	merkleRoot := crypto.ComputeMerkleRoot([]types.Hash{txHash})

	// Create block header
	header := types.BlockHeader{
		Version:       1,
		PrevBlockHash: types.Hash{}, // First block
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
	// Create coinbase transaction
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

	// Add 2-4 regular transactions
	numTxs := 2 + (blockNum % 3)
	for i := 0; i < numTxs; i++ {
		tx := types.Transaction{
			Version: 1,
			Inputs: []types.TxInput{
				{
					PrevTxHash:      crypto.DoubleSHA256([]byte(fmt.Sprintf("prev_%d_%d", blockNum, i))),
					OutputIndex:     uint32(i),
					SignatureScript: []byte(fmt.Sprintf("sig_%d_%d", blockNum, i)),
					Sequence:        0xFFFFFFFF,
				},
			},
			Outputs: []types.TxOutput{
				{
					Value:        int64(1000000 * (i + 1)),
					PubKeyScript: []byte(fmt.Sprintf("pubkey_%d_%d", blockNum, i)),
				},
			},
			LockTime: 0,
		}
		transactions = append(transactions, tx)
	}

	// Compute Merkle root
	var txHashes []types.Hash
	for _, tx := range transactions {
		hash, _ := serialization.HashTransaction(&tx)
		txHashes = append(txHashes, hash)
	}
	merkleRoot := crypto.ComputeMerkleRoot(txHashes)

	// Create block header
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