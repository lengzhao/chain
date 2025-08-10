package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 系统配置
type Config struct {
	Network   NetworkConfig   `yaml:"network"`
	Consensus ConsensusConfig `yaml:"consensus"`
	Storage   StorageConfig   `yaml:"storage"`
	Execution ExecutionConfig `yaml:"execution"`
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
	Algorithm string `yaml:"algorithm"`
	MaxFaulty int    `yaml:"max_faulty"`
	BlockTime int    `yaml:"block_time"`
	BatchSize int    `yaml:"batch_size"`
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

// Load 加载配置文件
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
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
			Algorithm: "pbft",
			MaxFaulty: 1,
			BlockTime: 1000,
			BatchSize: 1000,
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
	}
}
