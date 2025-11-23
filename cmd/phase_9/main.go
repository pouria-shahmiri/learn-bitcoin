package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/mining"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/rpc"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/utxo"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/wallet"
)

func main() {
	fmt.Println("=== Phase 9: Wallet + RPC/CLI Demo ===\n")

	// Initialize blockchain storage
	chain, err := storage.NewBlockchainStorage("./data/phase9")
	if err != nil {
		log.Fatal(err)
	}
	defer chain.Close()

	// Create wallet
	w := wallet.NewWallet()

	// Demo 1: Generate addresses
	fmt.Println("--- Demo 1: Generate Addresses ---")
	addr1, _ := w.GenerateAddress()
	addr2, _ := w.GenerateAddress()
	fmt.Printf("Address 1: %s\n", addr1)
	fmt.Printf("Address 2: %s\n", addr2)
	fmt.Printf("Initial balance: %d satoshis\n\n", w.GetBalance())

	// Demo 2: Initialize genesis block if needed
	isEmpty, _ := chain.IsEmpty()
	if isEmpty {
		fmt.Println("--- Creating Genesis Block ---")
		genesis := createGenesisBlock(addr1)
		if err := chain.SaveBlock(genesis, 0); err != nil {
			log.Fatal(err)
		}

		// Add genesis coinbase to wallet
		coinbaseTx := genesis.Transactions[0]
		txHash, _ := serialization.HashTransaction(&coinbaseTx)
		coinbaseUTXO := utxo.NewUTXO(
			txHash,
			0,
			coinbaseTx.Outputs[0],
			0,
			true,
		)
		w.AddUTXO(coinbaseUTXO)

		blockHash, _ := chain.GetBlockHash(genesis)
		fmt.Printf("Genesis block created: %s\n", blockHash)
		fmt.Printf("Wallet balance: %d satoshis\n\n", w.GetBalance())
	}

	// Demo 3: Mine additional blocks
	fmt.Println("--- Demo 2: Mine Blocks to Fund Wallet ---")
	currentHeight, _ := chain.GetBestBlockHeight()

	for i := 0; i < 2; i++ {
		fmt.Printf("Mining block %d to %s...\n", i+1, addr1)

		// Get previous block
		prevBlock, _, _ := chain.GetBestBlock()
		prevHash, _ := chain.GetBlockHash(prevBlock)

		// Create coinbase transaction
		newHeight := currentHeight + uint64(i) + 1
		coinbase, _ := mining.CreateCoinbase(newHeight, 0, addr1, uint64(i))

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
		miner := mining.NewMiner()
		block, err := miner.MineBlock(template, 1) // 1 leading zero byte
		if err != nil {
			log.Fatal(err)
		}

		// Save block
		if err := chain.SaveBlock(block, newHeight); err != nil {
			log.Fatal(err)
		}

		// Add coinbase UTXO to wallet
		coinbaseTx := block.Transactions[0]
		txHash, _ := serialization.HashTransaction(&coinbaseTx)
		coinbaseUTXO := utxo.NewUTXO(
			txHash,
			0,
			coinbaseTx.Outputs[0],
			newHeight,
			true,
		)
		w.AddUTXO(coinbaseUTXO)

		blockHash, _ := chain.GetBlockHash(block)
		fmt.Printf("  Block mined: %s (height: %d)\n", blockHash, newHeight)
		fmt.Printf("  Wallet balance: %d satoshis\n", w.GetBalance())
	}
	fmt.Println()

	// Demo 4: Create and send transaction
	fmt.Println("--- Demo 3: Create Transaction ---")
	sendAmount := int64(2500000000) // 25 BTC
	fmt.Printf("Creating transaction: %d satoshis to %s\n", sendAmount, addr2)

	tx, err := w.Send(addr2, sendAmount)
	if err != nil {
		log.Printf("Failed to create transaction: %v", err)
		log.Println("(This is expected if wallet doesn't have enough mature coins)")
	} else {
		txHash, _ := serialization.HashTransaction(tx)
		fmt.Printf("Transaction created: %s\n", txHash)
		fmt.Printf("  Inputs: %d\n", len(tx.Inputs))
		fmt.Printf("  Outputs: %d\n", len(tx.Outputs))
	}
	fmt.Println()

	// Demo 5: Start RPC server
	fmt.Println("--- Demo 4: RPC Server ---")
	fmt.Println("Starting RPC server on http://localhost:8332")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /getnewaddress")
	fmt.Println("  GET  /getbalance")
	fmt.Println("  POST /sendtoaddress")
	fmt.Println("  GET  /getblockcount")
	fmt.Println("  GET  /getblock?height=<height>")
	fmt.Println("  GET  /gettransaction?txhash=<hash>")
	fmt.Println("  GET  /listaddresses")
	fmt.Println()

	// Start server in goroutine
	server := rpc.NewServer(w, chain, ":8332")
	go func() {
		if err := server.Start(); err != nil {
			log.Printf("RPC server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Demo 6: Test RPC client
	fmt.Println("--- Demo 5: Test RPC Client ---")
	client := rpc.NewClient("http://localhost:8332")

	// Test getblockcount
	height, err := client.GetBlockCount()
	if err != nil {
		log.Printf("GetBlockCount error: %v", err)
	} else {
		fmt.Printf("Current height: %d\n", height)
	}

	// Test getbalance
	balance, err := client.GetBalance()
	if err != nil {
		log.Printf("GetBalance error: %v", err)
	} else {
		fmt.Printf("Wallet balance: %d satoshis (%.8f BTC)\n", balance, float64(balance)/100000000.0)
	}

	// Test listaddresses
	addresses, err := client.ListAddresses()
	if err != nil {
		log.Printf("ListAddresses error: %v", err)
	} else {
		fmt.Printf("Wallet addresses (%d):\n", len(addresses))
		for i, addr := range addresses {
			fmt.Printf("  %d. %s\n", i+1, addr)
		}
	}

	// Test getblock
	if height > 0 {
		blockInfo, err := client.GetBlock(height)
		if err != nil {
			log.Printf("GetBlock error: %v", err)
		} else {
			fmt.Printf("\nLatest block info:\n")
			fmt.Printf("  Hash: %s\n", blockInfo.Hash)
			fmt.Printf("  Height: %d\n", blockInfo.Height)
			fmt.Printf("  Transactions: %d\n", len(blockInfo.Transactions))
		}
	}

	fmt.Println("\n=== Phase 9 Demo Complete ===")
	fmt.Println("\nRPC server is running. You can test it with:")
	fmt.Println("  curl http://localhost:8332/getbalance")
	fmt.Println("  curl http://localhost:8332/getblockcount")
	fmt.Println("  curl http://localhost:8332/getnewaddress")
	fmt.Println("\nOr use the CLI tool:")
	fmt.Println("  go run cmd/bitcoin-cli/main.go getbalance")
	fmt.Println("  go run cmd/bitcoin-cli/main.go getblockcount")
	fmt.Println("  go run cmd/bitcoin-cli/main.go listaddresses")
	fmt.Println("\nPress Ctrl+C to exit...")

	// Keep server running
	select {}
}

func createGenesisBlock(minerAddr string) *types.Block {
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
