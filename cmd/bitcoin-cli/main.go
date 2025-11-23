package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/rpc"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Default RPC server address
	rpcAddr := flag.String("rpcaddr", "http://localhost:8332", "RPC server address")
	flag.Parse()

	client := rpc.NewClient(*rpcAddr)
	command := flag.Arg(0)

	switch command {
	case "getnewaddress":
		handleGetNewAddress(client)
	case "getbalance":
		handleGetBalance(client)
	case "sendtoaddress":
		handleSendToAddress(client)
	case "getblockcount":
		handleGetBlockCount(client)
	case "getblock":
		handleGetBlock(client)
	case "gettransaction":
		handleGetTransaction(client)
	case "listaddresses":
		handleListAddresses(client)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Bitcoin CLI Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  bitcoin-cli [options] <command> [args...]")
	fmt.Println("\nOptions:")
	fmt.Println("  -rpcaddr <url>    RPC server address (default: http://localhost:8332)")
	fmt.Println("\nCommands:")
	fmt.Println("  getnewaddress                    Generate a new address")
	fmt.Println("  getbalance                       Show wallet balance")
	fmt.Println("  sendtoaddress <address> <amount> Send coins to address")
	fmt.Println("  getblockcount                    Show current blockchain height")
	fmt.Println("  getblock <height>                Retrieve block by height")
	fmt.Println("  gettransaction <txhash>          Retrieve transaction by hash")
	fmt.Println("  listaddresses                    List all wallet addresses")
}

func handleGetNewAddress(client *rpc.Client) {
	address, err := client.GetNewAddress()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("New address: %s\n", address)
}

func handleGetBalance(client *rpc.Client) {
	balance, err := client.GetBalance()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Balance: %d satoshis (%.8f BTC)\n", balance, float64(balance)/100000000.0)
}

func handleSendToAddress(client *rpc.Client) {
	if flag.NArg() < 3 {
		fmt.Println("Usage: sendtoaddress <address> <amount>")
		os.Exit(1)
	}

	address := flag.Arg(1)
	amountStr := flag.Arg(2)

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid amount: %v\n", err)
		os.Exit(1)
	}

	txHash, err := client.SendToAddress(address, amount)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Transaction sent successfully!\n")
	fmt.Printf("TxHash: %s\n", txHash)
}

func handleGetBlockCount(client *rpc.Client) {
	height, err := client.GetBlockCount()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Current height: %d\n", height)
}

func handleGetBlock(client *rpc.Client) {
	if flag.NArg() < 2 {
		fmt.Println("Usage: getblock <height>")
		os.Exit(1)
	}

	heightStr := flag.Arg(1)
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid height: %v\n", err)
		os.Exit(1)
	}

	block, err := client.GetBlock(height)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Block Information\n")
	fmt.Fprintf(w, "=================\n")
	fmt.Fprintf(w, "Hash:\t%s\n", block.Hash)
	fmt.Fprintf(w, "Height:\t%d\n", block.Height)
	fmt.Fprintf(w, "Version:\t%d\n", block.Version)
	fmt.Fprintf(w, "Previous Hash:\t%s\n", block.PrevHash)
	fmt.Fprintf(w, "Merkle Root:\t%s\n", block.MerkleRoot)
	fmt.Fprintf(w, "Timestamp:\t%d\n", block.Timestamp)
	fmt.Fprintf(w, "Bits:\t0x%08x\n", block.Bits)
	fmt.Fprintf(w, "Nonce:\t%d\n", block.Nonce)
	fmt.Fprintf(w, "Transactions:\t%d\n", len(block.Transactions))
	w.Flush()

	if len(block.Transactions) > 0 {
		fmt.Println("\nTransactions:")
		for i, txHash := range block.Transactions {
			fmt.Printf("  %d. %s\n", i+1, txHash)
		}
	}
}

func handleGetTransaction(client *rpc.Client) {
	if flag.NArg() < 2 {
		fmt.Println("Usage: gettransaction <txhash>")
		os.Exit(1)
	}

	txHash := flag.Arg(1)
	tx, err := client.GetTransaction(txHash)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Transaction Information\n")
	fmt.Fprintf(w, "=======================\n")
	fmt.Fprintf(w, "TxHash:\t%s\n", tx.TxHash)
	fmt.Fprintf(w, "Version:\t%d\n", tx.Version)
	fmt.Fprintf(w, "LockTime:\t%d\n", tx.LockTime)
	fmt.Fprintf(w, "\n")
	w.Flush()

	fmt.Printf("Inputs (%d):\n", len(tx.Inputs))
	for i, input := range tx.Inputs {
		fmt.Printf("  %d. %s:%d\n", i, input.PrevTxHash, input.OutputIndex)
		fmt.Printf("     ScriptSig: %s\n", input.ScriptSig)
		fmt.Printf("     Sequence: %d\n", input.Sequence)
	}

	fmt.Printf("\nOutputs (%d):\n", len(tx.Outputs))
	for i, output := range tx.Outputs {
		fmt.Printf("  %d. Value: %d satoshis (%.8f BTC)\n", i, output.Value, float64(output.Value)/100000000.0)
		fmt.Printf("     ScriptPubKey: %s\n", output.ScriptPubKey)
	}
}

func handleListAddresses(client *rpc.Client) {
	addresses, err := client.ListAddresses()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(addresses) == 0 {
		fmt.Println("No addresses in wallet")
		return
	}

	fmt.Printf("Wallet Addresses (%d):\n", len(addresses))
	for i, addr := range addresses {
		fmt.Printf("  %d. %s\n", i+1, addr)
	}
}
