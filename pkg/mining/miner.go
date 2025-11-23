package mining

import (
	"fmt"
	"time"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/validation"
)

// MiningStats tracks mining performance
type MiningStats struct {
	StartTime    time.Time
	Attempts     uint64
	HashRate     float64 // Hashes per second
	CurrentNonce uint32
	Difficulty   uint32
	TargetZeros  int
}

// Miner performs proof-of-work mining
type Miner struct {
	stats MiningStats
}

// NewMiner creates a new miner
func NewMiner() *Miner {
	return &Miner{
		stats: MiningStats{},
	}
}

// MineBlock mines a block by finding a valid nonce
func (m *Miner) MineBlock(template *BlockTemplate, targetZeros int) (*types.Block, error) {
	fmt.Printf("\n🔨 Starting mining...\n")
	fmt.Printf("   Target: %d leading zero bytes\n", targetZeros)
	fmt.Printf("   Difficulty: %d\n", template.Bits)

	// Initialize stats
	m.stats = MiningStats{
		StartTime:   time.Now(),
		Attempts:    0,
		Difficulty:  template.Bits,
		TargetZeros: targetZeros,
	}

	// Try different nonces
	nonce := uint32(0)

	for {
		// Build block with current nonce
		block, err := BuildBlock(template, nonce)
		if err != nil {
			return nil, err
		}

		// Hash the block header
		blockHash, err := serialization.HashBlockHeader(&block.Header)
		if err != nil {
			return nil, err
		}

		// Update stats
		m.stats.Attempts++
		m.stats.CurrentNonce = nonce

		// Check if hash meets difficulty
		if m.checkProofOfWork(blockHash[:], targetZeros) {
			// Found valid block!
			m.calculateHashRate()
			m.printSuccess(blockHash, block)
			return block, nil
		}

		// Print progress every 100,000 attempts
		if m.stats.Attempts%100000 == 0 {
			m.printProgress()
		}

		// Increment nonce
		nonce++

		// Check for overflow (shouldn't happen with low difficulty)
		if nonce == 0 {
			return nil, fmt.Errorf("nonce overflow - difficulty too high")
		}
	}
}

// checkProofOfWork checks if hash meets difficulty target
func (m *Miner) checkProofOfWork(blockHash []byte, targetZeros int) bool {
	// Count leading zero bytes
	leadingZeros := 0
	for _, b := range blockHash {
		if b == 0 {
			leadingZeros++
		} else {
			break
		}
	}

	return leadingZeros >= targetZeros
}

// calculateHashRate calculates current hash rate
func (m *Miner) calculateHashRate() {
	elapsed := time.Since(m.stats.StartTime).Seconds()
	if elapsed > 0 {
		m.stats.HashRate = float64(m.stats.Attempts) / elapsed
	}
}

// printProgress prints mining progress
func (m *Miner) printProgress() {
	m.calculateHashRate()
	fmt.Printf("   ⛏️  Attempts: %d | Hash rate: %.0f H/s | Nonce: %d\n",
		m.stats.Attempts,
		m.stats.HashRate,
		m.stats.CurrentNonce,
	)
}

// printSuccess prints success message
func (m *Miner) printSuccess(blockHash types.Hash, block *types.Block) {
	elapsed := time.Since(m.stats.StartTime)

	fmt.Printf("\n✅ Block mined successfully!\n")
	fmt.Printf("   Block hash: %s\n", blockHash)
	fmt.Printf("   Nonce: %d\n", block.Header.Nonce)
	fmt.Printf("   Attempts: %d\n", m.stats.Attempts)
	fmt.Printf("   Time: %v\n", elapsed)
	fmt.Printf("   Hash rate: %.0f H/s\n", m.stats.HashRate)
	fmt.Printf("   Transactions: %d\n", len(block.Transactions))
}

// GetStats returns current mining stats
func (m *Miner) GetStats() MiningStats {
	return m.stats
}

// ValidateMinedBlock validates a mined block
func ValidateMinedBlock(block *types.Block, targetZeros int) error {
	// Hash the header
	blockHash, err := serialization.HashBlockHeader(&block.Header)
	if err != nil {
		return err
	}

	// Check proof of work
	leadingZeros := 0
	for _, b := range blockHash {
		if b == 0 {
			leadingZeros++
		} else {
			break
		}
	}

	if leadingZeros < targetZeros {
		return fmt.Errorf("insufficient proof of work: %d < %d", leadingZeros, targetZeros)
	}

	// Validate with consensus rules
	if !validation.IsValidProofOfWork(blockHash[:], block.Header.Bits) {
		return fmt.Errorf("invalid proof of work")
	}

	return nil
}

/*
**Proof-of-Work Mining Explained:**

1. **The Goal:**
   - Find a nonce that makes the block hash start with N zeros
   - Hash function is SHA256 (deterministic but unpredictable)
   - Only way to find valid hash: try different nonces

2. **Mining Process:**
   ```
   nonce = 0
   while true:
       header.nonce = nonce
       hash = SHA256(SHA256(header))
       if hash starts with N zeros:
           SUCCESS! Broadcast block
       nonce++
   ```

3. **Difficulty:**
   - More zeros = harder to find
   - Bitcoin adjusts every 2016 blocks (~2 weeks)
   - Target: 10 minutes per block
   - Current difficulty: ~50 trillion trillion hashes

4. **Why This Works:**
   - Can't predict hash output
   - Can't reverse engineer nonce from hash
   - Must try every possibility
   - Proof you did the work!

**Example:**
```
Target: 2 leading zero bytes

Attempt 1:
  Nonce: 0
  Hash: a7f3c2... ❌ (no leading zeros)

Attempt 2:
  Nonce: 1
  Hash: 3b9e1a... ❌

...

Attempt 45,231:
  Nonce: 45231
  Hash: 0000a3... ✅ (2 leading zeros!)
```

**Difficulty Levels:**
```
0 zeros: ~1 attempt (instant)
1 zero:  ~256 attempts
2 zeros: ~65,536 attempts
3 zeros: ~16 million attempts
4 zeros: ~4 billion attempts

Bitcoin mainnet: ~20 leading zeros!
```

**Visual:**
```
Block Header (80 bytes):
┌────────────────────────────────┐
│ Version | PrevHash | MerkleRoot│
│ Time | Bits | Nonce: 0         │
└────────────────────────────────┘
         ↓ SHA256(SHA256())
    Hash: a7f3c2... ❌

Nonce: 1
         ↓ SHA256(SHA256())
    Hash: 3b9e1a... ❌

Nonce: 45231
         ↓ SHA256(SHA256())
    Hash: 0000a3... ✅ FOUND!
```

**Hash Rate:**
- Measures mining speed
- Unit: H/s (hashes per second)
- Modern ASICs: 100+ TH/s (trillion hashes/sec)
- This demo: ~1-10 MH/s (million hashes/sec)
*/
