package mempool

import (
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// PriorityQueue manages transaction selection for block building
type PriorityQueue struct {
	mempool *Mempool
	entries []*MempoolEntry
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue(mempool *Mempool) *PriorityQueue {
	return &PriorityQueue{
		mempool: mempool,
		entries: make([]*MempoolEntry, 0),
	}
}

// Build builds the priority queue from mempool
func (pq *PriorityQueue) Build() {
	pq.mempool.mu.RLock()
	defer pq.mempool.mu.RUnlock()

	// Get all entries
	pq.entries = make([]*MempoolEntry, 0, len(pq.mempool.entries))
	for _, entry := range pq.mempool.entries {
		pq.entries = append(pq.entries, entry)
	}

	// Sort by ancestor fee rate (descending)
	pq.sortByAncestorFeeRate()
}

// sortByAncestorFeeRate sorts entries by ancestor fee rate (highest first)
func (pq *PriorityQueue) sortByAncestorFeeRate() {
	for i := 0; i < len(pq.entries)-1; i++ {
		for j := i + 1; j < len(pq.entries); j++ {
			rate1 := pq.getAncestorFeeRate(pq.entries[i])
			rate2 := pq.getAncestorFeeRate(pq.entries[j])

			if rate1 < rate2 {
				pq.entries[i], pq.entries[j] = pq.entries[j], pq.entries[i]
			}
		}
	}
}

// getAncestorFeeRate calculates the ancestor fee rate for an entry
func (pq *PriorityQueue) getAncestorFeeRate(entry *MempoolEntry) int64 {
	if entry.AncestorSize == 0 {
		return 0
	}
	return entry.AncestorFee / entry.AncestorSize
}

// SelectTransactions selects transactions for a block
func (pq *PriorityQueue) SelectTransactions(maxBlockSize int64) ([]*types.Transaction, error) {
	pq.Build()

	selected := make([]*types.Transaction, 0)
	selectedHashes := make(map[types.Hash]bool)
	currentSize := int64(0)

	// Reserve space for coinbase
	coinbaseSize := int64(200) // Approximate coinbase size
	maxBlockSize -= coinbaseSize

	for _, entry := range pq.entries {
		// Check if we have space
		if currentSize+entry.Size > maxBlockSize {
			continue
		}

		// Check if all parents are included
		allParentsIncluded := true
		for _, parentHash := range entry.Parents {
			if !selectedHashes[parentHash] {
				allParentsIncluded = false
				break
			}
		}

		if !allParentsIncluded {
			continue
		}

		// Add transaction
		selected = append(selected, entry.Tx)
		selectedHashes[entry.TxHash] = true
		currentSize += entry.Size
	}

	return selected, nil
}

// SelectTransactionsWithDependencies selects transactions ensuring dependencies are included
func (pq *PriorityQueue) SelectTransactionsWithDependencies(maxBlockSize int64) ([]*types.Transaction, error) {
	pq.Build()

	selected := make([]*types.Transaction, 0)
	selectedHashes := make(map[types.Hash]bool)
	currentSize := int64(0)

	// Reserve space for coinbase
	coinbaseSize := int64(200)
	maxBlockSize -= coinbaseSize

	// Process entries in priority order
	for _, entry := range pq.entries {
		if selectedHashes[entry.TxHash] {
			continue
		}

		// Calculate total size including unselected parents
		totalSize := entry.Size
		requiredParents := make([]*MempoolEntry, 0)

		for _, parentHash := range entry.Parents {
			if !selectedHashes[parentHash] {
				if parent, exists := pq.mempool.entries[parentHash]; exists {
					requiredParents = append(requiredParents, parent)
					totalSize += parent.Size
				}
			}
		}

		// Check if we have space for transaction and its parents
		if currentSize+totalSize > maxBlockSize {
			continue
		}

		// Add required parents first
		for _, parent := range requiredParents {
			selected = append(selected, parent.Tx)
			selectedHashes[parent.TxHash] = true
			currentSize += parent.Size
		}

		// Add the transaction
		selected = append(selected, entry.Tx)
		selectedHashes[entry.TxHash] = true
		currentSize += entry.Size
	}

	return selected, nil
}

// GetTopTransactions returns the top N transactions by fee rate
func (pq *PriorityQueue) GetTopTransactions(n int) []*MempoolEntry {
	pq.Build()

	if n > len(pq.entries) {
		n = len(pq.entries)
	}

	result := make([]*MempoolEntry, n)
	copy(result, pq.entries[:n])

	return result
}

// EstimateBlockFees estimates total fees for a block
func (pq *PriorityQueue) EstimateBlockFees(maxBlockSize int64) (int64, error) {
	txs, err := pq.SelectTransactions(maxBlockSize)
	if err != nil {
		return 0, err
	}

	totalFees := int64(0)
	pq.mempool.mu.RLock()
	defer pq.mempool.mu.RUnlock()

	// Find fees for each selected transaction
	for _, entry := range pq.mempool.entries {
		for _, tx := range txs {
			if entry.Tx == tx {
				totalFees += entry.Fee
				break
			}
		}
	}

	return totalFees, nil
}

// BlockTemplate represents a template for building a block
type BlockTemplate struct {
	Transactions []*types.Transaction
	TotalSize    int64
	TotalFees    int64
	TxCount      int
}

// CreateBlockTemplate creates a block template
func (pq *PriorityQueue) CreateBlockTemplate(maxBlockSize int64, coinbaseValue int64) (*BlockTemplate, error) {
	// Select transactions
	txs, err := pq.SelectTransactionsWithDependencies(maxBlockSize)
	if err != nil {
		return nil, fmt.Errorf("failed to select transactions: %w", err)
	}

	// Calculate totals
	totalSize := int64(0)
	totalFees := int64(0)

	pq.mempool.mu.RLock()
	for _, tx := range txs {
		// Find entry in mempool
		for _, entry := range pq.mempool.entries {
			if entry.Tx == tx {
				totalSize += entry.Size
				totalFees += entry.Fee
				break
			}
		}
	}
	pq.mempool.mu.RUnlock()

	return &BlockTemplate{
		Transactions: txs,
		TotalSize:    totalSize,
		TotalFees:    totalFees,
		TxCount:      len(txs),
	}, nil
}

// PackageSelector selects transaction packages (parent + children together)
type PackageSelector struct {
	mempool *Mempool
}

// NewPackageSelector creates a new package selector
func NewPackageSelector(mempool *Mempool) *PackageSelector {
	return &PackageSelector{
		mempool: mempool,
	}
}

// SelectPackages selects transaction packages for block inclusion
func (ps *PackageSelector) SelectPackages(maxBlockSize int64) ([][]*types.Transaction, error) {
	ps.mempool.mu.RLock()
	defer ps.mempool.mu.RUnlock()

	packages := make([][]*types.Transaction, 0)
	processed := make(map[types.Hash]bool)
	currentSize := int64(0)

	// Find root transactions (no parents in mempool)
	roots := make([]*MempoolEntry, 0)
	for _, entry := range ps.mempool.entries {
		if len(entry.Parents) == 0 {
			roots = append(roots, entry)
		}
	}

	// Sort roots by fee rate
	for i := 0; i < len(roots)-1; i++ {
		for j := i + 1; j < len(roots); j++ {
			if roots[i].FeeRate < roots[j].FeeRate {
				roots[i], roots[j] = roots[j], roots[i]
			}
		}
	}

	// Build packages from roots
	for _, root := range roots {
		if processed[root.TxHash] {
			continue
		}

		// Collect package (root + all descendants)
		pkg := ps.collectPackage(root, processed)
		pkgSize := ps.calculatePackageSize(pkg)

		// Check if package fits
		if currentSize+pkgSize > maxBlockSize {
			continue
		}

		packages = append(packages, pkg)
		currentSize += pkgSize

		// Mark all transactions as processed
		for _, tx := range pkg {
			for hash, entry := range ps.mempool.entries {
				if entry.Tx == tx {
					processed[hash] = true
					break
				}
			}
		}
	}

	return packages, nil
}

// collectPackage collects a transaction and all its descendants
func (ps *PackageSelector) collectPackage(root *MempoolEntry, processed map[types.Hash]bool) []*types.Transaction {
	pkg := make([]*types.Transaction, 0)
	pkg = append(pkg, root.Tx)

	// Recursively add children
	for _, childHash := range root.Children {
		if processed[childHash] {
			continue
		}

		if child, exists := ps.mempool.entries[childHash]; exists {
			childPkg := ps.collectPackage(child, processed)
			pkg = append(pkg, childPkg...)
		}
	}

	return pkg
}

// calculatePackageSize calculates total size of a package
func (ps *PackageSelector) calculatePackageSize(pkg []*types.Transaction) int64 {
	size := int64(0)
	for _, tx := range pkg {
		size += CalculateTransactionSize(tx)
	}
	return size
}
