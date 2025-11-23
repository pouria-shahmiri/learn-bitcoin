package monitoring

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics collects and tracks system metrics
type Metrics struct {
	mu sync.RWMutex

	// Block processing metrics
	blocksProcessed     uint64
	blockProcessingTime time.Duration
	lastBlockTime       time.Time

	// Transaction metrics
	txProcessed      uint64
	txValidationTime time.Duration

	// Peer metrics
	peerCount     int32
	inboundPeers  int32
	outboundPeers int32

	// Network metrics
	bytesReceived    uint64
	bytesSent        uint64
	messagesReceived uint64
	messagesSent     uint64

	// Mempool metrics
	mempoolSize  int32
	mempoolBytes uint64

	// UTXO metrics
	utxoSetSize     uint64
	utxoCacheHits   uint64
	utxoCacheMisses uint64

	// Reorg metrics
	reorgCount     uint64
	lastReorgDepth uint64

	// Performance metrics
	avgBlockTime time.Duration
	avgTxTime    time.Duration
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		lastBlockTime: time.Now(),
	}
}

// Block Processing Metrics

// RecordBlockProcessed records a processed block
func (m *Metrics) RecordBlockProcessed(processingTime time.Duration) {
	atomic.AddUint64(&m.blocksProcessed, 1)

	m.mu.Lock()
	m.blockProcessingTime += processingTime
	m.lastBlockTime = time.Now()

	// Update average
	if m.blocksProcessed > 0 {
		m.avgBlockTime = m.blockProcessingTime / time.Duration(m.blocksProcessed)
	}
	m.mu.Unlock()
}

// GetBlocksProcessed returns total blocks processed
func (m *Metrics) GetBlocksProcessed() uint64 {
	return atomic.LoadUint64(&m.blocksProcessed)
}

// GetAvgBlockProcessingTime returns average block processing time
func (m *Metrics) GetAvgBlockProcessingTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.avgBlockTime
}

// Transaction Metrics

// RecordTxProcessed records a processed transaction
func (m *Metrics) RecordTxProcessed(validationTime time.Duration) {
	atomic.AddUint64(&m.txProcessed, 1)

	m.mu.Lock()
	m.txValidationTime += validationTime

	// Update average
	if m.txProcessed > 0 {
		m.avgTxTime = m.txValidationTime / time.Duration(m.txProcessed)
	}
	m.mu.Unlock()
}

// GetTxProcessed returns total transactions processed
func (m *Metrics) GetTxProcessed() uint64 {
	return atomic.LoadUint64(&m.txProcessed)
}

// GetAvgTxValidationTime returns average transaction validation time
func (m *Metrics) GetAvgTxValidationTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.avgTxTime
}

// Peer Metrics

// SetPeerCount sets the current peer count
func (m *Metrics) SetPeerCount(count int) {
	atomic.StoreInt32(&m.peerCount, int32(count))
}

// IncrementPeerCount increments peer count
func (m *Metrics) IncrementPeerCount(inbound bool) {
	atomic.AddInt32(&m.peerCount, 1)
	if inbound {
		atomic.AddInt32(&m.inboundPeers, 1)
	} else {
		atomic.AddInt32(&m.outboundPeers, 1)
	}
}

// DecrementPeerCount decrements peer count
func (m *Metrics) DecrementPeerCount(inbound bool) {
	atomic.AddInt32(&m.peerCount, -1)
	if inbound {
		atomic.AddInt32(&m.inboundPeers, -1)
	} else {
		atomic.AddInt32(&m.outboundPeers, -1)
	}
}

// GetPeerCount returns current peer count
func (m *Metrics) GetPeerCount() int {
	return int(atomic.LoadInt32(&m.peerCount))
}

// GetInboundPeers returns inbound peer count
func (m *Metrics) GetInboundPeers() int {
	return int(atomic.LoadInt32(&m.inboundPeers))
}

// GetOutboundPeers returns outbound peer count
func (m *Metrics) GetOutboundPeers() int {
	return int(atomic.LoadInt32(&m.outboundPeers))
}

// Network Metrics

// RecordBytesReceived records bytes received
func (m *Metrics) RecordBytesReceived(bytes uint64) {
	atomic.AddUint64(&m.bytesReceived, bytes)
}

// RecordBytesSent records bytes sent
func (m *Metrics) RecordBytesSent(bytes uint64) {
	atomic.AddUint64(&m.bytesSent, bytes)
}

// RecordMessageReceived records a received message
func (m *Metrics) RecordMessageReceived() {
	atomic.AddUint64(&m.messagesReceived, 1)
}

// RecordMessageSent records a sent message
func (m *Metrics) RecordMessageSent() {
	atomic.AddUint64(&m.messagesSent, 1)
}

// GetBytesReceived returns total bytes received
func (m *Metrics) GetBytesReceived() uint64 {
	return atomic.LoadUint64(&m.bytesReceived)
}

// GetBytesSent returns total bytes sent
func (m *Metrics) GetBytesSent() uint64 {
	return atomic.LoadUint64(&m.bytesSent)
}

// Mempool Metrics

// SetMempoolSize sets current mempool size
func (m *Metrics) SetMempoolSize(size int, bytes uint64) {
	atomic.StoreInt32(&m.mempoolSize, int32(size))
	atomic.StoreUint64(&m.mempoolBytes, bytes)
}

// GetMempoolSize returns current mempool size
func (m *Metrics) GetMempoolSize() int {
	return int(atomic.LoadInt32(&m.mempoolSize))
}

// GetMempoolBytes returns current mempool bytes
func (m *Metrics) GetMempoolBytes() uint64 {
	return atomic.LoadUint64(&m.mempoolBytes)
}

// UTXO Metrics

// SetUTXOSetSize sets UTXO set size
func (m *Metrics) SetUTXOSetSize(size uint64) {
	atomic.StoreUint64(&m.utxoSetSize, size)
}

// RecordUTXOCacheHit records a UTXO cache hit
func (m *Metrics) RecordUTXOCacheHit() {
	atomic.AddUint64(&m.utxoCacheHits, 1)
}

// RecordUTXOCacheMiss records a UTXO cache miss
func (m *Metrics) RecordUTXOCacheMiss() {
	atomic.AddUint64(&m.utxoCacheMisses, 1)
}

// GetUTXOSetSize returns UTXO set size
func (m *Metrics) GetUTXOSetSize() uint64 {
	return atomic.LoadUint64(&m.utxoSetSize)
}

// GetUTXOCacheHitRate returns UTXO cache hit rate
func (m *Metrics) GetUTXOCacheHitRate() float64 {
	hits := atomic.LoadUint64(&m.utxoCacheHits)
	misses := atomic.LoadUint64(&m.utxoCacheMisses)
	total := hits + misses

	if total == 0 {
		return 0
	}

	return float64(hits) / float64(total)
}

// Reorg Metrics

// RecordReorg records a chain reorganization
func (m *Metrics) RecordReorg(depth uint64) {
	atomic.AddUint64(&m.reorgCount, 1)
	atomic.StoreUint64(&m.lastReorgDepth, depth)
}

// GetReorgCount returns total reorg count
func (m *Metrics) GetReorgCount() uint64 {
	return atomic.LoadUint64(&m.reorgCount)
}

// GetLastReorgDepth returns depth of last reorg
func (m *Metrics) GetLastReorgDepth() uint64 {
	return atomic.LoadUint64(&m.lastReorgDepth)
}

// Summary returns a metrics summary
func (m *Metrics) Summary() map[string]interface{} {
	return map[string]interface{}{
		"blocks_processed":    m.GetBlocksProcessed(),
		"avg_block_time_ms":   m.GetAvgBlockProcessingTime().Milliseconds(),
		"tx_processed":        m.GetTxProcessed(),
		"avg_tx_time_us":      m.GetAvgTxValidationTime().Microseconds(),
		"peer_count":          m.GetPeerCount(),
		"inbound_peers":       m.GetInboundPeers(),
		"outbound_peers":      m.GetOutboundPeers(),
		"bytes_received":      m.GetBytesReceived(),
		"bytes_sent":          m.GetBytesSent(),
		"mempool_size":        m.GetMempoolSize(),
		"mempool_bytes":       m.GetMempoolBytes(),
		"utxo_set_size":       m.GetUTXOSetSize(),
		"utxo_cache_hit_rate": m.GetUTXOCacheHitRate(),
		"reorg_count":         m.GetReorgCount(),
		"last_reorg_depth":    m.GetLastReorgDepth(),
	}
}

// Global metrics instance
var globalMetrics = NewMetrics()

// GetGlobalMetrics returns the global metrics instance
func GetGlobalMetrics() *Metrics {
	return globalMetrics
}
