package storage

import (
	"encoding/binary"
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// Key prefixes for different data types
const (
	// Block data: 'b' + block_hash -> serialized block
	PrefixBlock = 'b'

	// Block index: 'h' + height -> block_hash
	PrefixHeight = 'h'

	// Transaction index: 't' + tx_hash -> block_hash + tx_index
	PrefixTx = 't'

	// Chain state: 'c' + key -> value
	PrefixChainState = 'c'

	// Block height index: 'i' + block_hash -> height
	PrefixBlockHeight = 'i'
)

// Chain state keys
const (
	KeyBestBlockHash   = "bestblock"  // Current chain tip hash
	KeyBestBlockHeight = "bestheight" // Current chain height
)

// BlockKey creates key for storing block data
// Format: 'b' + block_hash
func BlockKey(hash types.Hash) []byte {
	key := make([]byte, 1+32)
	key[0] = PrefixBlock
	copy(key[1:], hash[:])
	return key
}

// HeightKey creates key for height index
// Format: 'h' + height (8 bytes, big-endian)
func HeightKey(height uint64) []byte {
	key := make([]byte, 1+8)
	key[0] = PrefixHeight
	binary.BigEndian.PutUint64(key[1:], height)
	return key
}

// TxKey creates key for transaction index
// Format: 't' + tx_hash
func TxKey(hash types.Hash) []byte {
	key := make([]byte, 1+32)
	key[0] = PrefixTx
	copy(key[1:], hash[:])
	return key
}

// BlockHeightKey creates key for block hash to height index
// Format: 'i' + block_hash
func BlockHeightKey(hash types.Hash) []byte {
	key := make([]byte, 1+32)
	key[0] = PrefixBlockHeight
	copy(key[1:], hash[:])
	return key
}

// ChainStateKey creates key for chain state
// Format: 'c' + string_key
func ChainStateKey(key string) []byte {
	result := make([]byte, 1+len(key))
	result[0] = PrefixChainState
	copy(result[1:], []byte(key))
	return result
}

// ParseBlockKey extracts hash from block key
func ParseBlockKey(key []byte) (types.Hash, error) {
	if len(key) != 33 || key[0] != PrefixBlock {
		return types.Hash{}, fmt.Errorf("invalid block key")
	}
	var hash types.Hash
	copy(hash[:], key[1:])
	return hash, nil
}

// ParseHeightKey extracts height from height key
func ParseHeightKey(key []byte) (uint64, error) {
	if len(key) != 9 || key[0] != PrefixHeight {
		return 0, fmt.Errorf("invalid height key")
	}
	return binary.BigEndian.Uint64(key[1:]), nil
}

/*
```
**Key Design Explained:**
```
Database Layout:
  'b' + <32-byte hash> → <serialized block>     (Block data)
  'h' + <8-byte height> → <32-byte hash>        (Height index)
  't' + <32-byte txid> → <32-byte block hash>   (Transaction lookup)
  'c' + "bestblock" → <32-byte hash>            (Chain tip)
  'c' + "bestheight" → <8-byte height>          (Chain height)
*/
