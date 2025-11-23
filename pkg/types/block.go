package types

// BlockHeader is the metadata miners hash
type BlockHeader struct {
	Version       int32  // Block version (consensus rules)
	PrevBlockHash Hash   // Previous block hash (creates the chain!)
	MerkleRoot    Hash   // Root of transaction Merkle tree
	Timestamp     uint32 // When block was mined (Unix time)
	Bits          uint32 // Difficulty target (compact form)
	Nonce         uint32 // Mining nonce (miners increment this)
}

// Block is a complete block
type Block struct {
	Header       BlockHeader   // 80 bytes that get hashed
	Transactions []Transaction // All transactions in block
}

/*

**Critical understanding:**

1. **BlockHeader (80 bytes):**
   - This is what miners hash during Proof-of-Work
   - Changing any field (especially nonce) changes the hash
   - Miners search for a nonce that makes the hash start with zeros

2. **PrevBlockHash:**
   - Points to parent block
   - Creates the chain: Genesis → Block 1 → Block 2 → ...
   - Changing an old block breaks all future blocks (immutability!)

3. **MerkleRoot:**
   - Single hash representing ALL transactions
   - Allows efficient verification without downloading all transactions

4. **Bits:**
   - Difficulty target in compact form
   - Lower target = harder to mine
   - Adjusts every 2016 blocks (~2 weeks)

**Visual:**

Block Header (80 bytes):
[Version][PrevHash][MerkleRoot][Time][Bits][Nonce]
         ^
         |
    Links blocks together

*/
