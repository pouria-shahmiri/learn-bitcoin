package wallet

import (
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/keys"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/script"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/transaction"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/utxo"
)

// Send creates a signed transaction sending amount to toAddress
func (w *Wallet) Send(toAddress string, amount int64) (*types.Transaction, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 1. Select UTXOs
	selectedUTXOs, totalValue, err := w.selectUTXOs(amount)
	if err != nil {
		return nil, err
	}

	// 2. Create Transaction Builder
	builder := transaction.NewTxBuilder()

	// Add Inputs
	for _, u := range selectedUTXOs {
		builder.AddInput(u.TxHash, u.OutputIndex)
	}

	// Add Recipient Output
	if _, err := builder.AddP2PKHOutput(amount, toAddress); err != nil {
		return nil, err
	}

	// Add Change Output
	change := totalValue - amount
	if change > 0 {
		// Get a change address (use first available key for now)
		var changeAddr string
		for k := range w.keys {
			changeAddr = k
			break
		}
		if _, err := builder.AddP2PKHOutput(change, changeAddr); err != nil {
			return nil, err
		}
	}

	// Build unsigned tx
	tx, err := builder.Build()
	if err != nil {
		return nil, err
	}

	// 3. Sign Inputs
	for i, input := range tx.Inputs {
		u := w.utxos[utxo.NewOutPoint(input.PrevTxHash, input.OutputIndex)]

		// Get private key
		hash, _ := script.ExtractP2PKHAddress(u.Output.PubKeyScript)

		// Try mainnet
		addr, _ := keys.NewAddress(keys.AddressTypeP2PKH, hash)
		privKey, ok := w.keys[addr.String()]
		if !ok {
			// Try testnet
			addrTest, _ := keys.NewAddress(keys.AddressTypeTestnetP2PKH, hash)
			privKey, ok = w.keys[addrTest.String()]
		}

		if !ok {
			return nil, fmt.Errorf("key not found for input %d", i)
		}

		// Sign
		err = transaction.SignInput(tx, i, privKey, u.Output.PubKeyScript, transaction.SigHashAll)
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func (w *Wallet) selectUTXOs(amount int64) ([]*utxo.UTXO, int64, error) {
	var selected []*utxo.UTXO
	var total int64

	for _, u := range w.utxos {
		selected = append(selected, u)
		total += u.Value()
		if total >= amount {
			return selected, total, nil
		}
	}

	return nil, 0, fmt.Errorf("insufficient funds: have %d, need %d", total, amount)
}
