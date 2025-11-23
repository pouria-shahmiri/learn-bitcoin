package consensus

import (
	"fmt"
	"time"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// ConsensusRules defines Bitcoin consensus rules
type ConsensusRules struct {
	// Network parameters
	MaxBlockSize           uint32
	MaxBlockWeight         uint32
	CoinbaseMaturity       uint32
	SubsidyHalvingInterval uint32

	// Time-based rules
	MaxFutureBlockTime time.Duration
	MedianTimeSpan     int

	// BIP activation heights
	BIP16Height  uint64
	BIP34Height  uint64
	BIP65Height  uint64
	BIP66Height  uint64
	SegWitHeight uint64
}

// NewMainnetRules returns consensus rules for mainnet
func NewMainnetRules() *ConsensusRules {
	return &ConsensusRules{
		MaxBlockSize:           1000000, // 1 MB
		MaxBlockWeight:         4000000, // 4 MB (SegWit)
		CoinbaseMaturity:       100,
		SubsidyHalvingInterval: 210000,
		MaxFutureBlockTime:     2 * time.Hour,
		MedianTimeSpan:         11,
		BIP16Height:            173805,
		BIP34Height:            227931,
		BIP65Height:            388381,
		BIP66Height:            363725,
		SegWitHeight:           481824,
	}
}

// NewTestnetRules returns consensus rules for testnet
func NewTestnetRules() *ConsensusRules {
	return &ConsensusRules{
		MaxBlockSize:           1000000,
		MaxBlockWeight:         4000000,
		CoinbaseMaturity:       100,
		SubsidyHalvingInterval: 210000,
		MaxFutureBlockTime:     2 * time.Hour,
		MedianTimeSpan:         11,
		BIP16Height:            0,
		BIP34Height:            0,
		BIP65Height:            0,
		BIP66Height:            0,
		SegWitHeight:           0,
	}
}

// NewRegtestRules returns consensus rules for regtest
func NewRegtestRules() *ConsensusRules {
	return &ConsensusRules{
		MaxBlockSize:           1000000,
		MaxBlockWeight:         4000000,
		CoinbaseMaturity:       100,
		SubsidyHalvingInterval: 150,
		MaxFutureBlockTime:     2 * time.Hour,
		MedianTimeSpan:         11,
		BIP16Height:            0,
		BIP34Height:            0,
		BIP65Height:            0,
		BIP66Height:            0,
		SegWitHeight:           0,
	}
}

// ValidateBlockTime validates block timestamp against consensus rules
func (cr *ConsensusRules) ValidateBlockTime(blockTime uint32, medianTimePast uint32, currentTime time.Time) error {
	// Block time must be greater than median time of past blocks
	if blockTime <= medianTimePast {
		return fmt.Errorf("block time %d is not greater than median time past %d", blockTime, medianTimePast)
	}

	// Block time must not be too far in the future
	maxTime := currentTime.Add(cr.MaxFutureBlockTime)
	if time.Unix(int64(blockTime), 0).After(maxTime) {
		return fmt.Errorf("block time %d is too far in the future (max: %v)", blockTime, maxTime)
	}

	return nil
}

// ValidateBlockSize validates block size against consensus rules
func (cr *ConsensusRules) ValidateBlockSize(blockSize uint32, height uint64) error {
	// Check if SegWit is active
	if height >= cr.SegWitHeight {
		// Use weight-based validation
		// For non-SegWit blocks, weight = size * 4
		weight := blockSize * 4
		if weight > cr.MaxBlockWeight {
			return fmt.Errorf("block weight %d exceeds maximum %d", weight, cr.MaxBlockWeight)
		}
	} else {
		// Use size-based validation
		if blockSize > cr.MaxBlockSize {
			return fmt.Errorf("block size %d exceeds maximum %d", blockSize, cr.MaxBlockSize)
		}
	}

	return nil
}

// IsBIP34Active checks if BIP34 is active at given height
func (cr *ConsensusRules) IsBIP34Active(height uint64) bool {
	return height >= cr.BIP34Height
}

// IsBIP65Active checks if BIP65 (CHECKLOCKTIMEVERIFY) is active
func (cr *ConsensusRules) IsBIP65Active(height uint64) bool {
	return height >= cr.BIP65Height
}

// IsBIP66Active checks if BIP66 (strict DER signatures) is active
func (cr *ConsensusRules) IsBIP66Active(height uint64) bool {
	return height >= cr.BIP66Height
}

// IsSegWitActive checks if SegWit is active
func (cr *ConsensusRules) IsSegWitActive(height uint64) bool {
	return height >= cr.SegWitHeight
}

// GetBlockSubsidy calculates block subsidy at given height
func (cr *ConsensusRules) GetBlockSubsidy(height uint64) uint64 {
	halvings := height / uint64(cr.SubsidyHalvingInterval)

	// Subsidy is cut in half every halving interval
	if halvings >= 64 {
		return 0
	}

	subsidy := uint64(50 * 100000000) // 50 BTC in satoshis
	subsidy >>= halvings

	return subsidy
}

// ValidateCoinbaseMaturity checks if coinbase outputs can be spent
func (cr *ConsensusRules) ValidateCoinbaseMaturity(coinbaseHeight, spendHeight uint64) error {
	if spendHeight < coinbaseHeight+uint64(cr.CoinbaseMaturity) {
		return fmt.Errorf("coinbase output not mature: height %d, spend height %d, maturity %d",
			coinbaseHeight, spendHeight, cr.CoinbaseMaturity)
	}
	return nil
}

// ValidateBIP34 validates BIP34 block height in coinbase
func (cr *ConsensusRules) ValidateBIP34(coinbase *types.Transaction, height uint64) error {
	if !cr.IsBIP34Active(height) {
		return nil // BIP34 not active yet
	}

	if len(coinbase.Inputs) == 0 {
		return fmt.Errorf("coinbase has no inputs")
	}

	// First input's script sig should start with block height
	scriptSig := coinbase.Inputs[0].SignatureScript
	if len(scriptSig) == 0 {
		return fmt.Errorf("coinbase script sig is empty")
	}

	// For simplicity, we'll just check that script sig is not empty
	// Full BIP34 validation would decode the height from script sig
	// and verify it matches the block height

	return nil
}
