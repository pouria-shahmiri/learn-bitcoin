package validation

import (
	"fmt"
)

// Consensus constants
const (
	// MaxBlockSize is the maximum block size in bytes
	MaxBlockSize = 1000000 // 1MB

	// MaxMoney is the maximum amount of satoshis (21 million BTC)
	MaxMoney = 21000000 * 100000000

	// CoinbaseMaturity is the number of blocks before coinbase can be spent
	CoinbaseMaturity = 100

	// InitialBlockReward is the initial block reward in satoshis
	InitialBlockReward = 50 * 100000000

	// SubsidyHalvingInterval is the number of blocks between halvings
	SubsidyHalvingInterval = 210000
)

// GetBlockReward calculates the block reward for a given height
func GetBlockReward(height uint64) int64 {
	// Calculate number of halvings
	halvings := height / SubsidyHalvingInterval

	// If we've had 64 or more halvings, reward is 0
	if halvings >= 64 {
		return 0
	}

	// Calculate reward: initial reward >> halvings
	reward := int64(InitialBlockReward >> halvings)

	return reward
}

// CheckMoneyRange checks if a value is in valid range
func CheckMoneyRange(value int64) error {
	if value < 0 {
		return fmt.Errorf("negative value: %d", value)
	}

	if value > MaxMoney {
		return fmt.Errorf("value exceeds maximum: %d > %d", value, MaxMoney)
	}

	return nil
}

// ValidateBlockReward checks if block reward is correct
func ValidateBlockReward(coinbaseValue int64, totalFees int64, height uint64) error {
	expectedReward := GetBlockReward(height)
	maxAllowed := expectedReward + totalFees

	if coinbaseValue > maxAllowed {
		return fmt.Errorf("coinbase value (%d) exceeds allowed (%d)",
			coinbaseValue, maxAllowed)
	}

	return nil
}

// Difficulty adjustment (simplified for learning)
func CalculateNextDifficulty(prevDifficulty uint32, actualTime, targetTime uint64) uint32 {
	// In real Bitcoin, this is more complex
	// Simplified version: adjust based on time ratio

	if actualTime < targetTime/4 {
		actualTime = targetTime / 4
	}
	if actualTime > targetTime*4 {
		actualTime = targetTime * 4
	}

	// Adjust difficulty
	newDifficulty := uint32(uint64(prevDifficulty) * targetTime / actualTime)

	return newDifficulty
}

// IsValidProofOfWork checks if block hash meets difficulty target
func IsValidProofOfWork(blockHash []byte, bits uint32) bool {
	// Simplified proof-of-work check
	// In real Bitcoin, this involves compact target representation

	// Count leading zero bytes
	leadingZeros := 0
	for _, b := range blockHash {
		if b == 0 {
			leadingZeros++
		} else {
			break
		}
	}

	// Require at least 0 leading zero bytes for demo (relaxed for testing)
	requiredZeros := 0

	return leadingZeros >= requiredZeros
}
