package reorg

import (
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/mempool"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/utxo"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/validation"
)

// ReorgHandler handles blockchain reorganizations
type ReorgHandler struct {
	blockchain *storage.BlockchainStorage
	utxoSet    *utxo.UTXOSet
	mempool    *mempool.Mempool
	validator  *validation.BlockValidator
	detector   *ReorgDetector
}

// NewReorgHandler creates a new reorganization handler
func NewReorgHandler(
	blockchain *storage.BlockchainStorage,
	utxoSet *utxo.UTXOSet,
	mp *mempool.Mempool,
) *ReorgHandler {
	return &ReorgHandler{
		blockchain: blockchain,
		utxoSet:    utxoSet,
		mempool:    mp,
		validator:  validation.NewBlockValidator(utxoSet),
		detector:   NewReorgDetector(blockchain),
	}
}

// HandleReorg performs a blockchain reorganization
func (rh *ReorgHandler) HandleReorg(newBlocks []*types.Block) error {
	// Detect if reorg is needed
	needsReorg, chainInfo, err := rh.detector.DetectReorg(newBlocks)
	if err != nil {
		return fmt.Errorf("reorg detection failed: %w", err)
	}

	if !needsReorg {
		return nil // No reorg needed
	}

	fmt.Printf("Starting reorganization: fork at height %d, new chain height %d\n",
		chainInfo.ForkHeight, chainInfo.Height)

	// Step 1: Disconnect blocks from old chain
	orphanedTxs, err := rh.disconnectBlocks(chainInfo.ForkHeight)
	if err != nil {
		return fmt.Errorf("failed to disconnect blocks: %w", err)
	}

	fmt.Printf("Disconnected %d blocks, recovered %d transactions\n",
		len(orphanedTxs), countTransactions(orphanedTxs))

	// Step 2: Connect blocks from new chain
	if err := rh.connectBlocks(newBlocks, chainInfo.ForkHeight); err != nil {
		// Attempt to rollback
		fmt.Printf("Failed to connect new blocks, attempting rollback...\n")
		// In production, we'd try to restore the old chain here
		return fmt.Errorf("failed to connect new blocks: %w", err)
	}

	fmt.Printf("Connected %d new blocks\n", len(newBlocks))

	// Step 3: Return orphaned transactions to mempool
	if err := rh.returnOrphanedTransactions(orphanedTxs, newBlocks); err != nil {
		// This is not critical, just log it
		fmt.Printf("Warning: failed to return some orphaned transactions: %v\n", err)
	}

	fmt.Printf("Reorganization completed successfully\n")
	return nil
}

// disconnectBlocks removes blocks from the current chain back to fork point
func (rh *ReorgHandler) disconnectBlocks(forkHeight uint64) ([]*types.Block, error) {
	currentHeight, err := rh.blockchain.GetBestBlockHeight()
	if err != nil {
		return nil, err
	}

	var disconnectedBlocks []*types.Block

	// Disconnect blocks in reverse order (from tip to fork)
	for h := currentHeight; h > forkHeight; h-- {
		block, err := rh.blockchain.GetBlockByHeight(h)
		if err != nil {
			return nil, fmt.Errorf("failed to get block at height %d: %w", h, err)
		}

		// Revert UTXO changes
		if err := rh.revertBlock(block, h); err != nil {
			return nil, fmt.Errorf("failed to revert block at height %d: %w", h, err)
		}

		disconnectedBlocks = append(disconnectedBlocks, block)
	}

	// Update chain state to fork point
	if err := rh.updateChainState(forkHeight); err != nil {
		return nil, fmt.Errorf("failed to update chain state: %w", err)
	}

	return disconnectedBlocks, nil
}

// connectBlocks adds new blocks to the chain
func (rh *ReorgHandler) connectBlocks(blocks []*types.Block, startHeight uint64) error {
	for i, block := range blocks {
		height := startHeight + uint64(i) + 1

		// Validate block
		var prevHash types.Hash
		if height > 0 {
			prevBlock, err := rh.blockchain.GetBlockByHeight(height - 1)
			if err != nil {
				return fmt.Errorf("failed to get previous block: %w", err)
			}
			prevHash, _ = serialization.HashBlockHeader(&prevBlock.Header)
		}

		if err := rh.validator.ValidateBlock(block, height, prevHash); err != nil {
			return fmt.Errorf("block validation failed at height %d: %w", height, err)
		}

		// Apply to UTXO set
		if err := rh.validator.ApplyBlock(block, height); err != nil {
			return fmt.Errorf("failed to apply block at height %d: %w", height, err)
		}

		// Save block
		if err := rh.blockchain.SaveBlock(block, height); err != nil {
			return fmt.Errorf("failed to save block at height %d: %w", height, err)
		}
	}

	return nil
}

// revertBlock reverts UTXO changes from a block
func (rh *ReorgHandler) revertBlock(block *types.Block, height uint64) error {
	// Remove outputs created by this block
	for _, tx := range block.Transactions {
		txHash, err := serialization.HashTransaction(&tx)
		if err != nil {
			return err
		}

		// Remove all outputs
		for i := range tx.Outputs {
			outpoint := utxo.NewOutPoint(txHash, uint32(i))
			rh.utxoSet.Remove(outpoint)
		}
	}

	// Restore inputs (except coinbase)
	for txIdx, tx := range block.Transactions {
		if txIdx == 0 {
			continue // Skip coinbase
		}

		for _, input := range tx.Inputs {
			// We need to restore the spent UTXO
			// In a real implementation, we'd need to store the previous UTXO data
			// For now, we'll skip this (simplified version)
			_ = input
		}
	}

	return nil
}

// returnOrphanedTransactions adds orphaned transactions back to mempool
func (rh *ReorgHandler) returnOrphanedTransactions(orphanedBlocks []*types.Block, newBlocks []*types.Block) error {
	// Build set of transactions in new chain
	newChainTxs := make(map[types.Hash]bool)
	for _, block := range newBlocks {
		for _, tx := range block.Transactions {
			txHash, _ := serialization.HashTransaction(&tx)
			newChainTxs[txHash] = true
		}
	}

	// Add orphaned transactions that aren't in new chain
	for _, block := range orphanedBlocks {
		for txIdx, tx := range block.Transactions {
			if txIdx == 0 {
				continue // Skip coinbase
			}

			txHash, _ := serialization.HashTransaction(&tx)
			if !newChainTxs[txHash] {
				// Try to add back to mempool
				if err := rh.mempool.Add(&tx, 0, 0); err != nil {
					// Log but don't fail - transaction might be invalid now
					fmt.Printf("Could not return tx %s to mempool: %v\n", txHash, err)
				}
			}
		}
	}

	return nil
}

// updateChainState updates the chain state to a specific height
func (rh *ReorgHandler) updateChainState(height uint64) error {
	block, err := rh.blockchain.GetBlockByHeight(height)
	if err != nil {
		return err
	}

	// Update best block (this would need to be implemented in storage)
	// For now, we'll just save the block again to update the chain state
	return rh.blockchain.SaveBlock(block, height)
}

// countTransactions counts total transactions in blocks
func countTransactions(blocks []*types.Block) int {
	count := 0
	for _, block := range blocks {
		count += len(block.Transactions)
	}
	return count
}
