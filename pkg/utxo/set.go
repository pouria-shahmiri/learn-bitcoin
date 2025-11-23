package utxo

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// UTXOSet represents the set of all unspent transaction outputs
type UTXOSet struct {
	utxos map[string]*UTXO // Map: outpoint -> UTXO
	mu    sync.RWMutex
}

// NewUTXOSet creates a new empty UTXO set
func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		utxos: make(map[string]*UTXO),
	}
}

// Add adds a UTXO to the set
func (us *UTXOSet) Add(utxo *UTXO) error {
	us.mu.Lock()
	defer us.mu.Unlock()

	key := utxo.OutPoint().String()

	// Check if already exists
	if _, exists := us.utxos[key]; exists {
		return fmt.Errorf("UTXO already exists: %s", key)
	}

	us.utxos[key] = utxo.Clone()
	return nil
}

// Remove removes a UTXO from the set (when it's spent)
func (us *UTXOSet) Remove(outpoint OutPoint) error {
	us.mu.Lock()
	defer us.mu.Unlock()

	key := outpoint.String()

	if _, exists := us.utxos[key]; !exists {
		return fmt.Errorf("UTXO not found: %s", key)
	}

	delete(us.utxos, key)
	return nil
}

// Get retrieves a UTXO from the set
func (us *UTXOSet) Get(outpoint OutPoint) (*UTXO, error) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	key := outpoint.String()
	utxo, exists := us.utxos[key]

	if !exists {
		return nil, fmt.Errorf("UTXO not found: %s", key)
	}

	return utxo.Clone(), nil
}

// Exists checks if a UTXO exists in the set
func (us *UTXOSet) Exists(outpoint OutPoint) bool {
	us.mu.RLock()
	defer us.mu.RUnlock()

	key := outpoint.String()
	_, exists := us.utxos[key]
	return exists
}

// Size returns the number of UTXOs in the set
func (us *UTXOSet) Size() int {
	us.mu.RLock()
	defer us.mu.RUnlock()

	return len(us.utxos)
}

// TotalValue returns the total value of all UTXOs
func (us *UTXOSet) TotalValue() int64 {
	us.mu.RLock()
	defer us.mu.RUnlock()

	total := int64(0)
	for _, utxo := range us.utxos {
		total += utxo.Value()
	}

	return total
}

// FindByScript returns all UTXOs matching a script (for wallet balance)
func (us *UTXOSet) FindByScript(script []byte) []*UTXO {
	us.mu.RLock()
	defer us.mu.RUnlock()

	var result []*UTXO

	for _, utxo := range us.utxos {
		if bytes.Equal(utxo.Output.PubKeyScript, script) {
			result = append(result, utxo.Clone())
		}
	}

	return result
}

// ApplyTransaction applies a transaction to the UTXO set
// - Removes spent UTXOs (inputs)
// - Adds new UTXOs (outputs)
func (us *UTXOSet) ApplyTransaction(tx *types.Transaction, txHash types.Hash, height uint64, isCoinbase bool) error {
	us.mu.Lock()
	defer us.mu.Unlock()

	// First, remove spent UTXOs (inputs)
	if !isCoinbase {
		for _, input := range tx.Inputs {
			outpoint := NewOutPoint(input.PrevTxHash, input.OutputIndex)
			key := outpoint.String()

			if _, exists := us.utxos[key]; !exists {
				return fmt.Errorf("trying to spend non-existent UTXO: %s", key)
			}

			delete(us.utxos, key)
		}
	}

	// Then, add new UTXOs (outputs)
	for i, output := range tx.Outputs {
		utxo := NewUTXO(txHash, uint32(i), output, height, isCoinbase)
		key := utxo.OutPoint().String()
		us.utxos[key] = utxo
	}

	return nil
}

// RevertTransaction reverts a transaction from the UTXO set
// Used during blockchain reorganization
func (us *UTXOSet) RevertTransaction(tx *types.Transaction, txHash types.Hash) error {
	us.mu.Lock()
	defer us.mu.Unlock()

	// Remove outputs created by this transaction
	for i := range tx.Outputs {
		outpoint := NewOutPoint(txHash, uint32(i))
		key := outpoint.String()
		delete(us.utxos, key)
	}

	// Note: Re-adding spent UTXOs requires the previous UTXOs
	// This is typically done by the caller who has the full block data

	return nil
}

// Clone creates a deep copy of the UTXO set
func (us *UTXOSet) Clone() *UTXOSet {
	us.mu.RLock()
	defer us.mu.RUnlock()

	clone := NewUTXOSet()

	for key, utxo := range us.utxos {
		clone.utxos[key] = utxo.Clone()
	}

	return clone
}

// Clear removes all UTXOs from the set
func (us *UTXOSet) Clear() {
	us.mu.Lock()
	defer us.mu.Unlock()

	us.utxos = make(map[string]*UTXO)
}

// GetAll returns all UTXOs (for iteration)
func (us *UTXOSet) GetAll() []*UTXO {
	us.mu.RLock()
	defer us.mu.RUnlock()

	result := make([]*UTXO, 0, len(us.utxos))

	for _, utxo := range us.utxos {
		result = append(result, utxo.Clone())
	}

	return result
}

// Statistics returns statistics about the UTXO set
type Statistics struct {
	Count         int
	TotalValue    int64
	CoinbaseCount int
	CoinbaseValue int64
	AverageValue  int64
}

// GetStatistics returns statistics about the UTXO set
func (us *UTXOSet) GetStatistics() Statistics {
	us.mu.RLock()
	defer us.mu.RUnlock()

	stats := Statistics{}

	for _, utxo := range us.utxos {
		stats.Count++
		stats.TotalValue += utxo.Value()

		if utxo.IsCoinbase {
			stats.CoinbaseCount++
			stats.CoinbaseValue += utxo.Value()
		}
	}

	if stats.Count > 0 {
		stats.AverageValue = stats.TotalValue / int64(stats.Count)
	}

	return stats
}
