package network

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/mempool"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/network/peer"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/network/protocol"
	syncmanager "github.com/pouria-shahmiri/learn-bitcoin/pkg/network/sync"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// Node represents a P2P node
type Node struct {
	Config      NodeConfig
	Blockchain  *storage.BlockchainStorage
	Mempool     *mempool.Mempool
	SyncManager *syncmanager.SyncManager

	peers    map[string]*peer.Peer
	peerLock sync.RWMutex

	quit chan struct{}
	wg   sync.WaitGroup
}

// NodeConfig holds configuration
type NodeConfig struct {
	ListenAddr string
	SeedNodes  []string
	UserAgent  string
}

// NewNode creates a new node
func NewNode(config NodeConfig, chain *storage.BlockchainStorage) *Node {
	// Mempool config: 300MB max size, 1 sat/byte min fee, 14 days max age
	mp := mempool.NewMempool(300*1024*1024, 1, 14*24*60*60)
	return &Node{
		Config:      config,
		Blockchain:  chain,
		Mempool:     mp,
		SyncManager: syncmanager.NewSyncManager(chain),
		peers:       make(map[string]*peer.Peer),
		quit:        make(chan struct{}),
	}
}

// Start starts the node
func (n *Node) Start() error {
	// Start listening
	listener, err := net.Listen("tcp", n.Config.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	n.wg.Add(1)
	go n.acceptLoop(listener)

	// Connect to seeds
	for _, seed := range n.Config.SeedNodes {
		go n.Connect(seed)
	}

	fmt.Printf("Node started on %s\n", n.Config.ListenAddr)
	return nil
}

// Stop stops the node
func (n *Node) Stop() {
	close(n.quit)

	n.peerLock.Lock()
	for _, p := range n.peers {
		p.Stop()
	}
	n.peerLock.Unlock()

	n.wg.Wait()
}

// Connect connects to a peer
func (n *Node) Connect(address string) {
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		fmt.Printf("Failed to connect to %s: %v\n", address, err)
		return
	}

	n.handlePeer(conn, false)
}

// acceptLoop accepts incoming connections
func (n *Node) acceptLoop(listener net.Listener) {
	defer n.wg.Done()
	defer listener.Close()

	for {
		select {
		case <-n.quit:
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go n.handlePeer(conn, true)
		}
	}
}

// handlePeer handles a new peer connection
func (n *Node) handlePeer(conn net.Conn, inbound bool) {
	p := peer.NewPeer(conn, inbound)

	n.peerLock.Lock()
	n.peers[p.Address()] = p
	n.peerLock.Unlock()

	fmt.Printf("New peer connected: %s (inbound=%v)\n", p.Address(), inbound)

	p.Start()

	// Initiate handshake if outbound
	if !inbound {
		localAddr := protocol.NetAddress{
			Services: protocol.SFNodeNetwork,
			IP:       [16]byte{}, // TODO: Get local IP
			Port:     0,          // TODO: Get local port
		}
		remoteAddr := protocol.NetAddress{
			Services: protocol.SFNodeNetwork,
			IP:       [16]byte{},
			Port:     0,
		}

		// Get best height
		height, _ := n.Blockchain.GetBestBlockHeight()

		version := protocol.NewVersionMessage(
			remoteAddr,
			localAddr,
			0, // Nonce
			n.Config.UserAgent,
			int32(height),
		)

		p.Handshake(version)
	}

	// Handle messages
	n.handleMessages(p)

	// Cleanup
	n.peerLock.Lock()
	delete(n.peers, p.Address())
	n.peerLock.Unlock()
	p.Stop()
	fmt.Printf("Peer disconnected: %s\n", p.Address())
}

// handleMessages processes messages from a peer
func (n *Node) handleMessages(p *peer.Peer) {
	for {
		select {
		case msg := <-p.Receive:
			if err := n.processMessage(p, msg); err != nil {
				fmt.Printf("Error processing message from %s: %v\n", p.Address(), err)
			}
		case <-p.Quit:
			return
		case <-n.quit:
			return
		}
	}
}

// processMessage handles a single message
func (n *Node) processMessage(p *peer.Peer, msg *protocol.Message) error {
	// fmt.Printf("Received %s from %s\n", msg.Command, p.Address())

	switch msg.Command {
	case protocol.CmdVersion:
		return n.handleVersion(p, msg.Payload)

	case protocol.CmdVerAck:
		p.VerAckReceived = true
		// Start sync after handshake
		return n.SyncManager.StartSync(p)

	case protocol.CmdInv:
		inv, err := protocol.DeserializeInv(msg.Payload)
		if err != nil {
			return err
		}
		return n.SyncManager.HandleInv(inv, p)

	case protocol.CmdGetData:
		gd, err := protocol.DeserializeGetData(msg.Payload)
		if err != nil {
			return err
		}
		return n.handleGetData(p, gd)

	case protocol.CmdBlock:
		// Deserialize block
		block, err := serialization.DeserializeBlock(msg.Payload)
		if err != nil {
			return fmt.Errorf("failed to deserialize block: %w", err)
		}

		return n.SyncManager.HandleBlock(block, p)

	case protocol.CmdGetBlocks:
		gb, err := protocol.DeserializeGetBlocks(msg.Payload)
		if err != nil {
			return err

		}
		return n.handleGetBlocks(p, gb)

	case protocol.CmdTx:
		// Deserialize transaction
		tx, err := serialization.DeserializeTransaction(bytes.NewReader(msg.Payload))
		if err != nil {
			return fmt.Errorf("failed to deserialize transaction: %w", err)
		}
		return n.handleTx(p, tx)

	default:
		// fmt.Printf("Unknown command: %s\n", msg.Command)
	}

	return nil
}

func (n *Node) handleVersion(p *peer.Peer, payload []byte) error {
	v, err := protocol.DeserializeVersion(payload)
	if err != nil {
		return err
	}

	p.Version = v

	// Send VerAck
	p.SendMessage(protocol.NewMessage(protocol.MagicMainnet, protocol.CmdVerAck, nil))

	// If inbound, send our Version
	if p.Inbound {
		localAddr := protocol.NetAddress{}
		remoteAddr := v.AddrFrom
		height, _ := n.Blockchain.GetBestBlockHeight()

		myVersion := protocol.NewVersionMessage(
			remoteAddr,
			localAddr,
			0,
			n.Config.UserAgent,
			int32(height),
		)
		p.Handshake(myVersion)
	}

	return nil
}

func (n *Node) handleGetData(p *peer.Peer, gd *protocol.GetDataMessage) error {
	for _, vect := range gd.Inventory {
		if vect.Type == protocol.InvTypeBlock {
			// Send block
			block, err := n.Blockchain.GetBlock(vect.Hash)
			if err != nil {
				// Not found
				continue
			}

			// Serialize block
			serialized, err := serialization.SerializeBlock(block)
			if err != nil {
				continue
			}

			p.SendMessage(protocol.NewMessage(protocol.MagicMainnet, protocol.CmdBlock, serialized))
		}
	}
	return nil
}

func (n *Node) handleGetBlocks(p *peer.Peer, gb *protocol.GetBlocksMessage) error {
	// Find the latest block we have that is in their locator
	var startHash types.Hash
	for _, hash := range gb.BlockLocator {
		exists, _ := n.Blockchain.HasBlock(hash)
		if exists {
			startHash = hash
			break
		}
	}

	if startHash.IsZero() {
		// No common block found, start from genesis?
		// Or maybe we don't have anything they have.
		return nil
	}

	// Send inv for blocks starting from startHash
	// Limit to 500 blocks
	inv := protocol.NewInvMessage()

	// Iterate forward from startHash
	// We need a way to get next block hash.
	// Our storage doesn't support "GetNextBlock".
	// We only have "GetBlockByHeight".

	height, err := n.Blockchain.GetBlockHeight(startHash)
	if err != nil {
		return err
	}

	// Send up to 500 blocks
	for i := uint64(1); i <= 500; i++ {
		block, err := n.Blockchain.GetBlockByHeight(height + i)
		if err != nil {
			break // End of chain
		}

		hash, _ := n.Blockchain.GetBlockHash(block)
		inv.AddInvVect(protocol.NewInvVect(protocol.InvTypeBlock, hash))

		if hash == gb.HashStop {
			break
		}
	}

	if len(inv.Inventory) > 0 {
		p.SendMessage(protocol.NewMessage(protocol.MagicMainnet, protocol.CmdInv, mustSerialize(inv)))
	}

	return nil
}

func (n *Node) handleTx(p *peer.Peer, tx *types.Transaction) error {
	// Calculate fee
	fee, err := n.calculateTxFee(tx)
	if err != nil {
		// Could not calculate fee (e.g. inputs not found)
		// In a real node we might request missing inputs
		return nil
	}

	// Get current height
	height, _ := n.Blockchain.GetBestBlockHeight()

	// Add to mempool
	if err := n.Mempool.Add(tx, fee, height); err != nil {
		// Already exists or invalid
		return nil
	}

	fmt.Printf("Received new transaction from %s\n", p.Address())

	// Relay to other peers
	n.RelayTransaction(tx, p.Address())

	return nil
}

// calculateTxFee calculates transaction fee by looking up inputs
func (n *Node) calculateTxFee(tx *types.Transaction) (int64, error) {
	inputValues := make([]int64, len(tx.Inputs))

	for i, input := range tx.Inputs {
		// Find previous transaction
		prevTxLoc, _, err := n.Blockchain.GetTransactionLocation(input.PrevTxHash)
		if err != nil {
			return 0, fmt.Errorf("prev tx not found: %w", err)
		}

		prevBlock, err := n.Blockchain.GetBlock(prevTxLoc)
		if err != nil {
			return 0, fmt.Errorf("prev block not found: %w", err)
		}

		// Find transaction in block
		var prevTx *types.Transaction
		for _, t := range prevBlock.Transactions {
			hash, _ := serialization.HashTransaction(&t)
			if hash == input.PrevTxHash {
				prevTx = &t
				break
			}
		}

		if prevTx == nil {
			return 0, fmt.Errorf("prev tx not found in block")
		}

		if int(input.OutputIndex) >= len(prevTx.Outputs) {
			return 0, fmt.Errorf("output index out of range")
		}

		inputValues[i] = prevTx.Outputs[input.OutputIndex].Value
	}

	return mempool.CalculateTransactionFee(tx, inputValues)
}

// RelayTransaction broadcasts a transaction to all peers except the source
func (n *Node) RelayTransaction(tx *types.Transaction, sourceAddr string) {
	txHash, _ := serialization.HashTransaction(tx)
	inv := protocol.NewInvMessage()
	inv.AddInvVect(protocol.NewInvVect(protocol.InvTypeTx, txHash))

	msg := protocol.NewMessage(protocol.MagicMainnet, protocol.CmdInv, mustSerialize(inv))

	n.peerLock.RLock()
	defer n.peerLock.RUnlock()

	for _, p := range n.peers {
		if p.Address() != sourceAddr {
			p.SendMessage(msg)
		}
	}
}

func mustSerialize(msg interface{ Serialize() ([]byte, error) }) []byte {
	b, err := msg.Serialize()
	if err != nil {
		panic(err)
	}
	return b
}
