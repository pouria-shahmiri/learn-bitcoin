package storage

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator" // ADD THIS LINE
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Database wraps LevelDB with Bitcoin-specific operations
type Database struct {
	db *leveldb.DB
}

// OpenDatabase opens or creates a LevelDB database
func OpenDatabase(path string) (*Database, error) {
	// Open with compression enabled
	opts := &opt.Options{
		Compression: opt.SnappyCompression,
	}

	db, err := leveldb.OpenFile(path, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Database{db: db}, nil
}

// Close closes the database
func (db *Database) Close() error {
	return db.db.Close()
}

// Get retrieves value for key
func (db *Database) Get(key []byte) ([]byte, error) {
	value, err := db.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return nil, nil // Return nil for not found (not an error)
	}
	return value, err
}

// Put stores key-value pair
func (db *Database) Put(key, value []byte) error {
	return db.db.Put(key, value, nil)
}

// Delete removes key
func (db *Database) Delete(key []byte) error {
	return db.db.Delete(key, nil)
}

// Has checks if key exists
func (db *Database) Has(key []byte) (bool, error) {
	return db.db.Has(key, nil)
}

// Batch represents an atomic batch of operations
type Batch struct {
	batch *leveldb.Batch
	db    *Database
}

// NewBatch creates a new batch for atomic writes
func (db *Database) NewBatch() *Batch {
	return &Batch{
		batch: new(leveldb.Batch),
		db:    db,
	}
}

// Put adds put operation to batch
func (b *Batch) Put(key, value []byte) {
	b.batch.Put(key, value)
}

// Delete adds delete operation to batch
func (b *Batch) Delete(key []byte) {
	b.batch.Delete(key)
}

// Write commits batch atomically
func (b *Batch) Write() error {
	return b.db.db.Write(b.batch, nil)
}

// Reset clears batch
func (b *Batch) Reset() {
	b.batch.Reset()
}

// Iterator for range queries
type Iterator struct {
	iter iterator.Iterator
}

// NewIterator creates iterator for prefix
func (db *Database) NewIterator(prefix []byte) *Iterator {
	iter := db.db.NewIterator(util.BytesPrefix(prefix), nil)
	return &Iterator{iter: iter}
}

// Next moves to next key
func (it *Iterator) Next() bool {
	return it.iter.Next()
}

// Key returns current key
func (it *Iterator) Key() []byte {
	return it.iter.Key()
}

// Value returns current value
func (it *Iterator) Value() []byte {
	return it.iter.Value()
}

// Release releases iterator resources
func (it *Iterator) Release() {
	it.iter.Release()
}

// Error returns any error encountered
func (it *Iterator) Error() error {
	return it.iter.Error()
}
