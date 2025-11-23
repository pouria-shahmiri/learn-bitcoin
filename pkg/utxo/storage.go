package utxo

import (
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/storage"
)

// UTXOStorage provides persistent storage for the UTXO set
type UTXOStorage struct {
	db *storage.Database
}

// NewUTXOStorage creates a new UTXO storage
func NewUTXOStorage(dbPath string) (*UTXOStorage, error) {
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		return nil, err
	}

	return &UTXOStorage{db: db}, nil
}

// Close closes the storage
func (us *UTXOStorage) Close() error {
	return us.db.Close()
}

// utxoKey creates a database key for a UTXO
// Format: 'u' + outpoint (32 bytes hash + 4 bytes index)
func utxoKey(outpoint OutPoint) []byte {
	key := make([]byte, 1+36)
	key[0] = 'u' // UTXO prefix
	copy(key[1:], outpoint.Bytes())
	return key
}

// Save saves a UTXO to storage
func (us *UTXOStorage) Save(utxo *UTXO) error {
	key := utxoKey(utxo.OutPoint())
	value := utxo.Serialize()

	return us.db.Put(key, value)
}

// Load loads a UTXO from storage
func (us *UTXOStorage) Load(outpoint OutPoint) (*UTXO, error) {
	key := utxoKey(outpoint)
	value, err := us.db.Get(key)

	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, fmt.Errorf("UTXO not found: %s", outpoint)
	}

	return DeserializeUTXO(value)
}

// Delete removes a UTXO from storage
func (us *UTXOStorage) Delete(outpoint OutPoint) error {
	key := utxoKey(outpoint)
	return us.db.Delete(key)
}

// Exists checks if a UTXO exists in storage
func (us *UTXOStorage) Exists(outpoint OutPoint) (bool, error) {
	key := utxoKey(outpoint)
	return us.db.Has(key)
}

// SaveSet saves an entire UTXO set to storage
func (us *UTXOStorage) SaveSet(set *UTXOSet) error {
	batch := us.db.NewBatch()

	for _, utxo := range set.GetAll() {
		key := utxoKey(utxo.OutPoint())
		value := utxo.Serialize()
		batch.Put(key, value)
	}

	return batch.Write()
}

// LoadAll loads all UTXOs from storage into a UTXO set
func (us *UTXOStorage) LoadAll() (*UTXOSet, error) {
	set := NewUTXOSet()

	// Create iterator for UTXO prefix
	iter := us.db.NewIterator([]byte{'u'})
	defer iter.Release()

	for iter.Next() {
		value := iter.Value()
		utxo, err := DeserializeUTXO(value)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize UTXO: %w", err)
		}

		set.Add(utxo)
	}

	if err := iter.Error(); err != nil {
		return nil, err
	}

	return set, nil
}

// ApplyBatch applies a batch of changes to the UTXO storage
type UTXOChange struct {
	Outpoint OutPoint
	Add      *UTXO // If not nil, add this UTXO
	Remove   bool  // If true, remove this outpoint
}

// ApplyChanges applies a batch of UTXO changes atomically
func (us *UTXOStorage) ApplyChanges(changes []UTXOChange) error {
	batch := us.db.NewBatch()

	for _, change := range changes {
		key := utxoKey(change.Outpoint)

		if change.Remove {
			batch.Delete(key)
		} else if change.Add != nil {
			value := change.Add.Serialize()
			batch.Put(key, value)
		}
	}

	return batch.Write()
}

// Count returns the number of UTXOs in storage
func (us *UTXOStorage) Count() (int, error) {
	count := 0

	iter := us.db.NewIterator([]byte{'u'})
	defer iter.Release()

	for iter.Next() {
		count++
	}

	if err := iter.Error(); err != nil {
		return 0, err
	}

	return count, nil
}
