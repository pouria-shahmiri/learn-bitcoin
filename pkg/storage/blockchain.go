package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// BlockchainStorage handles block storage and retrieval
type BlockchainStorage struct {
	db         *Database
	chainState *ChainState
}

// NewBlockchainStorage creates blockchain storage manager
func NewBlockchainStorage(dbPath string) (*BlockchainStorage, error) {
	db, err := OpenDatabase(dbPath)
	if err != nil {
		return nil, err
	}

	return &BlockchainStorage{
		db:         db,
		chainState: NewChainState(db),
	}, nil
}

// Close closes the database
func (bs *BlockchainStorage) Close() error {
	return bs.db.Close()
}

// SaveBlock stores a block with all indexes
func (bs *BlockchainStorage) SaveBlock(block *types.Block, height uint64) error {
	// Compute block hash
	blockHash, err := serialization.HashBlockHeader(&block.Header)
	if err != nil {
		return fmt.Errorf("failed to hash block: %w", err)
	}

	// Serialize block
	serializedBlock, err := serializeBlock(block)
	if err != nil {
		return fmt.Errorf("failed to serialize block: %w", err)
	}

	// Create atomic batch
	batch := bs.db.NewBatch()

	// 1. Store block data
	blockKey := BlockKey(blockHash)
	batch.Put(blockKey, serializedBlock)

	// 2. Store height index
	heightKey := HeightKey(height)
	batch.Put(heightKey, blockHash[:])

	// 3. Store transaction indexes
	for txIndex, tx := range block.Transactions {
		txHash, err := serialization.HashTransaction(&tx)
		if err != nil {
			return fmt.Errorf("failed to hash transaction: %w", err)
		}

		txKey := TxKey(txHash)
		txLocation := serializeTxLocation(blockHash, uint32(txIndex))
		batch.Put(txKey, txLocation)
	}

	// 4. Update chain state (if this is new tip)
	batch.Put(ChainStateKey(KeyBestBlockHash), blockHash[:])
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, height)
	batch.Put(ChainStateKey(KeyBestBlockHeight), heightBytes)

	// Commit everything atomically
	return batch.Write()
}

// GetBlock retrieves block by hash
func (bs *BlockchainStorage) GetBlock(hash types.Hash) (*types.Block, error) {
	key := BlockKey(hash)
	value, err := bs.db.Get(key)
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, fmt.Errorf("block not found: %s", hash)
	}

	return deserializeBlock(value)
}

// GetBlockByHeight retrieves block by height
func (bs *BlockchainStorage) GetBlockByHeight(height uint64) (*types.Block, error) {
	// First get hash from height index
	heightKey := HeightKey(height)
	hashBytes, err := bs.db.Get(heightKey)
	if err != nil {
		return nil, err
	}

	if hashBytes == nil {
		return nil, fmt.Errorf("no block at height %d", height)
	}

	var hash types.Hash
	copy(hash[:], hashBytes)

	// Then get block by hash
	return bs.GetBlock(hash)
}

// HasBlock checks if block exists
func (bs *BlockchainStorage) HasBlock(hash types.Hash) (bool, error) {
	key := BlockKey(hash)
	return bs.db.Has(key)
}

// GetBestBlock returns current chain tip
func (bs *BlockchainStorage) GetBestBlock() (*types.Block, uint64, error) {
	hash, err := bs.chainState.GetBestBlockHash()
	if err != nil {
		return nil, 0, err
	}

	if hash.IsZero() {
		return nil, 0, fmt.Errorf("blockchain is empty")
	}

	height, err := bs.chainState.GetBestBlockHeight()
	if err != nil {
		return nil, 0, err
	}

	block, err := bs.GetBlock(hash)
	if err != nil {
		return nil, 0, err
	}

	return block, height, nil
}

// GetTransactionLocation finds which block contains a transaction
func (bs *BlockchainStorage) GetTransactionLocation(txHash types.Hash) (blockHash types.Hash, txIndex uint32, err error) {
	key := TxKey(txHash)
	value, err := bs.db.Get(key)
	if err != nil {
		return types.Hash{}, 0, err
	}

	if value == nil {
		return types.Hash{}, 0, fmt.Errorf("transaction not found: %s", txHash)
	}

	return deserializeTxLocation(value)
}

// GetBlockCount returns total number of blocks
func (bs *BlockchainStorage) GetBlockCount() (uint64, error) {
	height, err := bs.chainState.GetBestBlockHeight()
	if err != nil {
		return 0, err
	}
	return height + 1, nil // Height is 0-indexed
}

// GetBestBlockHash returns the hash of the tip block
func (bs *BlockchainStorage) GetBestBlockHash() (types.Hash, error) {
	return bs.chainState.GetBestBlockHash()
}

// GetBestBlockHeight returns the height of the tip block
func (bs *BlockchainStorage) GetBestBlockHeight() (uint64, error) {
	return bs.chainState.GetBestBlockHeight()
}

// IsEmpty checks if the blockchain is empty
func (bs *BlockchainStorage) IsEmpty() (bool, error) {
	hash, err := bs.chainState.GetBestBlockHash()
	if err != nil {
		return false, err
	}
	return hash.IsZero(), nil
}

// Helper: Serialize block
func serializeBlock(block *types.Block) ([]byte, error) {
	var buf bytes.Buffer

	// Serialize header
	headerBytes, err := serialization.SerializeBlockHeader(&block.Header)
	if err != nil {
		return nil, err
	}
	buf.Write(headerBytes)

	// Transaction count
	if err := serialization.WriteVarInt(&buf, uint64(len(block.Transactions))); err != nil {
		return nil, err
	}

	// Serialize each transaction
	for _, tx := range block.Transactions {
		txBytes, err := serialization.SerializeTransaction(&tx)
		if err != nil {
			return nil, err
		}
		buf.Write(txBytes)
	}

	return buf.Bytes(), nil
}

// Helper: Deserialize block
func deserializeBlock(data []byte) (*types.Block, error) {
	r := bytes.NewReader(data)

	header, err := serialization.DeserializeBlockHeader(r)
	if err != nil {
		return nil, err
	}

	txCount, err := serialization.ReadVarInt(r)
	if err != nil {
		return nil, err
	}

	txs := make([]types.Transaction, txCount)
	for i := uint64(0); i < txCount; i++ {
		tx, err := serialization.DeserializeTransaction(r)
		if err != nil {
			return nil, err
		}
		txs[i] = *tx
	}

	return &types.Block{
		Header:       *header,
		Transactions: txs,
	}, nil
}

// Helper: Serialize transaction location
func serializeTxLocation(blockHash types.Hash, txIndex uint32) []byte {
	result := make([]byte, 32+4)
	copy(result[0:32], blockHash[:])
	binary.BigEndian.PutUint32(result[32:], txIndex)
	return result
}

// Helper: Deserialize transaction location
func deserializeTxLocation(data []byte) (types.Hash, uint32, error) {
	if len(data) != 36 {
		return types.Hash{}, 0, fmt.Errorf("invalid tx location length: %d", len(data))
	}

	var hash types.Hash
	copy(hash[:], data[0:32])
	index := binary.BigEndian.Uint32(data[32:])

	return hash, index, nil
}

// GetBlockHash returns the hash of a block
func (bs *BlockchainStorage) GetBlockHash(block *types.Block) (types.Hash, error) {
	return serialization.HashBlockHeader(&block.Header)
}
