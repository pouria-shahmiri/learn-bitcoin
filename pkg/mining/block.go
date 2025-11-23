package mining

import (
	"fmt"
	"time"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/crypto"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/mempool"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// BlockTemplate contains all data needed to mine a block
type BlockTemplate struct {
	Version       int32
	PrevBlockHash types.Hash
	Transactions  []types.Transaction
	Timestamp     uint32
	Bits          uint32
	Height        uint64
	TotalFees     int64
}

// BlockBuilder constructs blocks from mempool
type BlockBuilder struct {
	mempool *mempool.Mempool
}

// NewBlockBuilder creates a new block builder
func NewBlockBuilder(mp *mempool.Mempool) *BlockBuilder {
	return &BlockBuilder{
		mempool: mp,
	}
}

// CreateBlockTemplate creates a template ready for mining
func (bb *BlockBuilder) CreateBlockTemplate(
	prevBlockHash types.Hash,
	height uint64,
	minerAddress string,
	difficulty uint32,
) (*BlockTemplate, error) {

	// 1. Select transactions from mempool
	selectedTxs, totalFees := bb.selectTransactions()

	// 2. Create coinbase transaction
	coinbaseTx, err := CreateCoinbase(height, totalFees, minerAddress, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create coinbase: %w", err)
	}

	// 3. Assemble all transactions (coinbase first)
	allTxs := make([]types.Transaction, 0, len(selectedTxs)+1)
	allTxs = append(allTxs, *coinbaseTx)
	allTxs = append(allTxs, selectedTxs...)

	// 4. Create template
	template := &BlockTemplate{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Transactions:  allTxs,
		Timestamp:     uint32(time.Now().Unix()),
		Bits:          difficulty,
		Height:        height,
		TotalFees:     totalFees,
	}

	return template, nil
}

// selectTransactions selects transactions from mempool
// Returns transactions and total fees
func (bb *BlockBuilder) selectTransactions() ([]types.Transaction, int64) {
	// Create priority queue
	pq := mempool.NewPriorityQueue(bb.mempool)

	// Select transactions (1MB max block size)
	maxBlockSize := int64(1000000)
	selectedTxPtrs, err := pq.SelectTransactionsWithDependencies(maxBlockSize)
	if err != nil {
		// If selection fails, return empty
		return []types.Transaction{}, 0
	}

	// Convert from []*Transaction to []Transaction
	selected := make([]types.Transaction, len(selectedTxPtrs))
	for i, txPtr := range selectedTxPtrs {
		selected[i] = *txPtr
	}

	// Calculate total fees
	totalFees := int64(0)
	entries := bb.mempool.GetAllTransactions()
	for _, entry := range entries {
		for _, tx := range selectedTxPtrs {
			if entry.Tx == tx {
				totalFees += entry.Fee
				break
			}
		}
	}

	return selected, totalFees
}

// BuildBlock creates a complete block from template
func BuildBlock(template *BlockTemplate, nonce uint32) (*types.Block, error) {
	// Calculate merkle root
	merkleRoot, err := calculateMerkleRoot(template.Transactions)
	if err != nil {
		return nil, err
	}

	// Create header
	header := types.BlockHeader{
		Version:       template.Version,
		PrevBlockHash: template.PrevBlockHash,
		MerkleRoot:    merkleRoot,
		Timestamp:     template.Timestamp,
		Bits:          template.Bits,
		Nonce:         nonce,
	}

	// Create block
	block := &types.Block{
		Header:       header,
		Transactions: template.Transactions,
	}

	return block, nil
}

// calculateMerkleRoot computes merkle root from transactions
func calculateMerkleRoot(txs []types.Transaction) (types.Hash, error) {
	if len(txs) == 0 {
		return types.Hash{}, fmt.Errorf("no transactions")
	}

	// Hash all transactions
	txHashes := make([]types.Hash, len(txs))
	for i, tx := range txs {
		hash, err := serialization.HashTransaction(&tx)
		if err != nil {
			return types.Hash{}, err
		}
		txHashes[i] = hash
	}

	// Compute merkle root
	return crypto.ComputeMerkleRoot(txHashes), nil
}

/*
**Block Construction Process:**

1. **Select Transactions:**
   - Get transactions from mempool
   - Sort by fee rate (sat/byte)
   - Select highest paying transactions
   - Respect block size limit (1MB)

2. **Create Coinbase:**
   - Calculate block reward
   - Add transaction fees
   - Create coinbase transaction

3. **Calculate Merkle Root:**
   - Hash all transactions
   - Build merkle tree
   - Get root hash

4. **Assemble Header:**
   - Version
   - Previous block hash
   - Merkle root
   - Timestamp
   - Difficulty (bits)
   - Nonce (starts at 0)

5. **Ready to Mine:**
   - Increment nonce
   - Hash header
   - Check if hash meets difficulty
   - Repeat until valid

**Example:**
```
Mempool has 1000 transactions
Block size limit: 1MB

Selection:
1. Sort by fee rate
2. Pick highest paying
3. Stop at 1MB or 1000 txs

Result:
- 500 transactions selected
- Total fees: 0.5 BTC
- Block reward: 6.25 BTC
- Miner gets: 6.75 BTC
```

**Visual:**
```
Block Template:
┌─────────────────────────┐
│ Coinbase (reward+fees)  │
├─────────────────────────┤
│ Transaction 1 (high fee)│
│ Transaction 2           │
│ Transaction 3           │
│ ...                     │
│ Transaction N           │
└─────────────────────────┘
        ↓
   Calculate Merkle Root
        ↓
   Create Header
        ↓
   Ready to Mine!
```
*/
