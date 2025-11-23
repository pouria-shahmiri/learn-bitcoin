# Phase 10: Hardening & Reorgs

## 1. Objectives
In a decentralized network, two miners might solve a block at the same time. This creates a **Fork**. Phase 10 implements the logic to handle these forks and ensure all nodes eventually agree on the same history (Consensus). This is critical for network stability.

## 2. Key Concepts

### Chain Reorganization (Reorg)
When a node receives a new chain that is **longer** (more accumulated work) than its current active chain, it must switch.
1.  **Identify Fork Point**: Find the last common block.
2.  **Detach**: Disconnect blocks from the current tip back to the fork point.
3.  **Attach**: Connect blocks from the new chain starting from the fork point.
4.  **Mempool Update**: Transactions in the detached blocks that are NOT in the new chain must be returned to the mempool so they aren't lost.

### Orphan Blocks
Blocks whose parent is unknown.
*   **Scenario**: You receive Block 100 but you are only at Block 98. You are missing Block 99.
*   **Action**: Store Block 100 in an "Orphan Pool" and ask peers for Block 99.

## 3. Code Implementation

### `HandleBlock` Logic
Updated in `pkg/network` or `pkg/blockchain`.

```go
func (bc *Blockchain) AddBlock(block *Block) {
    if block.Height > bc.Height {
        // New longest chain!
        if block.PrevHash != bc.Tip {
            // Fork detected! The parent of this new block is NOT our current tip.
            bc.Reorganize(block)
        } else {
            // Normal append
            bc.Append(block)
        }
    }
}
```

## 4. Architecture Diagram

### Fork Resolution
Node switching from a shorter chain to a longer one.

```text
      (Common History)
          [ B0 ]
             |
          [ B1 ]
             |
      +------+------+
      |             |
   [ B2a ]       [ B2b ]
      |             |
   [ B3a ]       [ B3b ]
   (Old Tip)        |
                 [ B4b ]
                 (New Tip)

Action:
1. Detach B3a, B2a
2. Attach B2b, B3b, B4b
```

## 5. How to Run
This is best tested with the integration tests provided in the `tests/` folder, or by running multiple nodes and manually creating a split.

```bash
go test ./tests/...
```
