package main

import (
	"fmt"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/crypto"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

func main() {
	fmt.Println("=== Bitcoin Learning - Milestone 1 ===")
	fmt.Println("Data Models & Hashing Fundamentals\n")

	// Demo 1: Create and hash a simple transaction
	demoSimpleTransaction()
	
	// Demo 2: Build a block with Merkle tree
	demoBlockWithMerkleTree()
	
	// Demo 3: Genesis block recreation
	demoGenesisBlock()
}

func demoSimpleTransaction() {
	fmt.Println("--- Demo 1: Simple Transaction ---")
	
	// Create a coinbase transaction (mining reward)
	tx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      types.Hash{}, // No previous tx (coinbase)
				OutputIndex:     0xFFFFFFFF,
				SignatureScript: []byte("Block reward - Alice mines block 1"),
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        5000000000, // 50 BTC
				PubKeyScript: []byte("Pay to Alice"),
			},
		},
		LockTime: 0,
	}

	// Compute transaction hash
	txHash, err := serialization.HashTransaction(tx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Transaction ID: %s\n", txHash)
	fmt.Printf("Output Value: %d satoshis (%.8f BTC)\n", 
		tx.Outputs[0].Value, 
		float64(tx.Outputs[0].Value)/100000000)
	fmt.Println()
}

func demoBlockWithMerkleTree() {
	fmt.Println("--- Demo 2: Block with Merkle Tree ---")
	
	// Create 3 transactions
	txs := []*types.Transaction{
		createTransaction("Coinbase reward", 5000000000),
		createTransaction("Alice -> Bob: 1 BTC", 100000000),
		createTransaction("Bob -> Carol: 0.5 BTC", 50000000),
	}

	// Compute transaction hashes
	var txHashes []types.Hash
	for i, tx := range txs {
		hash, _ := serialization.HashTransaction(tx)
		txHashes = append(txHashes, hash)
		fmt.Printf("Transaction %d: %s\n", i+1, hash)
	}

	// Compute Merkle root
	merkleRoot := crypto.ComputeMerkleRoot(txHashes)
	fmt.Printf("\nMerkle Root: %s\n", merkleRoot)

	// Build full tree for visualization
	tree := crypto.BuildMerkleTree(txHashes)
	fmt.Printf("\nMerkle Tree Levels: %d\n", len(tree))
	for i, level := range tree {
		fmt.Printf("  Level %d: %d hashes\n", i, len(level))
	}
	fmt.Println()
}

func demoGenesisBlock() {
	fmt.Println("--- Demo 3: Genesis Block Recreation ---")
	
	// Create genesis coinbase transaction
	genesisTx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:  types.Hash{},
				OutputIndex: 0xFFFFFFFF,
				SignatureScript: []byte{
					0x04, 0xff, 0xff, 0x00, 0x1d, 0x01, 0x04,
					// Satoshi's message: "The Times 03/Jan/2009..."
				},
				Sequence: 0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        5000000000,
				PubKeyScript: []byte{0x41, 0x04}, // Simplified
			},
		},
		LockTime: 0,
	}

	genesisTxHash, _ := serialization.HashTransaction(genesisTx)
	fmt.Printf("Genesis Coinbase TX: %s\n", genesisTxHash)

	// Create genesis block header
	genesisHeader := &types.BlockHeader{
		Version:       1,
		PrevBlockHash: types.Hash{}, // First block
		MerkleRoot:    genesisTxHash,
		Timestamp:     1231006505, // Jan 3, 2009 18:15:05 GMT
		Bits:          0x1d00ffff,
		Nonce:         2083236893,
	}

	genesisHash, _ := serialization.HashBlockHeader(genesisHeader)
	fmt.Printf("Genesis Block Hash: %s\n", genesisHash)
	
	// Display in human-readable format
	fmt.Println("\nGenesis Block Details:")
	fmt.Printf("  Version: %d\n", genesisHeader.Version)
	fmt.Printf("  Previous Hash: %s\n", genesisHeader.PrevBlockHash)
	fmt.Printf("  Merkle Root: %s\n", genesisHeader.MerkleRoot)
	fmt.Printf("  Timestamp: %d (Jan 3, 2009)\n", genesisHeader.Timestamp)
	fmt.Printf("  Difficulty: 0x%x\n", genesisHeader.Bits)
	fmt.Printf("  Nonce: %d\n", genesisHeader.Nonce)
	
	// Count leading zero bytes
	leadingZeros := countLeadingZeroBytes(genesisHash)
	fmt.Printf("\nLeading zero bytes: %d\n", leadingZeros)
	fmt.Println("(This proves Proof-of-Work was done)")
}

// Helper function to create test transactions
func createTransaction(description string, value int64) *types.Transaction {
	return &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      crypto.DoubleSHA256([]byte(description)),
				OutputIndex:     0,
				SignatureScript: []byte(description),
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        value,
				PubKeyScript: []byte("pubkey"),
			},
		},
		LockTime: 0,
	}
}

// Count leading zero bytes in hash
func countLeadingZeroBytes(h types.Hash) int {
	count := 0
	for i := len(h) - 1; i >= 0; i-- {
		if h[i] == 0 {
			count++
		} else {
			break
		}
	}
	return count
}