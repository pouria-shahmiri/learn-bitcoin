package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// NodeConfig holds all configuration for a Bitcoin node
type NodeConfig struct {
	// Node Identity
	NodeID string

	// Network Configuration
	Network      string   // mainnet, testnet, regtest
	RPCPort      int      // RPC server port
	P2PPort      int      // P2P network port
	InitialPeers []string // List of initial peer addresses

	// Storage
	DataDir string // Data directory path

	// Mining Configuration
	MiningEnabled bool          // Enable mining
	MinerAddress  string        // Address to receive mining rewards
	AutoMine      bool          // Automatically mine blocks
	MineInterval  time.Duration // Interval between auto-mining attempts

	// Logging
	LogLevel string // debug, info, warn, error

	// Monitoring
	EnableMonitoring bool // Enable monitoring/metrics
}

// DefaultConfig returns the default configuration
func DefaultConfig() *NodeConfig {
	return &NodeConfig{
		NodeID:           "bitcoin-node",
		Network:          "regtest",
		RPCPort:          8332,
		P2PPort:          8333,
		DataDir:          "./data/node",
		MiningEnabled:    false,
		MinerAddress:     "",
		AutoMine:         false,
		MineInterval:     10 * time.Second,
		LogLevel:         "info",
		InitialPeers:     []string{},
		EnableMonitoring: false,
	}
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *NodeConfig {
	cfg := DefaultConfig()

	// Node Identity
	if nodeID := os.Getenv("NODE_ID"); nodeID != "" {
		cfg.NodeID = nodeID
	}

	// Network Configuration
	if network := os.Getenv("NETWORK"); network != "" {
		cfg.Network = network
	}

	if rpcPort := os.Getenv("RPC_PORT"); rpcPort != "" {
		if port, err := strconv.Atoi(rpcPort); err == nil {
			cfg.RPCPort = port
		}
	}

	if p2pPort := os.Getenv("P2P_PORT"); p2pPort != "" {
		if port, err := strconv.Atoi(p2pPort); err == nil {
			cfg.P2PPort = port
		}
	}

	if peers := os.Getenv("INITIAL_PEERS"); peers != "" {
		cfg.InitialPeers = strings.Split(peers, ",")
	}

	// Storage
	if dataDir := os.Getenv("DATA_DIR"); dataDir != "" {
		cfg.DataDir = dataDir
	}

	// Mining Configuration
	if miningEnabled := os.Getenv("MINING_ENABLED"); miningEnabled != "" {
		cfg.MiningEnabled = strings.ToLower(miningEnabled) == "true"
	}

	if minerAddr := os.Getenv("MINER_ADDRESS"); minerAddr != "" {
		cfg.MinerAddress = minerAddr
	}

	if autoMine := os.Getenv("AUTO_MINE"); autoMine != "" {
		cfg.AutoMine = strings.ToLower(autoMine) == "true"
	}

	if mineInterval := os.Getenv("MINE_INTERVAL"); mineInterval != "" {
		if interval, err := strconv.Atoi(mineInterval); err == nil {
			cfg.MineInterval = time.Duration(interval) * time.Second
		}
	}

	// Logging
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}

	// Monitoring
	if enableMonitoring := os.Getenv("ENABLE_MONITORING"); enableMonitoring != "" {
		cfg.EnableMonitoring = strings.ToLower(enableMonitoring) == "true"
	}

	return cfg
}

// Validate checks if the configuration is valid
func (c *NodeConfig) Validate() error {
	// Validate network
	validNetworks := map[string]bool{
		"mainnet": true,
		"testnet": true,
		"regtest": true,
	}
	if !validNetworks[c.Network] {
		return fmt.Errorf("invalid network: %s (must be mainnet, testnet, or regtest)", c.Network)
	}

	// Validate ports
	if c.RPCPort < 1 || c.RPCPort > 65535 {
		return fmt.Errorf("invalid RPC port: %d", c.RPCPort)
	}
	if c.P2PPort < 1 || c.P2PPort > 65535 {
		return fmt.Errorf("invalid P2P port: %d", c.P2PPort)
	}

	// Validate data directory
	if c.DataDir == "" {
		return fmt.Errorf("data directory cannot be empty")
	}

	// Validate mining configuration
	if c.MiningEnabled && c.MinerAddress == "" {
		return fmt.Errorf("miner address required when mining is enabled")
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}

	return nil
}

// String returns a string representation of the configuration
func (c *NodeConfig) String() string {
	return fmt.Sprintf(`Bitcoin Node Configuration:
  Node ID:          %s
  Network:          %s
  RPC Port:         %d
  P2P Port:         %d
  Data Directory:   %s
  Mining Enabled:   %v
  Miner Address:    %s
  Auto Mine:        %v
  Mine Interval:    %v
  Log Level:        %s
  Initial Peers:    %v
  Enable Monitoring: %v`,
		c.NodeID,
		c.Network,
		c.RPCPort,
		c.P2PPort,
		c.DataDir,
		c.MiningEnabled,
		c.MinerAddress,
		c.AutoMine,
		c.MineInterval,
		c.LogLevel,
		c.InitialPeers,
		c.EnableMonitoring,
	)
}

// GetRPCAddress returns the full RPC address
func (c *NodeConfig) GetRPCAddress() string {
	return fmt.Sprintf(":%d", c.RPCPort)
}

// GetP2PAddress returns the full P2P address
func (c *NodeConfig) GetP2PAddress() string {
	return fmt.Sprintf(":%d", c.P2PPort)
}
