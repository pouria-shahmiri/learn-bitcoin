package sync

import (
	"fmt"
	"sync"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/network/protocol"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// MessageSender defines interface for sending messages
type MessageSender interface {
	SendMessage(msg *protocol.Message)
	Address() string
}

// SyncManager handles block synchronization
type SyncManager struct {
	chain *storage.BlockchainStorage
	mutex sync.Mutex

	// Keep track of requested blocks to avoid duplicate requests
	requestedBlocks map[types.Hash]string // hash -> peer address
}

// NewSyncManager creates a new sync manager
func NewSyncManager(chain *storage.BlockchainStorage) *SyncManager {
	return &SyncManager{
		chain:           chain,
		requestedBlocks: make(map[types.Hash]string),
	}
}

// HandleInv handles inventory announcements
func (sm *SyncManager) HandleInv(msg *protocol.InvMessage, peer MessageSender) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	getData := protocol.NewGetDataMessage()

	for _, vect := range msg.Inventory {
		if vect.Type == protocol.InvTypeBlock {
			// Check if we already have this block
			exists, err := sm.chain.HasBlock(vect.Hash)
			if err != nil {
				return err
			}

			if !exists {
				// Check if already requested
				if _, requested := sm.requestedBlocks[vect.Hash]; !requested {
					// Request it
					getData.AddInvVect(vect)
					sm.requestedBlocks[vect.Hash] = peer.Address()
				}
			}
		} else if vect.Type == protocol.InvTypeTx {
			// For now, request all transactions we don't know about
			// In a real node, we'd check mempool
			getData.AddInvVect(vect)
		}
	}

	// Send getdata if we have items to request
	if len(getData.Inventory) > 0 {
		peer.SendMessage(protocol.NewMessage(
			protocol.MagicMainnet,
			protocol.CmdGetData,
			mustSerialize(getData),
		))
	}

	return nil
}

// HandleBlock handles a received block
func (sm *SyncManager) HandleBlock(block *types.Block, peer MessageSender) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Calculate hash
	hash, err := sm.chain.GetBlockHash(block)
	if err != nil {
		return err
	}

	// Remove from requested list
	delete(sm.requestedBlocks, hash)

	// Check if we already have it
	exists, err := sm.chain.HasBlock(hash)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// Save block
	// Note: In a real node, we would validate the block first!
	// We'll assume the chain.SaveBlock handles basic validation or we trust for now
	// But we need the height. For this phase, we might need to assume it's next block
	// or look up parent.

	prevHash := block.Header.PrevBlockHash

	// Check if parent exists
	exists, err = sm.chain.HasBlock(prevHash)
	if err != nil {
		return err
	}

	var height uint64
	if !exists {
		// Parent not found. This is an orphan block.
		if prevHash.IsZero() {
			height = 0
		} else {
			return fmt.Errorf("parent block %s not found", prevHash)
		}
	} else {
		// Get parent height
		prevHeight, err := sm.chain.GetBlockHeight(prevHash)
		if err != nil {
			return fmt.Errorf("failed to get parent height: %w", err)
		}
		height = prevHeight + 1
	}

	if err := sm.chain.SaveBlock(block, height); err != nil {
		return fmt.Errorf("failed to save block: %w", err)
	}

	fmt.Printf("Synced block %s at height %d from %s\n", hash, height, peer.Address())

	return nil
}

// StartSync initiates sync with a peer
func (sm *SyncManager) StartSync(peer MessageSender) error {
	// Send getblocks to find common history
	// Start from our best block
	locator := sm.getBlockLocator()

	msg := protocol.NewGetBlocksMessage(locator, types.Hash{}) // HashStop = 0

	peer.SendMessage(protocol.NewMessage(
		protocol.MagicMainnet,
		protocol.CmdGetBlocks,
		mustSerialize(msg),
	))

	return nil
}

// getBlockLocator returns block hashes from newest to oldest
// Used to find the divergence point between chains
func (sm *SyncManager) getBlockLocator() []types.Hash {
	// Simplified: just return the tip
	hash, err := sm.chain.GetBestBlockHash()
	if err != nil {
		return []types.Hash{}
	}
	return []types.Hash{hash}
}

func mustSerialize(msg interface{ Serialize() ([]byte, error) }) []byte {
	b, err := msg.Serialize()
	if err != nil {
		panic(err)
	}
	return b
}
