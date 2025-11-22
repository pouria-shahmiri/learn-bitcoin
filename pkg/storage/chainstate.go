package storage

import (
	"encoding/binary"
	"fmt"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// ChainState tracks the current state of the blockchain
type ChainState struct {
	db *Database
}

// NewChainState creates chain state manager
func NewChainState(db *Database) *ChainState {
	return &ChainState{db: db}
}

// GetBestBlockHash returns current chain tip hash
func (cs *ChainState) GetBestBlockHash() (types.Hash, error) {
	key := ChainStateKey(KeyBestBlockHash)
	value, err := cs.db.Get(key)
	if err != nil {
		return types.Hash{}, err
	}
	
	if value == nil {
		// No chain tip set (empty blockchain)
		return types.Hash{}, nil
	}
	
	if len(value) != 32 {
		return types.Hash{}, fmt.Errorf("invalid best block hash length: %d", len(value))
	}
	
	var hash types.Hash
	copy(hash[:], value)
	return hash, nil
}

// SetBestBlockHash updates chain tip
func (cs *ChainState) SetBestBlockHash(hash types.Hash) error {
	key := ChainStateKey(KeyBestBlockHash)
	return cs.db.Put(key, hash[:])
}

// GetBestBlockHeight returns current chain height
func (cs *ChainState) GetBestBlockHeight() (uint64, error) {
	key := ChainStateKey(KeyBestBlockHeight)
	value, err := cs.db.Get(key)
	if err != nil {
		return 0, err
	}
	
	if value == nil {
		// No height set (empty blockchain)
		return 0, nil
	}
	
	if len(value) != 8 {
		return 0, fmt.Errorf("invalid height length: %d", len(value))
	}
	
	return binary.BigEndian.Uint64(value), nil
}

// SetBestBlockHeight updates chain height
func (cs *ChainState) SetBestBlockHeight(height uint64) error {
	key := ChainStateKey(KeyBestBlockHeight)
	value := make([]byte, 8)
	binary.BigEndian.PutUint64(value, height)
	return cs.db.Put(key, value)
}

// UpdateBestBlock atomically updates both hash and height
func (cs *ChainState) UpdateBestBlock(hash types.Hash, height uint64) error {
	batch := cs.db.NewBatch()
	
	// Update hash
	hashKey := ChainStateKey(KeyBestBlockHash)
	batch.Put(hashKey, hash[:])
	
	// Update height
	heightKey := ChainStateKey(KeyBestBlockHeight)
	heightValue := make([]byte, 8)
	binary.BigEndian.PutUint64(heightValue, height)
	batch.Put(heightKey, heightValue)
	
	return batch.Write()
}

// IsEmpty checks if blockchain is empty
func (cs *ChainState) IsEmpty() (bool, error) {
	hash, err := cs.GetBestBlockHash()
	if err != nil {
		return false, err
	}
	return hash.IsZero(), nil
}