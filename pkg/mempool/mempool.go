package mempool

import (
	"fmt"
	"sync"
	"time"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// MempoolEntry represents a transaction in the mempool with metadata
type MempoolEntry struct {
	Tx           *types.Transaction
	TxHash       types.Hash
	Size         int64        // Transaction size in bytes
	Fee          int64        // Transaction fee in satoshis
	FeeRate      int64        // Fee per byte (satoshis/byte)
	Time         int64        // Time added to mempool (Unix timestamp)
	Height       uint64       // Block height when added
	Parents      []types.Hash // Parent transactions (dependencies)
	Children     []types.Hash // Child transactions
	AncestorFee  int64        // Total fee including ancestors
	AncestorSize int64        // Total size including ancestors
}

// Mempool manages the transaction pool
type Mempool struct {
	mu sync.RWMutex

	// Main storage: txid -> entry
	entries map[types.Hash]*MempoolEntry

	// Index by outpoint for quick dependency lookup
	// outpoint -> spending transaction hash
	spentOutputs map[types.OutPoint]types.Hash

	// Configuration
	maxSize       int64  // Maximum mempool size in bytes
	minFeeRate    int64  // Minimum fee rate (satoshis/byte)
	maxTxAge      int64  // Maximum transaction age in seconds
	currentSize   int64  // Current mempool size in bytes
	currentHeight uint64 // Current blockchain height
}

// NewMempool creates a new mempool
func NewMempool(maxSize int64, minFeeRate int64, maxTxAge int64) *Mempool {
	return &Mempool{
		entries:       make(map[types.Hash]*MempoolEntry),
		spentOutputs:  make(map[types.OutPoint]types.Hash),
		maxSize:       maxSize,
		minFeeRate:    minFeeRate,
		maxTxAge:      maxTxAge,
		currentSize:   0,
		currentHeight: 0,
	}
}

// Add adds a transaction to the mempool
func (m *Mempool) Add(tx *types.Transaction, fee int64, height uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Calculate transaction hash
	txHash, err := serialization.HashTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to hash transaction: %w", err)
	}

	// Check if already in mempool
	if _, exists := m.entries[txHash]; exists {
		return fmt.Errorf("transaction already in mempool")
	}

	// Calculate transaction size
	size := CalculateTransactionSize(tx)

	// Calculate fee rate
	feeRate := fee / size
	if feeRate < m.minFeeRate {
		return fmt.Errorf("fee rate too low: %d < %d", feeRate, m.minFeeRate)
	}

	// Check for conflicts (double-spends)
	for _, input := range tx.Inputs {
		outpoint := types.OutPoint{
			Hash:  input.PrevTxHash,
			Index: input.OutputIndex,
		}

		if existingTxHash, exists := m.spentOutputs[outpoint]; exists {
			// Check if this is Replace-By-Fee (RBF)
			existingEntry := m.entries[existingTxHash]
			if !m.canReplace(existingEntry, fee, feeRate) {
				return fmt.Errorf("output already spent by %s", existingTxHash.String())
			}

			// Remove the existing transaction (RBF)
			m.removeTransaction(existingTxHash)
		}
	}

	// Check mempool size limit
	if m.currentSize+size > m.maxSize {
		// Try to evict low-fee transactions
		if !m.evictTransactions(size) {
			return fmt.Errorf("mempool full and cannot evict transactions")
		}
	}

	// Find parent transactions (dependencies)
	parents := m.findParents(tx)

	// Create entry
	entry := &MempoolEntry{
		Tx:       tx,
		TxHash:   txHash,
		Size:     size,
		Fee:      fee,
		FeeRate:  feeRate,
		Time:     time.Now().Unix(),
		Height:   height,
		Parents:  parents,
		Children: make([]types.Hash, 0),
	}

	// Calculate ancestor fee and size
	entry.AncestorFee, entry.AncestorSize = m.calculateAncestorMetrics(entry)

	// Add to mempool
	m.entries[txHash] = entry
	m.currentSize += size

	// Update spent outputs index
	for _, input := range tx.Inputs {
		outpoint := types.OutPoint{
			Hash:  input.PrevTxHash,
			Index: input.OutputIndex,
		}
		m.spentOutputs[outpoint] = txHash
	}

	// Update parent-child relationships
	for _, parentHash := range parents {
		if parent, exists := m.entries[parentHash]; exists {
			parent.Children = append(parent.Children, txHash)
		}
	}

	return nil
}

// Remove removes a transaction from the mempool
func (m *Mempool) Remove(txHash types.Hash) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.removeTransaction(txHash)
}

// removeTransaction removes a transaction (internal, no lock)
func (m *Mempool) removeTransaction(txHash types.Hash) error {
	entry, exists := m.entries[txHash]
	if !exists {
		return fmt.Errorf("transaction not in mempool")
	}

	// Remove from entries
	delete(m.entries, txHash)
	m.currentSize -= entry.Size

	// Remove from spent outputs index
	for _, input := range entry.Tx.Inputs {
		outpoint := types.OutPoint{
			Hash:  input.PrevTxHash,
			Index: input.OutputIndex,
		}
		delete(m.spentOutputs, outpoint)
	}

	// Update parent-child relationships
	for _, parentHash := range entry.Parents {
		if parent, exists := m.entries[parentHash]; exists {
			parent.Children = removeHash(parent.Children, txHash)
		}
	}

	// Recursively remove children
	for _, childHash := range entry.Children {
		m.removeTransaction(childHash)
	}

	return nil
}

// Get retrieves a transaction from the mempool
func (m *Mempool) Get(txHash types.Hash) (*MempoolEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.entries[txHash]
	if !exists {
		return nil, fmt.Errorf("transaction not in mempool")
	}

	return entry, nil
}

// Exists checks if a transaction is in the mempool
func (m *Mempool) Exists(txHash types.Hash) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.entries[txHash]
	return exists
}

// Size returns the number of transactions in the mempool
func (m *Mempool) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.entries)
}

// GetMemoryUsage returns current memory usage in bytes
func (m *Mempool) GetMemoryUsage() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.currentSize
}

// findParents finds parent transactions in the mempool
func (m *Mempool) findParents(tx *types.Transaction) []types.Hash {
	parents := make([]types.Hash, 0)

	for _, input := range tx.Inputs {
		// Skip coinbase inputs
		if input.OutputIndex == 0xFFFFFFFF {
			continue
		}

		// Check if parent is in mempool
		if _, exists := m.entries[input.PrevTxHash]; exists {
			parents = append(parents, input.PrevTxHash)
		}
	}

	return parents
}

// calculateAncestorMetrics calculates total fee and size including ancestors
func (m *Mempool) calculateAncestorMetrics(entry *MempoolEntry) (int64, int64) {
	totalFee := entry.Fee
	totalSize := entry.Size

	// Add ancestor fees and sizes
	for _, parentHash := range entry.Parents {
		if parent, exists := m.entries[parentHash]; exists {
			totalFee += parent.AncestorFee
			totalSize += parent.AncestorSize
		}
	}

	return totalFee, totalSize
}

// canReplace checks if a transaction can be replaced (RBF)
func (m *Mempool) canReplace(existing *MempoolEntry, newFee int64, newFeeRate int64) bool {
	// New transaction must pay higher fee
	if newFee <= existing.Fee {
		return false
	}

	// New transaction must have higher fee rate
	if newFeeRate <= existing.FeeRate {
		return false
	}

	// Additional fee must be at least minFeeRate * size
	additionalFee := newFee - existing.Fee
	if additionalFee < m.minFeeRate*existing.Size {
		return false
	}

	return true
}

// evictTransactions evicts low-fee transactions to make room
func (m *Mempool) evictTransactions(neededSize int64) bool {
	// Get all transactions sorted by fee rate (ascending)
	entries := make([]*MempoolEntry, 0, len(m.entries))
	for _, entry := range m.entries {
		entries = append(entries, entry)
	}

	// Sort by fee rate (lowest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].FeeRate > entries[j].FeeRate {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Evict transactions until we have enough space
	freedSize := int64(0)
	for _, entry := range entries {
		m.removeTransaction(entry.TxHash)
		freedSize += entry.Size

		if freedSize >= neededSize {
			return true
		}
	}

	return false
}

// ExpireTransactions removes transactions older than maxTxAge
func (m *Mempool) ExpireTransactions() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.maxTxAge <= 0 {
		return 0
	}

	currentTime := time.Now().Unix()
	expired := make([]types.Hash, 0)

	for txHash, entry := range m.entries {
		age := currentTime - entry.Time
		if age > m.maxTxAge {
			expired = append(expired, txHash)
		}
	}

	// Remove expired transactions
	for _, txHash := range expired {
		m.removeTransaction(txHash)
	}

	return len(expired)
}

// GetAllTransactions returns all transactions in the mempool
func (m *Mempool) GetAllTransactions() []*MempoolEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := make([]*MempoolEntry, 0, len(m.entries))
	for _, entry := range m.entries {
		entries = append(entries, entry)
	}

	return entries
}

// Clear removes all transactions from the mempool
func (m *Mempool) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = make(map[types.Hash]*MempoolEntry)
	m.spentOutputs = make(map[types.OutPoint]types.Hash)
	m.currentSize = 0
}

// UpdateHeight updates the current blockchain height
func (m *Mempool) UpdateHeight(height uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentHeight = height
}

// Helper function to remove a hash from a slice
func removeHash(slice []types.Hash, hash types.Hash) []types.Hash {
	result := make([]types.Hash, 0, len(slice))
	for _, h := range slice {
		if h != hash {
			result = append(result, h)
		}
	}
	return result
}
