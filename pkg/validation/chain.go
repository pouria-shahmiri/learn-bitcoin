package validation

import (
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/utxo"
)

// ChainValidator validates blockchain
type ChainValidator struct {
	blockchain *storage.BlockchainStorage
	utxoSet    *utxo.UTXOSet
	validator  *BlockValidator
}

// NewChainValidator creates a new chain validator
func NewChainValidator(blockchain *storage.BlockchainStorage, utxoSet *utxo.UTXOSet) *ChainValidator {
	return &ChainValidator{
		blockchain: blockchain,
		utxoSet:    utxoSet,
		validator:  NewBlockValidator(utxoSet),
	}
}

// ValidateNewBlock validates a new block against the current chain
func (cv *ChainValidator) ValidateNewBlock(block *types.Block) error {
	// Get current chain tip
	bestBlock, bestHeight, err := cv.blockchain.GetBestBlock()
	if err != nil {
		return fmt.Errorf("failed to get best block: %w", err)
	}

	// Get previous block hash
	prevHash, err := serialization.HashBlockHeader(&bestBlock.Header)
	if err != nil {
		return err
	}

	// New block height
	newHeight := bestHeight + 1

	// Validate the block
	if err := cv.validator.ValidateBlock(block, newHeight, prevHash); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	return nil
}

// AcceptBlock validates and adds a block to the chain
func (cv *ChainValidator) AcceptBlock(block *types.Block) error {
	// Check if blockchain is empty (genesis block case)
	isEmpty, err := cv.blockchain.IsEmpty()
	if err != nil {
		return err
	}

	var newHeight uint64
	var prevHash types.Hash

	if isEmpty {
		// This is the genesis block
		newHeight = 0
		prevHash = types.Hash{} // No previous block
	} else {
		// Validate the block against current chain
		if err := cv.ValidateNewBlock(block); err != nil {
			return err
		}

		// Get new height
		_, bestHeight, err := cv.blockchain.GetBestBlock()
		if err != nil {
			return err
		}
		newHeight = bestHeight + 1
	}

	// Validate the block structure
	if err := cv.validator.ValidateBlock(block, newHeight, prevHash); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// Apply to UTXO set
	if err := cv.validator.ApplyBlock(block, newHeight); err != nil {
		return fmt.Errorf("failed to apply block: %w", err)
	}

	// Save to blockchain
	if err := cv.blockchain.SaveBlock(block, newHeight); err != nil {
		// Revert UTXO changes on failure
		cv.validator.RevertBlock(block)
		return fmt.Errorf("failed to save block: %w", err)
	}

	return nil
}

// GetBlockLocator returns block locator for sync
func (cv *ChainValidator) GetBlockLocator() ([]types.Hash, error) {
	var locator []types.Hash

	_, height, err := cv.blockchain.GetBestBlock()
	if err != nil {
		return nil, err
	}

	// Add blocks with exponentially increasing gaps
	step := uint64(1)
	for h := height; h > 0; h -= step {
		block, err := cv.blockchain.GetBlockByHeight(h)
		if err != nil {
			break
		}

		blockHash, err := serialization.HashBlockHeader(&block.Header)
		if err != nil {
			break
		}

		locator = append(locator, blockHash)

		// Exponential backoff
		if len(locator) > 10 {
			step *= 2
		}
	}

	// Always include genesis
	genesis, err := cv.blockchain.GetBlockByHeight(0)
	if err == nil {
		genesisHash, _ := serialization.HashBlockHeader(&genesis.Header)
		locator = append(locator, genesisHash)
	}

	return locator, nil
}

// FindFork finds the fork point between two chains
func (cv *ChainValidator) FindFork(otherHashes []types.Hash) (uint64, error) {
	for _, hash := range otherHashes {
		// Check if we have this block
		block, err := cv.blockchain.GetBlock(hash)
		if err == nil && block != nil {
			// Found common ancestor
			// Get its height
			_, height, err := cv.blockchain.GetBestBlock()
			if err != nil {
				return 0, err
			}
			return height, nil
		}
	}

	return 0, fmt.Errorf("no common ancestor found")
}

// Reorganize handles blockchain reorganization
func (cv *ChainValidator) Reorganize(newBlocks []*types.Block) error {
	// This is a simplified version
	// Real Bitcoin has more complex reorg logic

	// Find fork point
	// Disconnect blocks from old chain
	// Connect blocks from new chain

	// For now, we'll implement a basic version
	return fmt.Errorf("reorganization not fully implemented")
}

// GetChainWork calculates total chain work
func (cv *ChainValidator) GetChainWork() (uint64, error) {
	// Simplified: just return chain height
	// Real Bitcoin calculates actual proof-of-work

	_, height, err := cv.blockchain.GetBestBlock()
	if err != nil {
		return 0, err
	}

	return height, nil
}

// IsValidChain checks if the entire chain is valid
func (cv *ChainValidator) IsValidChain() error {
	_, height, err := cv.blockchain.GetBestBlock()
	if err != nil {
		return err
	}

	// Create temporary UTXO set for validation
	tempUTXO := utxo.NewUTXOSet()
	tempValidator := NewBlockValidator(tempUTXO)

	// Validate each block sequentially
	for h := uint64(0); h <= height; h++ {
		block, err := cv.blockchain.GetBlockByHeight(h)
		if err != nil {
			return fmt.Errorf("failed to get block at height %d: %w", h, err)
		}

		// Get previous block hash
		var prevHash types.Hash
		if h > 0 {
			prevBlock, err := cv.blockchain.GetBlockByHeight(h - 1)
			if err != nil {
				return err
			}
			prevHash, _ = serialization.HashBlockHeader(&prevBlock.Header)
		}

		// Validate
		if err := tempValidator.ValidateBlock(block, h, prevHash); err != nil {
			return fmt.Errorf("block at height %d invalid: %w", h, err)
		}

		// Apply to temp UTXO
		if err := tempValidator.ApplyBlock(block, h); err != nil {
			return fmt.Errorf("failed to apply block at height %d: %w", h, err)
		}
	}

	return nil
}

// GetBlockHeight returns the height of a block by hash
func (cv *ChainValidator) GetBlockHeight(hash types.Hash) (uint64, error) {
	// This requires a hash->height index in storage
	// For now, we'll search linearly (inefficient)

	_, maxHeight, err := cv.blockchain.GetBestBlock()
	if err != nil {
		return 0, err
	}

	for h := uint64(0); h <= maxHeight; h++ {
		block, err := cv.blockchain.GetBlockByHeight(h)
		if err != nil {
			continue
		}

		blockHash, err := serialization.HashBlockHeader(&block.Header)
		if err != nil {
			continue
		}

		if blockHash == hash {
			return h, nil
		}
	}

	return 0, fmt.Errorf("block not found")
}
