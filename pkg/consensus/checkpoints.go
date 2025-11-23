package consensus

import (
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// Checkpoint represents a known-good block hash at a specific height
type Checkpoint struct {
	Height uint64
	Hash   types.Hash
}

// CheckpointVerifier verifies blocks against known checkpoints
type CheckpointVerifier struct {
	checkpoints []Checkpoint
	enabled     bool
}

// NewCheckpointVerifier creates a new checkpoint verifier
func NewCheckpointVerifier(enabled bool) *CheckpointVerifier {
	return &CheckpointVerifier{
		checkpoints: getMainnetCheckpoints(),
		enabled:     enabled,
	}
}

// getMainnetCheckpoints returns hardcoded mainnet checkpoints
func getMainnetCheckpoints() []Checkpoint {
	// These are example checkpoints - in production, use real Bitcoin checkpoints
	return []Checkpoint{
		{Height: 11111, Hash: hashFromString("0000000069e244f73d78e8fd29ba2fd2ed618bd6fa2ee92559f542fdb26e7c1d")},
		{Height: 33333, Hash: hashFromString("000000002dd5588a74784eaa7ab0507a18ad16a236e7b1ce69f00d7ddfb5d0a6")},
		{Height: 74000, Hash: hashFromString("0000000000573993a3c9e41ce34471c079dcf5f52a0e824a81e7f953b8661a20")},
		{Height: 105000, Hash: hashFromString("00000000000291ce28027faea320c8d2b054b2e0fe44a773f3eefb151d6bdc97")},
		{Height: 134444, Hash: hashFromString("00000000000005b12ffd4cd315cd34ffd4a594f430ac814c91184a0d42d2b0fe")},
		{Height: 168000, Hash: hashFromString("000000000000099e61ea72015e79632f216fe6cb33d7899acb35b75c8303b763")},
		{Height: 193000, Hash: hashFromString("000000000000059f452a5f7340de6682a977387c17010ff6e6c3bd83ca8b1317")},
		{Height: 210000, Hash: hashFromString("000000000000048b95347e83192f69cf0366076336c639f9b7228e9ba171342e")},
		{Height: 216116, Hash: hashFromString("00000000000001b4f4b433e81ee46494af945cf96014816a4e2370f11b23df4e")},
		{Height: 225430, Hash: hashFromString("00000000000001c108384350f74090433e7fcf79a606b8e797f065b130575932")},
	}
}

// getTestnetCheckpoints returns hardcoded testnet checkpoints
func getTestnetCheckpoints() []Checkpoint {
	return []Checkpoint{
		{Height: 546, Hash: hashFromString("000000002a936ca763904c3c35fce2f3556c559c0214345d31b1bcebf76acb70")},
	}
}

// AddCheckpoint adds a custom checkpoint
func (cv *CheckpointVerifier) AddCheckpoint(height uint64, hash types.Hash) {
	cv.checkpoints = append(cv.checkpoints, Checkpoint{
		Height: height,
		Hash:   hash,
	})
}

// VerifyCheckpoint verifies a block against checkpoints
func (cv *CheckpointVerifier) VerifyCheckpoint(height uint64, hash types.Hash) error {
	if !cv.enabled {
		return nil // Checkpoints disabled
	}

	// Find checkpoint at this height
	for _, cp := range cv.checkpoints {
		if cp.Height == height {
			if hash != cp.Hash {
				return fmt.Errorf("checkpoint verification failed at height %d: expected %s, got %s",
					height, cp.Hash, hash)
			}
			return nil
		}
	}

	// No checkpoint at this height
	return nil
}

// GetLastCheckpoint returns the last checkpoint before or at given height
func (cv *CheckpointVerifier) GetLastCheckpoint(height uint64) *Checkpoint {
	var lastCP *Checkpoint

	for i := range cv.checkpoints {
		if cv.checkpoints[i].Height <= height {
			if lastCP == nil || cv.checkpoints[i].Height > lastCP.Height {
				lastCP = &cv.checkpoints[i]
			}
		}
	}

	return lastCP
}

// IsCheckpoint checks if a height has a checkpoint
func (cv *CheckpointVerifier) IsCheckpoint(height uint64) bool {
	for _, cp := range cv.checkpoints {
		if cp.Height == height {
			return true
		}
	}
	return false
}

// ValidateAgainstCheckpoints validates that a chain doesn't conflict with checkpoints
func (cv *CheckpointVerifier) ValidateAgainstCheckpoints(blocks map[uint64]types.Hash) error {
	if !cv.enabled {
		return nil
	}

	for _, cp := range cv.checkpoints {
		if hash, exists := blocks[cp.Height]; exists {
			if hash != cp.Hash {
				return fmt.Errorf("chain conflicts with checkpoint at height %d", cp.Height)
			}
		}
	}

	return nil
}

// hashFromString converts a hex string to a Hash (helper for checkpoints)
func hashFromString(s string) types.Hash {
	var hash types.Hash
	// In production, properly decode the hex string
	// For now, return empty hash as placeholder
	return hash
}
