package reorg

import (
	"fmt"
	"math/big"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// ReorgDetector detects when a chain reorganization is needed
type ReorgDetector struct {
	blockchain *storage.BlockchainStorage
}

// NewReorgDetector creates a new reorganization detector
func NewReorgDetector(blockchain *storage.BlockchainStorage) *ReorgDetector {
	return &ReorgDetector{
		blockchain: blockchain,
	}
}

// ChainInfo contains information about a chain
type ChainInfo struct {
	Tip        types.Hash
	Height     uint64
	TotalWork  *big.Int
	Blocks     []*types.Block
	ForkHeight uint64
}

// DetectReorg checks if a new chain has more work than the current chain
func (rd *ReorgDetector) DetectReorg(newBlocks []*types.Block) (bool, *ChainInfo, error) {
	if len(newBlocks) == 0 {
		return false, nil, fmt.Errorf("no blocks provided")
	}

	// Get current chain info
	currentTip, currentHeight, err := rd.blockchain.GetBestBlock()
	if err != nil {
		return false, nil, fmt.Errorf("failed to get current tip: %w", err)
	}

	currentWork, err := rd.calculateChainWork(currentHeight)
	if err != nil {
		return false, nil, fmt.Errorf("failed to calculate current work: %w", err)
	}

	// Find fork point
	forkHeight, err := rd.findForkPoint(newBlocks)
	if err != nil {
		return false, nil, fmt.Errorf("failed to find fork point: %w", err)
	}

	// Calculate new chain work
	newHeight := forkHeight + uint64(len(newBlocks))
	newWork, err := rd.calculateNewChainWork(forkHeight, newBlocks)
	if err != nil {
		return false, nil, fmt.Errorf("failed to calculate new work: %w", err)
	}

	// Check if new chain has more work
	needsReorg := newWork.Cmp(currentWork) > 0

	if needsReorg {
		currentTipHash, _ := serialization.HashBlockHeader(&currentTip.Header)
		newTipHash, _ := serialization.HashBlockHeader(&newBlocks[len(newBlocks)-1].Header)

		chainInfo := &ChainInfo{
			Tip:        newTipHash,
			Height:     newHeight,
			TotalWork:  newWork,
			Blocks:     newBlocks,
			ForkHeight: forkHeight,
		}

		fmt.Printf("Reorg detected: current=%s (work=%s) new=%s (work=%s)\n",
			currentTipHash, currentWork, newTipHash, newWork)

		return true, chainInfo, nil
	}

	return false, nil, nil
}

// findForkPoint finds the height where the new chain diverges from current chain
func (rd *ReorgDetector) findForkPoint(newBlocks []*types.Block) (uint64, error) {
	if len(newBlocks) == 0 {
		return 0, fmt.Errorf("no blocks provided")
	}

	// Start with the first block's previous hash
	prevHash := newBlocks[0].Header.PrevBlockHash

	// Find this block in our chain
	height, err := rd.blockchain.GetBlockHeight(prevHash)
	if err != nil {
		// If we can't find it, check if it's genesis
		isEmpty, _ := rd.blockchain.IsEmpty()
		if isEmpty {
			return 0, nil
		}
		return 0, fmt.Errorf("fork point not found: %w", err)
	}

	return height, nil
}

// calculateChainWork calculates total work for current chain up to given height
func (rd *ReorgDetector) calculateChainWork(height uint64) (*big.Int, error) {
	totalWork := big.NewInt(0)

	for h := uint64(0); h <= height; h++ {
		block, err := rd.blockchain.GetBlockByHeight(h)
		if err != nil {
			return nil, fmt.Errorf("failed to get block at height %d: %w", h, err)
		}

		work := calculateBlockWork(block.Header.Bits)
		totalWork.Add(totalWork, work)
	}

	return totalWork, nil
}

// calculateNewChainWork calculates work for a new chain from fork point
func (rd *ReorgDetector) calculateNewChainWork(forkHeight uint64, newBlocks []*types.Block) (*big.Int, error) {
	// Start with work up to fork point
	totalWork, err := rd.calculateChainWork(forkHeight)
	if err != nil {
		return nil, err
	}

	// Add work from new blocks
	for _, block := range newBlocks {
		work := calculateBlockWork(block.Header.Bits)
		totalWork.Add(totalWork, work)
	}

	return totalWork, nil
}

// calculateBlockWork calculates work for a single block based on difficulty bits
func calculateBlockWork(bits uint32) *big.Int {
	// Work = 2^256 / (target + 1)
	// For simplicity, we'll use: work = 2^(256 - difficulty_bits)

	// Extract exponent and coefficient from compact bits
	exponent := bits >> 24
	coefficient := bits & 0x00ffffff

	// Calculate target
	target := big.NewInt(int64(coefficient))
	target.Lsh(target, uint(8*(exponent-3)))

	// Calculate work = 2^256 / (target + 1)
	maxTarget := new(big.Int).Lsh(big.NewInt(1), 256)
	target.Add(target, big.NewInt(1))
	work := new(big.Int).Div(maxTarget, target)

	return work
}

// GetForkBlocks returns blocks that need to be disconnected during reorg
func (rd *ReorgDetector) GetForkBlocks(forkHeight uint64) ([]*types.Block, error) {
	currentHeight, err := rd.blockchain.GetBestBlockHeight()
	if err != nil {
		return nil, err
	}

	var blocks []*types.Block
	for h := currentHeight; h > forkHeight; h-- {
		block, err := rd.blockchain.GetBlockByHeight(h)
		if err != nil {
			return nil, fmt.Errorf("failed to get block at height %d: %w", h, err)
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}
