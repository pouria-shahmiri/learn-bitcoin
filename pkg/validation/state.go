package validation

import (
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/utxo"
)

// ChainState manages the blockchain state
type ChainState struct {
	blockchain *storage.BlockchainStorage
	utxoSet    *utxo.UTXOSet
	validator  *BlockValidator
}

// NewChainState creates a new chain state
func NewChainState(dbPath string) (*ChainState, error) {
	// Open blockchain storage
	blockchain, err := storage.NewBlockchainStorage(dbPath)
	if err != nil {
		return nil, err
	}

	// Create UTXO set
	utxoSet := utxo.NewUTXOSet()

	// Create validator
	validator := NewBlockValidator(utxoSet)

	return &ChainState{
		blockchain: blockchain,
		utxoSet:    utxoSet,
		validator:  validator,
	}, nil
}

// Close closes the chain state
func (cs *ChainState) Close() error {
	return cs.blockchain.Close()
}

// AddBlock validates and adds a block to the chain
func (cs *ChainState) AddBlock(block *types.Block) error {
	// Get current best block
	_, currentHeight, err := cs.blockchain.GetBestBlock()
	if err != nil {
		// If no blocks exist, this should be genesis
		if currentHeight == 0 {
			return cs.addGenesisBlock(block)
		}
		return err
	}

	// Get previous block hash
	prevBlock, _, err := cs.blockchain.GetBestBlock()
	if err != nil {
		return err
	}
	prevHash, err := cs.blockchain.GetBlockHash(prevBlock)
	if err != nil {
		return err
	}

	newHeight := currentHeight + 1

	// Validate block
	if err := cs.validator.ValidateBlock(block, newHeight, prevHash); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// Apply block to UTXO set
	if err := cs.validator.ApplyBlock(block, newHeight); err != nil {
		return fmt.Errorf("failed to apply block: %w", err)
	}

	// Save block to storage
	if err := cs.blockchain.SaveBlock(block, newHeight); err != nil {
		// Revert UTXO changes on failure
		cs.validator.RevertBlock(block)
		return fmt.Errorf("failed to save block: %w", err)
	}

	return nil
}

// addGenesisBlock adds the genesis block
func (cs *ChainState) addGenesisBlock(block *types.Block) error {
	// Genesis block has no previous block
	// Minimal validation

	if len(block.Transactions) == 0 {
		return fmt.Errorf("genesis block has no transactions")
	}

	// Apply genesis to UTXO set
	if err := cs.validator.ApplyBlock(block, 0); err != nil {
		return err
	}

	// Save to storage
	if err := cs.blockchain.SaveBlock(block, 0); err != nil {
		return err
	}

	return nil
}

// GetBestBlock returns the current best block
func (cs *ChainState) GetBestBlock() (*types.Block, uint64, error) {
	return cs.blockchain.GetBestBlock()
}

// GetBlock returns a block by hash
func (cs *ChainState) GetBlock(hash types.Hash) (*types.Block, error) {
	return cs.blockchain.GetBlock(hash)
}

// GetBlockByHeight returns a block by height
func (cs *ChainState) GetBlockByHeight(height uint64) (*types.Block, error) {
	return cs.blockchain.GetBlockByHeight(height)
}

// GetUTXOSet returns the UTXO set
func (cs *ChainState) GetUTXOSet() *utxo.UTXOSet {
	return cs.utxoSet
}

// RebuildUTXOSet rebuilds the UTXO set from the blockchain
func (cs *ChainState) RebuildUTXOSet() error {
	// Clear current UTXO set
	cs.utxoSet.Clear()

	// Get blockchain height
	_, height, err := cs.blockchain.GetBestBlock()
	if err != nil {
		return err
	}

	// Replay all blocks
	for h := uint64(0); h <= height; h++ {
		block, err := cs.blockchain.GetBlockByHeight(h)
		if err != nil {
			return fmt.Errorf("failed to get block at height %d: %w", h, err)
		}

		if err := cs.validator.ApplyBlock(block, h); err != nil {
			return fmt.Errorf("failed to apply block at height %d: %w", h, err)
		}
	}

	return nil
}

// ValidateChain validates the entire blockchain
func (cs *ChainState) ValidateChain() error {
	_, height, err := cs.blockchain.GetBestBlock()
	if err != nil {
		return err
	}

	// Create temporary UTXO set for validation
	tempUTXO := utxo.NewUTXOSet()
	tempValidator := NewBlockValidator(tempUTXO)

	// Validate each block
	for h := uint64(0); h <= height; h++ {
		block, err := cs.blockchain.GetBlockByHeight(h)
		if err != nil {
			return fmt.Errorf("failed to get block at height %d: %w", h, err)
		}

		// Get previous block hash
		var prevHash types.Hash
		if h > 0 {
			prevBlock, err := cs.blockchain.GetBlockByHeight(h - 1)
			if err != nil {
				return err
			}
			prevHash, _ = cs.blockchain.GetBlockHash(prevBlock)
		}

		// Validate block
		if err := tempValidator.ValidateBlock(block, h, prevHash); err != nil {
			return fmt.Errorf("block at height %d invalid: %w", h, err)
		}

		// Apply to temp UTXO
		if err := tempValidator.ApplyBlock(block, h); err != nil {
			return err
		}
	}

	return nil
}

// GetBlockCount returns the number of blocks in the chain
func (cs *ChainState) GetBlockCount() (uint64, error) {
	return cs.blockchain.GetBlockCount()
}
