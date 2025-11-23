package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/config"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/mining"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/network"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/rpc"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/utxo"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/wallet"
)

// Node represents a full Bitcoin node
type Node struct {
	config    *config.NodeConfig
	chain     *storage.BlockchainStorage
	wallet    *wallet.Wallet
	p2pServer *network.Server
	rpcServer *rpc.Server
	miner     *mining.Miner
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func main() {
	// Load configuration from environment
	cfg := config.LoadFromEnv()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Print configuration
	logInfo("=== Bitcoin Node Starting ===")
	logInfo(cfg.String())
	logInfo("")

	// Create and start node
	node, err := NewNode(cfg)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	// Start the node
	if err := node.Start(); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logInfo("Shutdown signal received, stopping node...")
	node.Stop()
	logInfo("Node stopped gracefully")
}

// NewNode creates a new Bitcoin node
func NewNode(cfg *config.NodeConfig) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize blockchain storage
	chain, err := storage.NewBlockchainStorage(cfg.DataDir)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize blockchain: %w", err)
	}

	// Create wallet
	w := wallet.NewWallet()

	// Initialize genesis block if needed
	isEmpty, _ := chain.IsEmpty()
	if isEmpty {
		logInfo("Blockchain is empty, creating genesis block...")
		genesis := createGenesisBlock(cfg.MinerAddress)
		if err := chain.SaveBlock(genesis, 0); err != nil {
			chain.Close()
			cancel()
			return nil, fmt.Errorf("failed to save genesis block: %w", err)
		}

		// Add genesis coinbase to wallet if this is a miner
		if cfg.MiningEnabled {
			coinbaseTx := genesis.Transactions[0]
			txHash, _ := serialization.HashTransaction(&coinbaseTx)
			coinbaseUTXO := utxo.NewUTXO(txHash, 0, coinbaseTx.Outputs[0], 0, true)
			w.AddUTXO(coinbaseUTXO)
		}

		blockHash, _ := chain.GetBlockHash(genesis)
		logInfo(fmt.Sprintf("Genesis block created: %s", blockHash))
	}

	// Create P2P server
	p2pServer := network.NewServer(cfg.GetP2PAddress(), chain)

	// Create RPC server
	rpcServer := rpc.NewServer(w, chain, cfg.GetRPCAddress())

	// Create miner if mining is enabled
	var miner *mining.Miner
	if cfg.MiningEnabled {
		miner = mining.NewMiner()
	}

	return &Node{
		config:    cfg,
		chain:     chain,
		wallet:    w,
		p2pServer: p2pServer,
		rpcServer: rpcServer,
		miner:     miner,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Start starts all node components
func (n *Node) Start() error {
	logInfo("Starting node components...")

	// Start P2P server
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		logInfo(fmt.Sprintf("Starting P2P server on %s", n.config.GetP2PAddress()))
		if err := n.p2pServer.Start(); err != nil {
			logError(fmt.Sprintf("P2P server error: %v", err))
		}
	}()

	// Give P2P server time to start
	time.Sleep(500 * time.Millisecond)

	// Connect to initial peers
	for _, peer := range n.config.InitialPeers {
		logInfo(fmt.Sprintf("Connecting to peer: %s", peer))
		n.wg.Add(1)
		go func(addr string) {
			defer n.wg.Done()
			time.Sleep(2 * time.Second) // Wait for other nodes to start
			if err := n.p2pServer.ConnectToPeer(addr); err != nil {
				logWarn(fmt.Sprintf("Failed to connect to peer %s: %v", addr, err))
			} else {
				logInfo(fmt.Sprintf("Connected to peer: %s", addr))
			}
		}(peer)
	}

	// Start RPC server
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		logInfo(fmt.Sprintf("Starting RPC server on %s", n.config.GetRPCAddress()))
		if err := n.rpcServer.Start(); err != nil {
			logError(fmt.Sprintf("RPC server error: %v", err))
		}
	}()

	// Start auto-mining if enabled
	if n.config.MiningEnabled && n.config.AutoMine {
		n.wg.Add(1)
		go func() {
			defer n.wg.Done()
			n.autoMineLoop()
		}()
	}

	// Start status reporter
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		n.statusReporter()
	}()

	logInfo("Node started successfully!")
	logInfo(fmt.Sprintf("RPC endpoint: http://localhost%s", n.config.GetRPCAddress()))
	logInfo(fmt.Sprintf("P2P listening on: %s", n.config.GetP2PAddress()))
	if n.config.MiningEnabled {
		logInfo(fmt.Sprintf("Mining enabled: %v (Auto: %v, Interval: %v)",
			n.config.MiningEnabled, n.config.AutoMine, n.config.MineInterval))
	}
	logInfo("")

	return nil
}

// Stop stops all node components
func (n *Node) Stop() {
	n.cancel()

	// Stop P2P server
	if n.p2pServer != nil {
		n.p2pServer.Stop()
	}

	// Close blockchain storage
	if n.chain != nil {
		n.chain.Close()
	}

	// Wait for all goroutines to finish
	n.wg.Wait()
}

// autoMineLoop automatically mines blocks at regular intervals
func (n *Node) autoMineLoop() {
	logInfo(fmt.Sprintf("Auto-mining started (interval: %v)", n.config.MineInterval))
	ticker := time.NewTicker(n.config.MineInterval)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			if err := n.mineBlock(); err != nil {
				logError(fmt.Sprintf("Mining error: %v", err))
			}
		}
	}
}

// mineBlock mines a single block
func (n *Node) mineBlock() error {
	// Get current height
	currentHeight, err := n.chain.GetBestBlockHeight()
	if err != nil {
		return fmt.Errorf("failed to get best block height: %w", err)
	}

	// Get previous block
	prevBlock, _, err := n.chain.GetBestBlock()
	if err != nil {
		return fmt.Errorf("failed to get best block: %w", err)
	}

	prevHash, err := n.chain.GetBlockHash(prevBlock)
	if err != nil {
		return fmt.Errorf("failed to get block hash: %w", err)
	}

	// Create coinbase transaction
	newHeight := currentHeight + 1
	coinbase, err := mining.CreateCoinbase(newHeight, 0, n.config.MinerAddress, newHeight)
	if err != nil {
		return fmt.Errorf("failed to create coinbase: %w", err)
	}

	// Create block template
	template := &mining.BlockTemplate{
		Version:       1,
		PrevBlockHash: prevHash,
		Transactions:  []types.Transaction{*coinbase},
		Timestamp:     uint32(time.Now().Unix()),
		Bits:          0x1d00ffff,
		Height:        newHeight,
		TotalFees:     0,
	}

	// Mine block
	startTime := time.Now()
	block, err := n.miner.MineBlock(template, 1) // 1 leading zero byte
	if err != nil {
		return fmt.Errorf("failed to mine block: %w", err)
	}
	miningTime := time.Since(startTime)

	// Save block
	if err := n.chain.SaveBlock(block, newHeight); err != nil {
		return fmt.Errorf("failed to save block: %w", err)
	}

	// Add coinbase UTXO to wallet
	coinbaseTx := block.Transactions[0]
	txHash, _ := serialization.HashTransaction(&coinbaseTx)
	coinbaseUTXO := utxo.NewUTXO(txHash, 0, coinbaseTx.Outputs[0], newHeight, true)
	n.wallet.AddUTXO(coinbaseUTXO)

	blockHash, _ := n.chain.GetBlockHash(block)
	logInfo(fmt.Sprintf("[%s] Mined block %d: %s (time: %v, nonce: %d)",
		n.config.NodeID, newHeight, blockHash, miningTime, block.Header.Nonce))

	// Broadcast block to peers
	if n.p2pServer != nil {
		n.p2pServer.BroadcastBlock(block)
	}

	return nil
}

// statusReporter periodically reports node status
func (n *Node) statusReporter() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			height, _ := n.chain.GetBestBlockHeight()
			peerCount := n.p2pServer.GetPeerCount()
			balance := n.wallet.GetBalance()

			logInfo(fmt.Sprintf("[%s] Status - Height: %d, Peers: %d, Balance: %d sats",
				n.config.NodeID, height, peerCount, balance))
		}
	}
}

// createGenesisBlock creates the genesis block
func createGenesisBlock(minerAddr string) *types.Block {
	// Use a default address if none provided
	if minerAddr == "" {
		minerAddr = "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	}

	// Create genesis coinbase
	coinbase, _ := mining.CreateCoinbase(0, 0, minerAddr, 0)

	// Create block template
	template := &mining.BlockTemplate{
		Version:       1,
		PrevBlockHash: types.Hash{},
		Transactions:  []types.Transaction{*coinbase},
		Timestamp:     uint32(time.Now().Unix()),
		Bits:          0x1d00ffff,
		Height:        0,
		TotalFees:     0,
	}

	// Mine the genesis block
	miner := mining.NewMiner()
	block, _ := miner.MineBlock(template, 1) // 1 leading zero byte

	return block
}

// Logging helpers
func logInfo(msg string) {
	log.Printf("[INFO] %s", msg)
}

func logWarn(msg string) {
	log.Printf("[WARN] %s", msg)
}

func logError(msg string) {
	log.Printf("[ERROR] %s", msg)
}
