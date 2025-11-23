package wallet

import (
	"sync"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/keys"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/script"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/utxo"
)

// Wallet manages private keys and tracks UTXOs
type Wallet struct {
	mu    sync.RWMutex
	keys  map[string]*keys.PrivateKey // address -> private key
	utxos map[utxo.OutPoint]*utxo.UTXO
}

// NewWallet creates a new empty wallet
func NewWallet() *Wallet {
	return &Wallet{
		keys:  make(map[string]*keys.PrivateKey),
		utxos: make(map[utxo.OutPoint]*utxo.UTXO),
	}
}

// GenerateAddress creates a new private key and returns its address
func (w *Wallet) GenerateAddress() (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	privKey, err := keys.GeneratePrivateKey()
	if err != nil {
		return "", err
	}

	pubKey := privKey.PublicKey()
	address := pubKey.P2PKHAddress()

	w.keys[address] = privKey
	return address, nil
}

// GetBalance calculates the total balance of the wallet
func (w *Wallet) GetBalance() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var balance int64
	for _, u := range w.utxos {
		balance += u.Value()
	}
	return balance
}

// AddUTXO adds a UTXO to the wallet if it belongs to one of our addresses
func (w *Wallet) AddUTXO(u *utxo.UTXO) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if script is P2PKH
	if !script.IsP2PKH(u.Output.PubKeyScript) {
		return
	}

	// Extract hash
	hash, err := script.ExtractP2PKHAddress(u.Output.PubKeyScript)
	if err != nil {
		return
	}

	// Check against Mainnet P2PKH
	addr, _ := keys.NewAddress(keys.AddressTypeP2PKH, hash)
	if _, ok := w.keys[addr.String()]; ok {
		w.utxos[u.OutPoint()] = u.Clone()
		return
	}

	// Check against Testnet P2PKH
	addrTest, _ := keys.NewAddress(keys.AddressTypeTestnetP2PKH, hash)
	if _, ok := w.keys[addrTest.String()]; ok {
		w.utxos[u.OutPoint()] = u.Clone()
		return
	}
}

// GetAddress returns the private key for a given address
func (w *Wallet) GetKey(address string) (*keys.PrivateKey, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	key, ok := w.keys[address]
	return key, ok
}

// ListAddresses returns all addresses in the wallet
func (w *Wallet) ListAddresses() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	addrs := make([]string, 0, len(w.keys))
	for k := range w.keys {
		addrs = append(addrs, k)
	}
	return addrs
}
