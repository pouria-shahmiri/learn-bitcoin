# Learn Bitcoin: A Go Implementation
Welcome to the documentation for the **Learn Bitcoin** project. This project is a step-by-step implementation of a Bitcoin-like cryptocurrency in Go, designed to teach the core concepts of blockchain technology, from basic hashing to a full Dockerized P2P network.

## Project Overview
This repository is divided into **11 Phases**, each adding a layer of complexity and functionality to the system. By following these phases, you build a complete, functional Bitcoin node from scratch.

## Documentation Index

### [Phase 1: Basic Blockchain Structure](./phase_01.md)
*   **Concepts**: Blocks, Hashing (SHA-256), Blockchain structure.
*   **Goal**: Create a simple chain of blocks linked by hashes.

### [Phase 2: Proof of Work](./phase_02.md)
*   **Concepts**: Mining difficulty, Nonce, Target, Consensus rules.
*   **Goal**: Implement the Proof-of-Work algorithm to secure the chain.

### [Phase 3: Persistence & CLI](./phase_03.md)
*   **Concepts**: Database storage (LevelDB/Badger), Serialization (Gob/JSON), Command Line Interface.
*   **Goal**: Persist the blockchain to disk and interact with it via CLI.

### [Phase 4: Transactions](./phase_04.md)
*   **Concepts**: Inputs, Outputs, ScriptSig, ScriptPubKey, Transaction ID.
*   **Goal**: Implement the transaction structure and basic validation.

### [Phase 5: UTXO Set & Validation](./phase_05.md)
*   **Concepts**: Unspent Transaction Outputs (UTXO), Chain Validation, Double Spend Protection.
*   **Goal**: Maintain a UTXO set for fast balance lookups and validation.

### [Phase 6: Mempool & Fee Policy](./phase_06.md)
*   **Concepts**: Memory Pool, Transaction Fees, Transaction Selection.
*   **Goal**: Store unconfirmed transactions and select them for mining based on fees.

### [Phase 7: Mining & Coinbase](./phase_07.md)
*   **Concepts**: Coinbase Transaction, Block Reward, Merkle Tree.
*   **Goal**: Construct valid blocks with transactions and mining rewards.

### [Phase 8: P2P Network](./phase_08.md)
*   **Concepts**: TCP/IP, Handshake, Message Protocol (Version, Inv, GetData, Block).
*   **Goal**: Connect nodes to form a peer-to-peer network and sync blocks.

### [Phase 9: Wallet & RPC](./phase_09.md)
*   **Concepts**: Private/Public Keys (ECC), Addresses, Digital Signatures, JSON-RPC.
*   **Goal**: Create a wallet to manage funds and an API for external interaction.

### [Phase 10: Hardening & Reorgs](./phase_10.md)
*   **Concepts**: Chain Reorganization, Fork Resolution, Orphan Blocks.
*   **Goal**: Handle network inconsistencies and competing chains robustly.

### [Phase 11: Docker Deployment](./phase_11.md)
*   **Concepts**: Containerization, Orchestration (Docker Compose), Integration Testing.
*   **Goal**: Deploy a multi-node network in isolated containers.

## Getting Started
To start exploring the code, check out the `cmd/` directory for the entry points of each phase.
```bash
# Example: Run Phase 1
go run cmd/phase_1/main.go
```

## Diagrams
Each phase documentation includes **Mermaid diagrams** to visualize the architecture and flow.
