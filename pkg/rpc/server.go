package rpc

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/wallet"
)

// Server represents the RPC server
type Server struct {
	wallet     *wallet.Wallet
	blockchain *storage.BlockchainStorage
	addr       string
}

// NewServer creates a new RPC server
func NewServer(w *wallet.Wallet, bc *storage.BlockchainStorage, addr string) *Server {
	return &Server{
		wallet:     w,
		blockchain: bc,
		addr:       addr,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	http.HandleFunc("/getnewaddress", s.handleGetNewAddress)
	http.HandleFunc("/getbalance", s.handleGetBalance)
	http.HandleFunc("/sendtoaddress", s.handleSendToAddress)
	http.HandleFunc("/getblockcount", s.handleGetBlockCount)
	http.HandleFunc("/getblock", s.handleGetBlock)
	http.HandleFunc("/gettransaction", s.handleGetTransaction)
	http.HandleFunc("/listaddresses", s.handleListAddresses)

	log.Printf("RPC server listening on %s", s.addr)
	return http.ListenAndServe(s.addr, nil)
}

// Response structures
type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type NewAddressResponse struct {
	Address string `json:"address"`
}

type BalanceResponse struct {
	Balance int64 `json:"balance"`
}

type SendResponse struct {
	TxHash string `json:"txhash"`
}

type BlockCountResponse struct {
	Height uint64 `json:"height"`
}

type BlockResponse struct {
	Hash         string   `json:"hash"`
	Height       uint64   `json:"height"`
	Version      int32    `json:"version"`
	PrevHash     string   `json:"prev_hash"`
	MerkleRoot   string   `json:"merkle_root"`
	Timestamp    uint32   `json:"timestamp"`
	Bits         uint32   `json:"bits"`
	Nonce        uint32   `json:"nonce"`
	Transactions []string `json:"transactions"`
}

type TransactionResponse struct {
	TxHash   string       `json:"txhash"`
	Version  int32        `json:"version"`
	Inputs   []InputInfo  `json:"inputs"`
	Outputs  []OutputInfo `json:"outputs"`
	LockTime uint32       `json:"locktime"`
}

type InputInfo struct {
	PrevTxHash  string `json:"prev_txhash"`
	OutputIndex uint32 `json:"output_index"`
	ScriptSig   string `json:"script_sig"`
	Sequence    uint32 `json:"sequence"`
}

type OutputInfo struct {
	Value        int64  `json:"value"`
	ScriptPubKey string `json:"script_pubkey"`
}

type ListAddressesResponse struct {
	Addresses []string `json:"addresses"`
}

// Handler functions
func (s *Server) handleGetNewAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		s.sendError(w, "method not allowed")
		return
	}

	address, err := s.wallet.GenerateAddress()
	if err != nil {
		s.sendError(w, fmt.Sprintf("failed to generate address: %v", err))
		return
	}

	s.sendSuccess(w, NewAddressResponse{Address: address})
}

func (s *Server) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, "method not allowed")
		return
	}

	balance := s.wallet.GetBalance()
	s.sendSuccess(w, BalanceResponse{Balance: balance})
}

func (s *Server) handleSendToAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, "method not allowed")
		return
	}

	// Parse request
	var req struct {
		Address string `json:"address"`
		Amount  int64  `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, fmt.Sprintf("invalid request: %v", err))
		return
	}

	// Create and sign transaction
	tx, err := s.wallet.Send(req.Address, req.Amount)
	if err != nil {
		s.sendError(w, fmt.Sprintf("failed to create transaction: %v", err))
		return
	}

	// Get transaction hash
	txHash, err := serialization.HashTransaction(tx)
	if err != nil {
		s.sendError(w, fmt.Sprintf("failed to hash transaction: %v", err))
		return
	}

	s.sendSuccess(w, SendResponse{TxHash: txHash.String()})
}

func (s *Server) handleGetBlockCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, "method not allowed")
		return
	}

	height, err := s.blockchain.GetBestBlockHeight()
	if err != nil {
		s.sendError(w, fmt.Sprintf("failed to get height: %v", err))
		return
	}
	s.sendSuccess(w, BlockCountResponse{Height: height})
}

func (s *Server) handleGetBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, "method not allowed")
		return
	}

	// Get height from query parameter
	heightStr := r.URL.Query().Get("height")
	if heightStr == "" {
		s.sendError(w, "missing height parameter")
		return
	}

	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		s.sendError(w, fmt.Sprintf("invalid height: %v", err))
		return
	}

	// Get block
	block, err := s.blockchain.GetBlockByHeight(height)
	if err != nil {
		s.sendError(w, fmt.Sprintf("block not found: %v", err))
		return
	}

	// Convert to response format
	txHashes := make([]string, len(block.Transactions))
	for i, tx := range block.Transactions {
		txHash, _ := serialization.HashTransaction(&tx)
		txHashes[i] = txHash.String()
	}

	blockHash, _ := s.blockchain.GetBlockHash(block)
	blockResp := BlockResponse{
		Hash:         blockHash.String(),
		Height:       height,
		Version:      block.Header.Version,
		PrevHash:     block.Header.PrevBlockHash.String(),
		MerkleRoot:   block.Header.MerkleRoot.String(),
		Timestamp:    block.Header.Timestamp,
		Bits:         block.Header.Bits,
		Nonce:        block.Header.Nonce,
		Transactions: txHashes,
	}

	s.sendSuccess(w, blockResp)
}

func (s *Server) handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, "method not allowed")
		return
	}

	// Get txhash from query parameter
	txHashStr := r.URL.Query().Get("txhash")
	if txHashStr == "" {
		s.sendError(w, "missing txhash parameter")
		return
	}

	// Parse hash
	txHash, err := types.NewHashFromString(txHashStr)
	if err != nil {
		s.sendError(w, fmt.Sprintf("invalid txhash: %v", err))
		return
	}

	// Get transaction location
	blockHash, txIndex, err := s.blockchain.GetTransactionLocation(txHash)
	if err != nil {
		s.sendError(w, fmt.Sprintf("transaction not found: %v", err))
		return
	}

	// Get block containing transaction
	block, err := s.blockchain.GetBlock(blockHash)
	if err != nil {
		s.sendError(w, fmt.Sprintf("block not found: %v", err))
		return
	}

	// Get transaction from block
	if int(txIndex) >= len(block.Transactions) {
		s.sendError(w, "invalid transaction index")
		return
	}
	tx := &block.Transactions[txIndex]

	// Convert to response format
	inputs := make([]InputInfo, len(tx.Inputs))
	for i, input := range tx.Inputs {
		inputs[i] = InputInfo{
			PrevTxHash:  input.PrevTxHash.String(),
			OutputIndex: input.OutputIndex,
			ScriptSig:   fmt.Sprintf("%x", input.SignatureScript),
			Sequence:    input.Sequence,
		}
	}

	outputs := make([]OutputInfo, len(tx.Outputs))
	for i, output := range tx.Outputs {
		outputs[i] = OutputInfo{
			Value:        output.Value,
			ScriptPubKey: fmt.Sprintf("%x", output.PubKeyScript),
		}
	}

	respTxHash, _ := serialization.HashTransaction(tx)
	txResp := TransactionResponse{
		TxHash:   respTxHash.String(),
		Version:  tx.Version,
		Inputs:   inputs,
		Outputs:  outputs,
		LockTime: tx.LockTime,
	}

	s.sendSuccess(w, txResp)
}

func (s *Server) handleListAddresses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, "method not allowed")
		return
	}

	addresses := s.wallet.ListAddresses()
	s.sendSuccess(w, ListAddressesResponse{Addresses: addresses})
}

// Helper functions
func (s *Server) sendSuccess(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Result: result})
}

func (s *Server) sendError(w http.ResponseWriter, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(Response{Error: errMsg})
}
