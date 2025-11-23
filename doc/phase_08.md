# Phase 8: P2P Network

## 1. Objectives
A blockchain is useless if it stays on one computer. Phase 8 builds the **Peer-to-Peer (P2P) Network**, allowing nodes to communicate, discover each other, and synchronize the blockchain. This turns our local database into a distributed system.

## 2. Key Concepts

### Node Roles
*   **Full Node**: Stores full chain, validates everything.
*   **Miner**: Mines new blocks.
*   **Wallet**: Lightweight, just manages keys (SPV).

### Protocol Messages
We implement a simplified Bitcoin protocol over TCP:
*   `version`: Handshake. "I am version X, my chain height is Y".
*   `getblocks`: "I am at block X, give me hashes of what comes next".
*   `inv` (Inventory): "I have these blocks/txs available".
*   `getdata`: "Send me the full data for this hash".
*   `block` / `tx`: The actual data payload.

## 3. Code Implementation

### `Server` Struct
Located in `pkg/network`. It manages the listening socket and known peers.

```go
type Server struct {
    nodeAddress string
    miningAddress string
    knownNodes []string
    // ...
}
```

### Handling Connections
We use Go's `net` package (TCP).

```go
func (s *Server) Start() {
    ln, _ := net.Listen("tcp", s.nodeAddress)
    for {
        conn, _ := ln.Accept()
        go s.handleConnection(conn)
    }
}
```

## 4. Architecture Diagram

### Synchronization Flow
How Node A gets new blocks from Node B.

```text
Node A (Height 10)                Node B (Height 15)
      |                                   |
      | --- version (H:10) -------------> |
      | <------------ version (H:15) ---- |
      |                                   |
      | --- getblocks (Last: 10) -------> |
      | <------------ inv (11,12...15) -- |
      |                                   |
      | --- getdata (Block 11) ---------> |
      | <------------ block (Data 11) --- |
      | (Validate & Add)                  |
      |                                   |
      | --- getdata (Block 12) ---------> |
      ...
```

## 5. How to Run
You need multiple terminals to simulate a network.

**Terminal 1 (Central Node):**
```bash
export NODE_ID=3000
go run cmd/phase_8/main.go createblockchain -address "CENTRAL_NODE"
go run cmd/phase_8/main.go startnode
```

**Terminal 2 (Wallet Node):**
```bash
export NODE_ID=3001
go run cmd/phase_8/main.go startnode
```
