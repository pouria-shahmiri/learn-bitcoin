package security

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu         sync.Mutex
	rate       int     // tokens per second
	burst      int     // max tokens
	tokens     float64 // current tokens
	lastUpdate time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate, burst int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

// Allow checks if an action is allowed under rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()

	// Add tokens based on elapsed time
	rl.tokens += elapsed * float64(rl.rate)
	if rl.tokens > float64(rl.burst) {
		rl.tokens = float64(rl.burst)
	}

	rl.lastUpdate = now

	// Check if we have tokens
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}

// AllowN checks if N actions are allowed
func (rl *RateLimiter) AllowN(n int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()

	// Add tokens based on elapsed time
	rl.tokens += elapsed * float64(rl.rate)
	if rl.tokens > float64(rl.burst) {
		rl.tokens = float64(rl.burst)
	}

	rl.lastUpdate = now

	// Check if we have enough tokens
	if rl.tokens >= float64(n) {
		rl.tokens -= float64(n)
		return true
	}

	return false
}

// ConnectionRateLimiter limits connections per IP
type ConnectionRateLimiter struct {
	mu              sync.RWMutex
	limiters        map[string]*RateLimiter
	maxConnections  int
	connectionRate  int
	burstSize       int
	cleanupInterval time.Duration
	lastCleanup     time.Time
}

// NewConnectionRateLimiter creates a connection rate limiter
func NewConnectionRateLimiter(maxConnections, connectionRate, burstSize int) *ConnectionRateLimiter {
	crl := &ConnectionRateLimiter{
		limiters:        make(map[string]*RateLimiter),
		maxConnections:  maxConnections,
		connectionRate:  connectionRate,
		burstSize:       burstSize,
		cleanupInterval: 5 * time.Minute,
		lastCleanup:     time.Now(),
	}

	return crl
}

// AllowConnection checks if a connection from an IP is allowed
func (crl *ConnectionRateLimiter) AllowConnection(ip string) bool {
	crl.mu.Lock()
	defer crl.mu.Unlock()

	// Periodic cleanup
	if time.Since(crl.lastCleanup) > crl.cleanupInterval {
		crl.cleanup()
	}

	// Get or create limiter for this IP
	limiter, exists := crl.limiters[ip]
	if !exists {
		limiter = NewRateLimiter(crl.connectionRate, crl.burstSize)
		crl.limiters[ip] = limiter
	}

	return limiter.Allow()
}

// cleanup removes old limiters
func (crl *ConnectionRateLimiter) cleanup() {
	// Remove limiters that haven't been used recently
	now := time.Now()
	for ip, limiter := range crl.limiters {
		if now.Sub(limiter.lastUpdate) > crl.cleanupInterval {
			delete(crl.limiters, ip)
		}
	}
	crl.lastCleanup = now
}

// BandwidthLimiter limits bandwidth per connection
type BandwidthLimiter struct {
	mu             sync.RWMutex
	limiters       map[string]*RateLimiter
	bytesPerSecond int
	burstBytes     int
}

// NewBandwidthLimiter creates a bandwidth limiter
func NewBandwidthLimiter(bytesPerSecond, burstBytes int) *BandwidthLimiter {
	return &BandwidthLimiter{
		limiters:       make(map[string]*RateLimiter),
		bytesPerSecond: bytesPerSecond,
		burstBytes:     burstBytes,
	}
}

// AllowBytes checks if sending/receiving N bytes is allowed
func (bl *BandwidthLimiter) AllowBytes(peer string, bytes int) bool {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	limiter, exists := bl.limiters[peer]
	if !exists {
		limiter = NewRateLimiter(bl.bytesPerSecond, bl.burstBytes)
		bl.limiters[peer] = limiter
	}

	return limiter.AllowN(bytes)
}

// RemovePeer removes a peer's bandwidth limiter
func (bl *BandwidthLimiter) RemovePeer(peer string) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	delete(bl.limiters, peer)
}

// DoSProtection provides DoS protection mechanisms
type DoSProtection struct {
	mu                sync.RWMutex
	connectionLimiter *ConnectionRateLimiter
	bandwidthLimiter  *BandwidthLimiter
	bannedIPs         map[string]time.Time
	banDuration       time.Duration
	maxBanScore       int
	banScores         map[string]int
}

// NewDoSProtection creates DoS protection
func NewDoSProtection() *DoSProtection {
	return &DoSProtection{
		connectionLimiter: NewConnectionRateLimiter(100, 10, 20),
		bandwidthLimiter:  NewBandwidthLimiter(1024*1024, 10*1024*1024), // 1MB/s, 10MB burst
		bannedIPs:         make(map[string]time.Time),
		banDuration:       24 * time.Hour,
		maxBanScore:       100,
		banScores:         make(map[string]int),
	}
}

// AllowConnection checks if a connection is allowed
func (dp *DoSProtection) AllowConnection(addr net.Addr) error {
	ip := extractIP(addr)

	// Check if IP is banned
	if dp.isBanned(ip) {
		return fmt.Errorf("IP is banned: %s", ip)
	}

	// Check rate limit
	if !dp.connectionLimiter.AllowConnection(ip) {
		dp.increaseBanScore(ip, 10)
		return fmt.Errorf("connection rate limit exceeded: %s", ip)
	}

	return nil
}

// AllowBytes checks if bandwidth usage is allowed
func (dp *DoSProtection) AllowBytes(peer string, bytes int) error {
	if !dp.bandwidthLimiter.AllowBytes(peer, bytes) {
		dp.increaseBanScore(extractIPFromPeer(peer), 5)
		return fmt.Errorf("bandwidth limit exceeded: %s", peer)
	}
	return nil
}

// isBanned checks if an IP is banned
func (dp *DoSProtection) isBanned(ip string) bool {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	banTime, exists := dp.bannedIPs[ip]
	if !exists {
		return false
	}

	// Check if ban has expired
	if time.Since(banTime) > dp.banDuration {
		delete(dp.bannedIPs, ip)
		return false
	}

	return true
}

// increaseBanScore increases ban score for an IP
func (dp *DoSProtection) increaseBanScore(ip string, points int) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	dp.banScores[ip] += points

	// Ban if score exceeds threshold
	if dp.banScores[ip] >= dp.maxBanScore {
		dp.bannedIPs[ip] = time.Now()
		delete(dp.banScores, ip)
	}
}

// BanIP manually bans an IP
func (dp *DoSProtection) BanIP(ip string) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.bannedIPs[ip] = time.Now()
}

// UnbanIP unbans an IP
func (dp *DoSProtection) UnbanIP(ip string) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	delete(dp.bannedIPs, ip)
	delete(dp.banScores, ip)
}

// GetBannedIPs returns list of banned IPs
func (dp *DoSProtection) GetBannedIPs() []string {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	var ips []string
	for ip := range dp.bannedIPs {
		ips = append(ips, ip)
	}
	return ips
}

// extractIP extracts IP from net.Addr
func extractIP(addr net.Addr) string {
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		return tcpAddr.IP.String()
	}
	return addr.String()
}

// extractIPFromPeer extracts IP from peer string
func extractIPFromPeer(peer string) string {
	host, _, err := net.SplitHostPort(peer)
	if err != nil {
		return peer
	}
	return host
}
