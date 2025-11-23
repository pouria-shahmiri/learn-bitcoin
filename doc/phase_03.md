# Phase 3: Persistence & CLI

## 1. Objectives
In previous phases, the blockchain lived only in the computer's RAM. If you stopped the program, the data was lost. Phase 3 introduces **Persistence** using a database to save blocks to disk. We also build a **Command Line Interface (CLI)** to interact with the blockchain, allowing us to add blocks and inspect the chain without modifying code.

## 2. Key Concepts

### Database (LevelDB)
We use **LevelDB**, a fast key-value storage library developed by Google (and used in Bitcoin Core).
*   **Key-Value Store**: Unlike SQL, we just store `Key -> Value`.
*   **Schema**:
    *   `"l"` -> Hash of the last block (The "Tip" of the chain).
    *   `Block Hash` -> Serialized Block Data.

### Serialization (Gob)
To store a Go struct (like `Block`) in the database, we must convert it into a byte array. We use Go's `encoding/gob` package for this.
*   **Serialize**: Struct -> Bytes (for storage).
*   **Deserialize**: Bytes -> Struct (for reading).

### CLI (Command Line Interface)
A way to run commands like `addblock` or `printchain`. It parses arguments provided by the user and calls the appropriate blockchain methods.

## 3. Code Implementation

### `Database` Wrapper
We wrap LevelDB in `pkg/storage/database.go` to handle opening, closing, and basic Get/Put operations.

```go
func OpenDatabase(path string) (*Database, error) {
    // ... opens LevelDB file at specific path
}
```

### `BlockchainIterator`
Since we can't just loop through an array anymore (blocks are scattered on disk), we use an iterator to traverse the chain backwards from the tip.

```go
type BlockchainIterator struct {
    currentHash []byte
    db          *Database
}

func (i *BlockchainIterator) Next() *Block {
    // 1. Get block bytes from DB using currentHash
    // 2. Deserialize bytes to Block struct
    // 3. Set currentHash to block.PrevBlockHash (move back one step)
    // 4. Return block
}
```

### CLI Structure
We use Go's `flag` package to parse arguments.

```go
func (cli *CLI) Run() {
    // Define flags
    addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
    
    // Parse
    switch os.Args[1] {
    case "addblock":
        addBlockCmd.Parse(os.Args[2:])
        cli.addBlock(*addBlockData)
    // ...
    }
}
```

## 4. Architecture Diagram

### Data Flow
How a user command travels to the disk.

```text
User Command: "addblock 'Hello'"
       |
       v
+-------------+      +----------------+      +------------------+
|     CLI     | ---> |   Blockchain   | ---> |     Database     |
| (Parse Arg) |      |  (Mine Block)  |      | (LevelDB Store)  |
+-------------+      +----------------+      +------------------+
                            |                         |
                            v                         v
                     1. Create Block           1. Store Block Hash -> Bytes
                     2. Run PoW                2. Update "l" -> New Hash
```

## 5. How to Run
First, build the CLI or run it directly.

```bash
# Initialize the chain (if needed)
go run cmd/phase_3/main.go createblockchain -address "YourAddress"

# Add a block
go run cmd/phase_3/main.go addblock -data "Send 5 BTC to Ivan"

# Print the chain
go run cmd/phase_3/main.go printchain
```

**Expected Output:**
```text
= Block 00f2... =
Prev: ...
Data: Send 5 BTC to Ivan
PoW: true

= Block 00a1... =
Prev: ...
Data: Genesis Block
PoW: true
```
