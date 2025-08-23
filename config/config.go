package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 系统配置
type Config struct {
	Network    NetworkConfig    `yaml:"network"`
	Consensus  ConsensusConfig  `yaml:"consensus"`
	Storage    StorageConfig    `yaml:"storage"`
	Execution  ExecutionConfig  `yaml:"execution"`
	Governance GovernanceConfig `yaml:"governance"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	Port           int      `yaml:"port"`
	Host           string   `yaml:"host"`
	MaxPeers       int      `yaml:"max_peers"`
	BootstrapPeers []string `yaml:"bootstrap_peers"`  // DHT bootstrap 节点列表
	PrivateKeyPath string   `yaml:"private_key_path"` // 私钥文件路径
}

// ConsensusConfig 共识配置
type ConsensusConfig struct {
	Algorithm       string `yaml:"algorithm"`
	ValidatorsCount int    `yaml:"validators_count"`
	BlockTime       int    `yaml:"block_time"`
	RoundBlocks     int    `yaml:"round_blocks"`
	MinStake        int64  `yaml:"min_stake"`
	VotingPeriod    int    `yaml:"voting_period"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	DataDir     string `yaml:"data_dir"`
	MaxSize     int64  `yaml:"max_size"`
	CacheSize   int    `yaml:"cache_size"`
	Compression bool   `yaml:"compression"`
}

// ExecutionConfig 执行配置
type ExecutionConfig struct {
	MaxThreads int `yaml:"max_threads"`
	BatchSize  int `yaml:"batch_size"`
	Timeout    int `yaml:"timeout"`
}

// GovernanceConfig 治理配置
type GovernanceConfig struct {
	MinStake     int64 `yaml:"min_stake"`
	VotingPeriod int   `yaml:"voting_period"`
}

// Load 加载配置文件
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Network: NetworkConfig{
			Port:           26656,
			Host:           "0.0.0.0",
			MaxPeers:       50,
			BootstrapPeers: []string{}, // 默认空的 bootstrap 节点列表
			PrivateKeyPath: "",         // 默认空，表示自动生成私钥
		},
		Consensus: ConsensusConfig{
			Algorithm:       "dpos",
			ValidatorsCount: 21,
			BlockTime:       3000,
			RoundBlocks:     21,
			MinStake:        1000000,
			VotingPeriod:    86400,
		},
		Storage: StorageConfig{
			DataDir:     "./data",
			MaxSize:     1024 * 1024 * 1024, // 1GB
			CacheSize:   1000,
			Compression: true,
		},
		Execution: ExecutionConfig{
			MaxThreads: 8,
			BatchSize:  100,
			Timeout:    5000,
		},
		Governance: GovernanceConfig{
			MinStake:     1000000,
			VotingPeriod: 86400,
		},
	}
}
