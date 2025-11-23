package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/network"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
)

func main() {
	// Parse command line arguments
	port := flag.Int("port", 8333, "Port to listen on")
	dbPath := flag.String("db", "data/blockchain_p8", "Path to blockchain database")
	connect := flag.String("connect", "", "Peer to connect to (optional)")
	flag.Parse()

	fmt.Printf("Starting Bitcoin Node (Phase 8)...\n")
	fmt.Printf("Database: %s\n", *dbPath)
	fmt.Printf("Port: %d\n", *port)

	// Initialize blockchain
	chain, err := storage.NewBlockchainStorage(*dbPath)
	if err != nil {
		fmt.Printf("Failed to open blockchain: %v\n", err)
		os.Exit(1)
	}
	defer chain.Close()

	// Check chain state
	count, _ := chain.GetBlockCount()
	fmt.Printf("Current block count: %d\n", count)

	// Configure node
	config := network.NodeConfig{
		ListenAddr: fmt.Sprintf(":%d", *port),
		UserAgent:  "/LearnBitcoin:0.1/",
	}

	if *connect != "" {
		config.SeedNodes = []string{*connect}
	}

	// Create and start node
	node := network.NewNode(config, chain)

	if err := node.Start(); err != nil {
		fmt.Printf("Failed to start node: %v\n", err)
		os.Exit(1)
	}
	defer node.Stop()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Node running. Press Ctrl+C to stop.")

	// Keep running until signal
	<-sigChan
	fmt.Println("\nShutting down...")
}
