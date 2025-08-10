package main

import (
	"fmt"
	"log"
	"time"

	"github.com/govm-net/chain/config"
	"github.com/govm-net/chain/consensus"
	"github.com/govm-net/chain/execution"
	"github.com/govm-net/chain/network"
	"github.com/govm-net/chain/storage"
)

func main() {
	// 创建包含可信节点的配置
	cfg := config.NetworkConfig{
		Port:           26656,
		Host:           "127.0.0.1",
		MaxPeers:       10,
		PrivateKeyPath: "./private_key.pem", // 私钥文件路径（网络模块会自动处理生成/加载）
		BootstrapPeers: []string{
			// 示例 bootstrap 节点地址（需要替换为实际的 bootstrap 节点地址）
			"/ip4/192.168.1.100/tcp/26656/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
			"/ip4/192.168.1.101/tcp/26656/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
		},
	}

	// 创建存储实例
	storage, err := storage.New(config.StorageConfig{
		DataDir:     "./data",
		MaxSize:     1024 * 1024 * 1024, // 1GB
		CacheSize:   1000,
		Compression: true,
	})
	if err != nil {
		log.Fatalf("创建存储失败: %v", err)
	}
	defer storage.Stop()

	// 创建执行引擎
	exec, err := execution.New(config.ExecutionConfig{
		MaxThreads: 8,
		BatchSize:  100,
		Timeout:    5000,
	}, storage)
	if err != nil {
		log.Fatalf("创建执行引擎失败: %v", err)
	}

	// 创建共识实例
	consensus, err := consensus.New(config.ConsensusConfig{
		Algorithm: "pbft",
		MaxFaulty: 1,
		BlockTime: 1000,
		BatchSize: 1000,
	}, exec, storage)
	if err != nil {
		log.Fatalf("创建共识失败: %v", err)
	}

	// 创建网络实例
	network, err := network.New(cfg, consensus)
	if err != nil {
		log.Fatalf("创建网络失败: %v", err)
	}
	defer network.Stop()

	// 启动网络
	if err := network.Start(); err != nil {
		log.Fatalf("启动网络失败: %v", err)
	}

	fmt.Println("网络已启动，正在连接 bootstrap 节点...")

	// 等待一段时间让连接建立
	time.Sleep(2 * time.Second)

	// 显示配置的 bootstrap 节点信息
	configBootstrapPeers := len(cfg.BootstrapPeers)
	fmt.Printf("配置的 bootstrap 节点数量: %d\n", configBootstrapPeers)
	for i, peerAddr := range cfg.BootstrapPeers {
		fmt.Printf("Bootstrap 节点 %d: %s\n", i+1, peerAddr)
	}

	// 显示连接统计信息
	connStats := network.GetConnectionStats()
	fmt.Printf("\n=== 连接统计信息 ===\n")
	fmt.Printf("当前连接节点数: %v\n", connStats["current_peers"])
	fmt.Printf("最大节点数: %v\n", connStats["max_peers"])
	fmt.Printf("连接使用率: %.2f%%\n", connStats["usage_percentage"])
	fmt.Printf("Bootstrap 节点数: %v\n", connStats["bootstrap_peers"])

	// 显示连接的节点
	peers := network.GetPeers()
	fmt.Printf("\n当前连接的节点数量: %d\n", len(peers))
	for i, peerID := range peers {
		status := "已连接"
		if !network.IsPeerConnected(peerID) {
			status = "未连接"
		}
		fmt.Printf("节点 %d: %s (%s)\n", i+1, peerID.String(), status)
	}

	// 示例：Bootstrap 节点由 DHT 自动管理
	fmt.Println("\n示例：Bootstrap 节点由 DHT 自动管理...")
	fmt.Println("Bootstrap 节点现在由 DHT 自动管理，无需手动添加或检查")
	fmt.Println("DHT 会自动处理与 bootstrap 节点的连接和路由")

	fmt.Println("\n网络运行中，按 Ctrl+C 退出...")

	// 保持运行
	select {}
}
