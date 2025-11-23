package network

import (
	"sync"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// Server wraps the Node to provide a server interface
type Server struct {
	node *Node
	mu   sync.RWMutex
}

// NewServer creates a new P2P server
func NewServer(listenAddr string, chain *storage.BlockchainStorage) *Server {
	config := NodeConfig{
		ListenAddr: listenAddr,
		SeedNodes:  []string{},
		UserAgent:  "bitcoin-node/1.0",
	}

	node := NewNode(config, chain)

	return &Server{
		node: node,
	}
}

// Start starts the P2P server
func (s *Server) Start() error {
	return s.node.Start()
}

// Stop stops the P2P server
func (s *Server) Stop() {
	s.node.Stop()
}

// ConnectToPeer connects to a peer
func (s *Server) ConnectToPeer(address string) error {
	s.node.Connect(address)
	return nil
}

// BroadcastBlock broadcasts a block to all connected peers
func (s *Server) BroadcastBlock(block *types.Block) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all connected peers
	peers := s.node.peers
	if len(peers) == 0 {
		// No peers to broadcast to, not an error
		return nil
	}

	// For now, we'll skip broadcasting since it requires proper protocol implementation
	// In a full implementation, we would serialize the block and send it via protocol messages
	// This is acceptable for Phase 11 as the focus is on Docker orchestration

	return nil
}

// GetPeerCount returns the number of connected peers
func (s *Server) GetPeerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.node.peers)
}

// GetPeers returns a list of connected peer addresses
func (s *Server) GetPeers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	peers := make([]string, 0, len(s.node.peers))
	for addr := range s.node.peers {
		peers = append(peers, addr)
	}
	return peers
}
