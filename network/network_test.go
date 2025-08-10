package network

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/govm-net/chain/config"
	"github.com/govm-net/chain/consensus"
	"github.com/govm-net/chain/execution"
	"github.com/govm-net/chain/storage"
	"github.com/govm-net/chain/types"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// 测试辅助函数：创建测试网络实例
func createTestNetwork(t testing.TB, port int) (*Network, func()) {
	return createTestNetworkWithConfig(t, port, false)
}

func createTestNetworkWithConfig(t testing.TB, port int, withTrustedPeers bool) (*Network, func()) {
	// 创建测试配置
	cfg := config.NetworkConfig{
		Port:     port,
		Host:     "127.0.0.1",
		MaxPeers: 10,
	}

	// 如果需要可信节点，添加配置
	if withTrustedPeers {
		cfg.BootstrapPeers = []string{
			"/ip4/192.168.1.100/tcp/26656/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
		}
	}

	// 创建存储实例
	storage, err := storage.New(config.StorageConfig{
		DataDir:     "./test_data",
		MaxSize:     1024 * 1024,
		CacheSize:   100,
		Compression: false,
	})
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}

	// 创建执行引擎
	exec, err := execution.New(config.ExecutionConfig{
		MaxThreads: 4,
		BatchSize:  100,
		Timeout:    5000,
	}, storage)
	if err != nil {
		t.Fatalf("创建执行引擎失败: %v", err)
	}

	// 创建共识实例
	consensus, err := consensus.New(config.ConsensusConfig{
		Algorithm: "pbft",
		MaxFaulty: 1,
		BlockTime: 1000,
		BatchSize: 100,
	}, exec, storage)
	if err != nil {
		t.Fatalf("创建共识失败: %v", err)
	}

	// 创建网络实例
	network, err := New(cfg, consensus)
	if err != nil {
		t.Fatalf("创建网络失败: %v", err)
	}

	// 返回清理函数
	cleanup := func() {
		network.Stop()
		storage.Stop()
	}

	return network, cleanup
}

func TestNetworkCreation(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26656)
	defer cleanup()

	if network == nil {
		t.Fatal("网络实例为空")
	}

	// 测试获取主机
	host := network.GetHost()
	if host == nil {
		t.Fatal("主机实例为空")
	}

	// 测试启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}

	// 等待一段时间确保启动完成
	time.Sleep(100 * time.Millisecond)
}

func TestNetworkStartStop(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26657)
	defer cleanup()

	// 测试启动
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}

	// 等待启动完成
	time.Sleep(100 * time.Millisecond)

	// 测试停止
	network.Stop()

	// 验证停止后状态
	peers := network.GetPeers()
	if len(peers) != 0 {
		t.Errorf("停止后应该没有连接的节点，但找到了 %d 个", len(peers))
	}
}

func TestGetPeers(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26658)
	defer cleanup()

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}
	defer network.Stop()

	// 获取节点列表
	peers := network.GetPeers()
	if peers == nil {
		t.Fatal("节点列表不应该为nil")
	}

	// 初始状态下应该没有连接的节点
	if len(peers) != 0 {
		t.Errorf("初始状态下应该没有连接的节点，但找到了 %d 个", len(peers))
	}
}

func TestRegisterMessageHandler(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26659)
	defer cleanup()

	// 测试消息处理器
	messageReceived := make(chan bool, 1)
	handler := func(peerID peer.ID, data []byte) error {
		messageReceived <- true
		return nil
	}

	// 注册消息处理器
	network.RegisterMessageHandler("test", handler)

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}
	defer network.Stop()

	// 等待一段时间确保处理器注册完成
	time.Sleep(100 * time.Millisecond)

	// 验证处理器已注册（通过内部状态检查）
	// 注意：这里我们无法直接访问内部状态，但可以通过功能测试验证
}

func TestBroadcastMessage(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26660)
	defer cleanup()

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}
	defer network.Stop()

	// 测试广播消息（没有连接的节点时）
	testData := []byte("test message")
	err = network.BroadcastMessage("test", testData)
	if err != nil {
		t.Fatalf("广播消息失败: %v", err)
	}

	// 等待一段时间
	time.Sleep(100 * time.Millisecond)
}

func TestConnectToPeer(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26661)
	defer cleanup()

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}
	defer network.Stop()

	// 测试连接到无效地址
	invalidAddr := "/ip4/127.0.0.1/tcp/99999/p2p/invalid-peer-id"
	err = network.ConnectToPeer(invalidAddr)
	if err == nil {
		t.Error("连接到无效地址应该失败")
	}

	// 测试连接到有效格式但不存在的主机
	validFormatAddr := "/ip4/127.0.0.1/tcp/9999/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N"
	err = network.ConnectToPeer(validFormatAddr)
	// 这个可能会失败，但应该不会panic
	if err != nil {
		t.Logf("连接到不存在的主机失败（预期）: %v", err)
	}
}

func TestMessageSerialization(t *testing.T) {
	// 测试消息序列化和反序列化
	originalMessage := &types.Message{
		Type: "test",
		Data: []byte("test data"),
		From: "peer1",
		To:   "peer2",
		Time: time.Now(),
	}

	// 序列化
	serialized, err := originalMessage.Serialize()
	if err != nil {
		t.Fatalf("序列化消息失败: %v", err)
	}

	// 反序列化
	deserialized, err := types.DeserializeMessage(serialized)
	if err != nil {
		t.Fatalf("反序列化消息失败: %v", err)
	}

	// 验证字段
	if deserialized.Type != originalMessage.Type {
		t.Errorf("消息类型不匹配: 期望 %s, 得到 %s", originalMessage.Type, deserialized.Type)
	}

	if string(deserialized.Data) != string(originalMessage.Data) {
		t.Errorf("消息数据不匹配: 期望 %s, 得到 %s", string(originalMessage.Data), string(deserialized.Data))
	}

	if deserialized.From != originalMessage.From {
		t.Errorf("消息来源不匹配: 期望 %s, 得到 %s", originalMessage.From, deserialized.From)
	}

	if deserialized.To != originalMessage.To {
		t.Errorf("消息目标不匹配: 期望 %s, 得到 %s", originalMessage.To, deserialized.To)
	}
}

func TestNetworkConfigValidation(t *testing.T) {
	// 测试无效配置
	invalidConfigs := []config.NetworkConfig{
		{
			Port:     -1, // 无效端口
			Host:     "127.0.0.1",
			MaxPeers: 10,
		},
		{
			Port:     26656,
			Host:     "", // 无效主机
			MaxPeers: 10,
		},
	}

	for i, cfg := range invalidConfigs {
		// 创建存储和执行引擎（使用默认配置）
		storage, err := storage.New(config.StorageConfig{
			DataDir:     "./test_data",
			MaxSize:     1024 * 1024,
			CacheSize:   100,
			Compression: false,
		})
		if err != nil {
			t.Fatalf("创建存储失败: %v", err)
		}

		exec, err := execution.New(config.ExecutionConfig{
			MaxThreads: 4,
			BatchSize:  100,
			Timeout:    5000,
		}, storage)
		if err != nil {
			t.Fatalf("创建执行引擎失败: %v", err)
		}

		consensus, err := consensus.New(config.ConsensusConfig{
			Algorithm: "pbft",
			MaxFaulty: 1,
			BlockTime: 1000,
			BatchSize: 100,
		}, exec, storage)
		if err != nil {
			t.Fatalf("创建共识失败: %v", err)
		}

		// 尝试创建网络实例
		_, err = New(cfg, consensus)
		if err == nil {
			t.Errorf("配置 %d 应该失败，但没有失败", i)
		}

		storage.Stop()
	}
}

func TestNetworkConcurrency(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26662)
	defer cleanup()

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}
	defer network.Stop()

	// 并发测试：多个goroutine同时访问网络
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// 并发获取节点列表
			peers := network.GetPeers()
			if peers == nil {
				t.Errorf("goroutine %d: 节点列表为nil", id)
				return
			}

			// 并发注册消息处理器
			handler := func(peerID peer.ID, data []byte) error {
				return nil
			}
			network.RegisterMessageHandler("test", handler)

			// 并发广播消息
			testData := []byte("concurrent test message")
			err := network.BroadcastMessage("test", testData)
			if err != nil {
				t.Errorf("goroutine %d: 广播消息失败: %v", id, err)
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestNetworkCleanup(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26663)
	defer cleanup()

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}

	// 等待一段时间
	time.Sleep(100 * time.Millisecond)

	// 停止网络
	network.Stop()

	// 验证清理完成
	peers := network.GetPeers()
	if len(peers) != 0 {
		t.Errorf("清理后应该没有连接的节点，但找到了 %d 个", len(peers))
	}
}

// 基准测试
func BenchmarkNetworkCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		network, cleanup := createTestNetwork(b, 26664)
		_ = network // 使用network变量以避免未使用错误
		cleanup()
	}
}

func BenchmarkBroadcastMessage(b *testing.B) {
	network, cleanup := createTestNetwork(b, 26665)
	defer cleanup()

	err := network.Start()
	if err != nil {
		b.Fatalf("启动网络失败: %v", err)
	}

	testData := []byte("benchmark test message")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := network.BroadcastMessage("test", testData)
		if err != nil {
			b.Fatalf("广播消息失败: %v", err)
		}
	}
}

func BenchmarkGetPeers(b *testing.B) {
	network, cleanup := createTestNetwork(b, 26666)
	defer cleanup()

	err := network.Start()
	if err != nil {
		b.Fatalf("启动网络失败: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		peers := network.GetPeers()
		if peers == nil {
			b.Fatal("节点列表为nil")
		}
	}
}

// 测试错误处理
func TestNetworkErrorHandling(t *testing.T) {
	// 测试无效配置
	invalidConfig := config.NetworkConfig{
		Port:     -1, // 无效端口
		Host:     "127.0.0.1",
		MaxPeers: 10,
	}

	// 创建存储和执行引擎
	storage, err := storage.New(config.StorageConfig{
		DataDir:     "./test_data",
		MaxSize:     1024 * 1024,
		CacheSize:   100,
		Compression: false,
	})
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer storage.Stop()

	exec, err := execution.New(config.ExecutionConfig{
		MaxThreads: 4,
		BatchSize:  100,
		Timeout:    5000,
	}, storage)
	if err != nil {
		t.Fatalf("创建执行引擎失败: %v", err)
	}

	consensus, err := consensus.New(config.ConsensusConfig{
		Algorithm: "pbft",
		MaxFaulty: 1,
		BlockTime: 1000,
		BatchSize: 100,
	}, exec, storage)
	if err != nil {
		t.Fatalf("创建共识失败: %v", err)
	}

	// 尝试创建网络实例（应该失败）
	_, err = New(invalidConfig, consensus)
	if err == nil {
		t.Error("使用无效配置创建网络应该失败")
	}
}

// 测试消息处理
func TestMessageProcessing(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26667)
	defer cleanup()

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}
	defer network.Stop()

	// 注册消息处理器
	messageReceived := make(chan []byte, 1)
	handler := func(peerID peer.ID, data []byte) error {
		messageReceived <- data
		return nil
	}
	network.RegisterMessageHandler("test", handler)

	// 等待处理器注册完成
	time.Sleep(100 * time.Millisecond)

	// 测试消息序列化和反序列化
	testMessage := &types.Message{
		Type: "test",
		Data: []byte("test data"),
		From: "peer1",
		To:   "peer2",
		Time: time.Now(),
	}

	// 序列化消息
	serialized, err := testMessage.Serialize()
	if err != nil {
		t.Fatalf("序列化消息失败: %v", err)
	}

	// 反序列化消息
	deserialized, err := types.DeserializeMessage(serialized)
	if err != nil {
		t.Fatalf("反序列化消息失败: %v", err)
	}

	// 验证消息内容
	if deserialized.Type != testMessage.Type {
		t.Errorf("消息类型不匹配: 期望 %s, 得到 %s", testMessage.Type, deserialized.Type)
	}

	if string(deserialized.Data) != string(testMessage.Data) {
		t.Errorf("消息数据不匹配: 期望 %s, 得到 %s", string(testMessage.Data), string(deserialized.Data))
	}
}

// 测试连接管理
func TestConnectionManagement(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26668)
	defer cleanup()

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}
	defer network.Stop()

	// 测试连接管理
	peers := network.GetPeers()
	if len(peers) != 0 {
		t.Errorf("初始状态下应该没有连接的节点，但找到了 %d 个", len(peers))
	}

	// 等待连接管理启动
	time.Sleep(100 * time.Millisecond)
}

// 测试网络接口实现
func TestNetworkInterface(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26669)
	defer cleanup()

	// 测试network.Notifiee接口实现
	// 这些方法应该不会panic
	network.Listen(nil, nil)
	network.ListenClose(nil, nil)
	network.OpenedStream(nil, nil)
	network.ClosedStream(nil, nil)

	// 测试mdns.Notifee接口实现
	// 这个方法应该不会panic
	network.HandlePeerFound(peer.AddrInfo{})
}

// 测试网络配置边界值
func TestNetworkConfigBoundaries(t *testing.T) {
	testCases := []struct {
		name       string
		config     config.NetworkConfig
		shouldFail bool
	}{
		{
			name: "有效配置",
			config: config.NetworkConfig{
				Port:     26656,
				Host:     "127.0.0.1",
				MaxPeers: 10,
			},
			shouldFail: false,
		},
		{
			name: "无效主机",
			config: config.NetworkConfig{
				Port:     26656,
				Host:     "",
				MaxPeers: 10,
			},
			shouldFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建存储和执行引擎
			storage, err := storage.New(config.StorageConfig{
				DataDir:     "./test_data",
				MaxSize:     1024 * 1024,
				CacheSize:   100,
				Compression: false,
			})
			if err != nil {
				t.Fatalf("创建存储失败: %v", err)
			}
			defer storage.Stop()

			exec, err := execution.New(config.ExecutionConfig{
				MaxThreads: 4,
				BatchSize:  100,
				Timeout:    5000,
			}, storage)
			if err != nil {
				t.Fatalf("创建执行引擎失败: %v", err)
			}

			consensus, err := consensus.New(config.ConsensusConfig{
				Algorithm: "pbft",
				MaxFaulty: 1,
				BlockTime: 1000,
				BatchSize: 100,
			}, exec, storage)
			if err != nil {
				t.Fatalf("创建共识失败: %v", err)
			}

			// 尝试创建网络实例
			_, err = New(tc.config, consensus)
			if tc.shouldFail && err == nil {
				t.Error("应该失败但没有失败")
			} else if !tc.shouldFail && err != nil {
				t.Errorf("不应该失败但失败了: %v", err)
			}
		})
	}
}

// 测试网络性能
func TestNetworkPerformance(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26670)
	defer cleanup()

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}
	defer network.Stop()

	// 测试大量消息处理
	const numMessages = 1000
	start := time.Now()

	for i := 0; i < numMessages; i++ {
		testData := []byte(fmt.Sprintf("test message %d", i))
		err := network.BroadcastMessage("test", testData)
		if err != nil {
			t.Fatalf("广播消息失败: %v", err)
		}
	}

	duration := time.Since(start)
	t.Logf("处理 %d 条消息耗时: %v", numMessages, duration)
	t.Logf("平均每条消息处理时间: %v", duration/time.Duration(numMessages))
}

// 测试网络稳定性
func TestNetworkStability(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26671)
	defer cleanup()

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}
	defer network.Stop()

	// 模拟长时间运行
	const testDuration = 2 * time.Second
	start := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for time.Since(start) < testDuration {
		<-ticker.C
		// 定期检查网络状态
		peers := network.GetPeers()
		if peers == nil {
			t.Fatal("节点列表为nil")
		}

		// 发送测试消息
		testData := []byte("stability test message")
		err := network.BroadcastMessage("test", testData)
		if err != nil {
			t.Fatalf("广播消息失败: %v", err)
		}
	}
}

// 测试可信节点功能
func TestTrustedPeers(t *testing.T) {
	// 创建包含可信节点的配置
	cfg := config.NetworkConfig{
		Port:     26672,
		Host:     "127.0.0.1",
		MaxPeers: 10,
		BootstrapPeers: []string{
			"/ip4/127.0.0.1/tcp/26673/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
		},
	}

	// 创建存储实例
	storage, err := storage.New(config.StorageConfig{
		DataDir:     "./test_data",
		MaxSize:     1024 * 1024,
		CacheSize:   100,
		Compression: false,
	})
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer storage.Stop()

	// 创建执行引擎
	exec, err := execution.New(config.ExecutionConfig{
		MaxThreads: 4,
		BatchSize:  100,
		Timeout:    5000,
	}, storage)
	if err != nil {
		t.Fatalf("创建执行引擎失败: %v", err)
	}

	// 创建共识实例
	consensus, err := consensus.New(config.ConsensusConfig{
		Algorithm: "pbft",
		MaxFaulty: 1,
		BlockTime: 1000,
		BatchSize: 100,
	}, exec, storage)
	if err != nil {
		t.Fatalf("创建共识失败: %v", err)
	}

	// 创建网络实例
	network, err := New(cfg, consensus)
	if err != nil {
		t.Fatalf("创建网络失败: %v", err)
	}
	defer network.Stop()

	// 启动网络
	err = network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}

	// 测试配置的可信节点数量
	configBootstrapPeers := len(network.config.BootstrapPeers)
	if configBootstrapPeers != 1 {
		t.Errorf("期望1个配置的 bootstrap 节点，但得到了 %d 个", configBootstrapPeers)
	}

	// 测试连接统计信息中的 bootstrap 节点数量
	stats := network.GetConnectionStats()
	bootstrapPeersCount := stats["bootstrap_peers"].(int)
	if bootstrapPeersCount != 1 {
		t.Errorf("期望1个 bootstrap 节点，但得到了 %d 个", bootstrapPeersCount)
	}

	// 验证 bootstrap 节点配置正确
	expectedPeerAddr := "/ip4/127.0.0.1/tcp/26673/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N"
	found := false
	for _, addr := range network.config.BootstrapPeers {
		if addr == expectedPeerAddr {
			found = true
			break
		}
	}
	if !found {
		t.Error("配置中应该包含预期的可信节点地址")
	}
}

// 测试可信节点连接
func TestTrustedPeerConnection(t *testing.T) {
	network, cleanup := createTestNetworkWithConfig(t, 26673, true)
	defer cleanup()

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}

	// 等待一段时间让连接尝试完成
	time.Sleep(100 * time.Millisecond)

	// 验证配置的 bootstrap 节点数量
	configBootstrapPeers := len(network.config.BootstrapPeers)
	if configBootstrapPeers != 1 {
		t.Errorf("期望1个配置的 bootstrap 节点，但得到了 %d 个", configBootstrapPeers)
	}
}

// 测试可信节点发现优先级
func TestTrustedPeerDiscoveryPriority(t *testing.T) {
	network, cleanup := createTestNetworkWithConfig(t, 26674, true)
	defer cleanup()

	// 测试HandlePeerFound方法对可信节点的处理
	peerID, err := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	if err != nil {
		t.Fatalf("解析节点ID失败: %v", err)
	}

	// 创建模拟的AddrInfo
	addrInfo := peer.AddrInfo{
		ID:    peerID,
		Addrs: []multiaddr.Multiaddr{},
	}

	// 测试HandlePeerFound方法（这应该不会panic）
	network.HandlePeerFound(addrInfo)

	// 验证配置的 bootstrap 节点数量
	configBootstrapPeers := len(network.config.BootstrapPeers)
	if configBootstrapPeers != 1 {
		t.Errorf("期望1个配置的 bootstrap 节点，但得到了 %d 个", configBootstrapPeers)
	}
}

// 测试Gossipsub功能
func TestGossipsubFunctionality(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26678)
	defer cleanup()

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}

	// 等待网络启动
	time.Sleep(1 * time.Second)

	// 测试获取Gossipsub实例
	pubsub := network.GetPubsub()
	if pubsub == nil {
		t.Fatal("Gossipsub实例应该不为nil")
	}

	// 测试列出主题
	topics := network.ListTopics()
	if topics == nil {
		t.Log("主题列表为空（预期）")
	}
}

// 测试Gossipsub消息广播
func TestGossipsubBroadcast(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26679)
	defer cleanup()

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}

	// 等待网络启动
	time.Sleep(1 * time.Second)

	// 测试消息广播
	testData := []byte("test gossipsub message")
	err = network.BroadcastMessage("test-topic", testData)
	if err != nil {
		t.Fatalf("广播消息失败: %v", err)
	}

	t.Log("Gossipsub消息广播成功")
}

// 测试Gossipsub订阅
func TestGossipsubSubscription(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26680)
	defer cleanup()

	// 启动网络
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}

	// 等待网络启动
	time.Sleep(1 * time.Second)

	// 注册消息处理器
	messageReceived := make(chan []byte, 1)
	handler := func(peerID peer.ID, data []byte) error {
		messageReceived <- data
		return nil
	}
	network.RegisterMessageHandler("test-topic", handler)

	// 订阅主题
	err = network.SubscribeToTopic("test-topic")
	if err != nil {
		t.Fatalf("订阅主题失败: %v", err)
	}

	// 等待订阅建立
	time.Sleep(100 * time.Millisecond)

	// 测试获取主题节点
	peers := network.GetTopicPeers("test-topic")
	t.Logf("主题中的节点数量: %d", len(peers))

	t.Log("Gossipsub订阅成功")
}

// TestMDNSFunctionality 测试mDNS功能
func TestMDNSFunctionality(t *testing.T) {
	// 创建启用mDNS的网络实例
	cfg := config.NetworkConfig{
		Port:     26658,
		Host:     "127.0.0.1",
		MaxPeers: 10,
	}

	// 创建存储实例
	storage, err := storage.New(config.StorageConfig{
		DataDir:     "./test_data",
		MaxSize:     1024 * 1024,
		CacheSize:   100,
		Compression: false,
	})
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer storage.Stop()

	// 创建执行引擎
	exec, err := execution.New(config.ExecutionConfig{
		MaxThreads: 4,
		BatchSize:  100,
		Timeout:    5000,
	}, storage)
	if err != nil {
		t.Fatalf("创建执行引擎失败: %v", err)
	}

	// 创建共识实例
	consensus, err := consensus.New(config.ConsensusConfig{
		Algorithm: "pbft",
		MaxFaulty: 1,
		BlockTime: 1000,
		BatchSize: 100,
	}, exec, storage)
	if err != nil {
		t.Fatalf("创建共识失败: %v", err)
	}

	// 创建网络实例
	network, err := New(cfg, consensus)
	if err != nil {
		t.Fatalf("创建网络失败: %v", err)
	}
	defer network.Stop()

	// 测试mDNS状态（启动前应该为false，因为服务还未创建）
	status := network.GetMDNSStatus()
	if status["enabled"].(bool) {
		t.Error("mDNS在启动前应该为false")
	}

	// 启动网络
	err = network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}

	// 等待mDNS服务启动
	time.Sleep(500 * time.Millisecond)

	// 检查mDNS是否启用
	if !network.IsMDNSEnabled() {
		t.Error("mDNS应该被启用")
	}

	// 检查mDNS服务是否创建
	status = network.GetMDNSStatus()
	if !status["service"].(bool) {
		t.Error("mDNS服务应该被创建")
	}

	// 检查服务名
	if status["service_name"] != "chain-network" {
		t.Errorf("期望服务名为 'chain-network', 实际为 '%s'", status["service_name"])
	}
}

// TestMDNSDiscovery 测试mDNS发现功能
func TestMDNSDiscovery(t *testing.T) {
	// 创建第一个网络实例
	cfg1 := config.NetworkConfig{
		Port:     26659,
		Host:     "127.0.0.1",
		MaxPeers: 10,
	}

	storage1, err := storage.New(config.StorageConfig{
		DataDir:     "./test_data1",
		MaxSize:     1024 * 1024,
		CacheSize:   100,
		Compression: false,
	})
	if err != nil {
		t.Fatalf("创建存储1失败: %v", err)
	}
	defer storage1.Stop()

	exec1, err := execution.New(config.ExecutionConfig{
		MaxThreads: 4,
		BatchSize:  100,
		Timeout:    5000,
	}, storage1)
	if err != nil {
		t.Fatalf("创建执行引擎1失败: %v", err)
	}

	consensus1, err := consensus.New(config.ConsensusConfig{
		Algorithm: "pbft",
		MaxFaulty: 1,
		BlockTime: 1000,
		BatchSize: 100,
	}, exec1, storage1)
	if err != nil {
		t.Fatalf("创建共识1失败: %v", err)
	}

	network1, err := New(cfg1, consensus1)
	if err != nil {
		t.Fatalf("创建网络1失败: %v", err)
	}
	defer network1.Stop()

	// 创建第二个网络实例
	cfg2 := config.NetworkConfig{
		Port:     26660,
		Host:     "127.0.0.1",
		MaxPeers: 10,
	}

	storage2, err := storage.New(config.StorageConfig{
		DataDir:     "./test_data2",
		MaxSize:     1024 * 1024,
		CacheSize:   100,
		Compression: false,
	})
	if err != nil {
		t.Fatalf("创建存储2失败: %v", err)
	}
	defer storage2.Stop()

	exec2, err := execution.New(config.ExecutionConfig{
		MaxThreads: 4,
		BatchSize:  100,
		Timeout:    5000,
	}, storage2)
	if err != nil {
		t.Fatalf("创建执行引擎2失败: %v", err)
	}

	consensus2, err := consensus.New(config.ConsensusConfig{
		Algorithm: "pbft",
		MaxFaulty: 1,
		BlockTime: 1000,
		BatchSize: 100,
	}, exec2, storage2)
	if err != nil {
		t.Fatalf("创建共识2失败: %v", err)
	}

	network2, err := New(cfg2, consensus2)
	if err != nil {
		t.Fatalf("创建网络2失败: %v", err)
	}
	defer network2.Stop()

	// 启动两个网络
	err = network1.Start()
	if err != nil {
		t.Fatalf("启动网络1失败: %v", err)
	}

	err = network2.Start()
	if err != nil {
		t.Fatalf("启动网络2失败: %v", err)
	}

	// 等待mDNS发现
	time.Sleep(2 * time.Second)

	// 检查网络1的节点
	peers1 := network1.GetPeers()
	t.Logf("网络1发现的节点: %v", peers1)

	// 检查网络2的节点
	peers2 := network2.GetPeers()
	t.Logf("网络2发现的节点: %v", peers2)

	// 注意：在测试环境中，mDNS可能不会发现其他节点
	// 这是因为测试环境通常是隔离的
	// 这个测试主要是验证mDNS服务能正常启动和运行
}

// TestMDNSHandlePeerFound 测试mDNS节点发现处理
func TestMDNSHandlePeerFound(t *testing.T) {
	network, cleanup := createTestNetworkWithConfig(t, 26661, true)
	defer cleanup()

	// 启用发现（mDNS现在总是启用的）

	// 创建模拟的节点信息
	testPeerID, err := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	if err != nil {
		t.Fatalf("解码测试节点ID失败: %v", err)
	}

	testAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/26662")
	if err != nil {
		t.Fatalf("创建测试地址失败: %v", err)
	}

	addrInfo := peer.AddrInfo{
		ID:    testPeerID,
		Addrs: []multiaddr.Multiaddr{testAddr},
	}

	// 测试HandlePeerFound方法
	// 这应该不会panic
	network.HandlePeerFound(addrInfo)

	// 测试 bootstrap 节点配置（现在由 DHT 自动管理）
	configBootstrapPeers := len(network.config.BootstrapPeers)
	if configBootstrapPeers != 1 {
		t.Errorf("期望1个配置的 bootstrap 节点，但得到了 %d 个", configBootstrapPeers)
	}

	// 再次测试HandlePeerFound
	network.HandlePeerFound(addrInfo)
}

// TestMDNSDisabled 测试mDNS禁用状态
func TestMDNSDisabled(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26664)
	defer cleanup()

	// mDNS现在总是启用的，无法禁用

	// 检查mDNS状态（启动前应该为false）
	status := network.GetMDNSStatus()
	if status["enabled"].(bool) {
		t.Error("mDNS在启动前应该为false")
	}

	// 检查IsMDNSEnabled（启动前应该为false）
	if network.IsMDNSEnabled() {
		t.Error("IsMDNSEnabled在启动前应该返回false")
	}

	// 启动网络（现在总是启动mDNS）
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}

	// 等待一段时间
	time.Sleep(100 * time.Millisecond)

	// 检查mDNS服务是否已创建
	status = network.GetMDNSStatus()
	if !status["service"].(bool) {
		t.Error("mDNS服务应该被创建")
	}
}

// TestMDNSStatus 测试mDNS状态信息
func TestMDNSStatus(t *testing.T) {
	network, cleanup := createTestNetwork(t, 26665)
	defer cleanup()

	// 测试禁用状态（mDNS现在总是启用的）
	status := network.GetMDNSStatus()

	expectedKeys := []string{"enabled", "service", "service_name"}
	for _, key := range expectedKeys {
		if _, exists := status[key]; !exists {
			t.Errorf("状态信息缺少键: %s", key)
		}
	}

	if status["enabled"].(bool) {
		t.Error("enabled应该为false")
	}

	if status["service"].(bool) {
		t.Error("service应该为false")
	}

	if status["service_name"] != "chain-network" {
		t.Errorf("期望服务名为 'chain-network', 实际为 '%s'", status["service_name"])
	}

	// 测试启用状态（启动后mDNS应该被启用）
	err := network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}

	// 等待mDNS服务启动
	time.Sleep(500 * time.Millisecond)

	status = network.GetMDNSStatus()
	if !status["enabled"].(bool) {
		t.Error("enabled应该为true")
	}

	if !status["service"].(bool) {
		t.Error("service应该为true")
	}
}

// BenchmarkMDNSDiscovery 基准测试mDNS发现性能
func BenchmarkMDNSDiscovery(b *testing.B) {
	cfg := config.NetworkConfig{
		Port:     26666,
		Host:     "127.0.0.1",
		MaxPeers: 10,
	}

	storage, err := storage.New(config.StorageConfig{
		DataDir:     "./test_data_bench",
		MaxSize:     1024 * 1024,
		CacheSize:   100,
		Compression: false,
	})
	if err != nil {
		b.Fatalf("创建存储失败: %v", err)
	}
	defer storage.Stop()

	exec, err := execution.New(config.ExecutionConfig{
		MaxThreads: 4,
		BatchSize:  100,
		Timeout:    5000,
	}, storage)
	if err != nil {
		b.Fatalf("创建执行引擎失败: %v", err)
	}

	consensus, err := consensus.New(config.ConsensusConfig{
		Algorithm: "pbft",
		MaxFaulty: 1,
		BlockTime: 1000,
		BatchSize: 100,
	}, exec, storage)
	if err != nil {
		b.Fatalf("创建共识失败: %v", err)
	}

	network, err := New(cfg, consensus)
	if err != nil {
		b.Fatalf("创建网络失败: %v", err)
	}
	defer network.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 测试mDNS状态检查性能
		network.GetMDNSStatus()
		network.IsMDNSEnabled()
	}
}

// TestTrustedPeerRealConnection 测试两个真实节点之间的可信节点连接
func TestTrustedPeerRealConnection(t *testing.T) {
	// 创建第一个网络实例
	cfg1 := config.NetworkConfig{
		Port:     26680,
		Host:     "127.0.0.1",
		MaxPeers: 10,
	}

	storage1, err := storage.New(config.StorageConfig{
		DataDir:     "./test_data_trusted1",
		MaxSize:     1024 * 1024,
		CacheSize:   100,
		Compression: false,
	})
	if err != nil {
		t.Fatalf("创建存储1失败: %v", err)
	}
	defer storage1.Stop()

	exec1, err := execution.New(config.ExecutionConfig{
		MaxThreads: 4,
		BatchSize:  100,
		Timeout:    5000,
	}, storage1)
	if err != nil {
		t.Fatalf("创建执行引擎1失败: %v", err)
	}

	consensus1, err := consensus.New(config.ConsensusConfig{
		Algorithm: "pbft",
		MaxFaulty: 1,
		BlockTime: 1000,
		BatchSize: 100,
	}, exec1, storage1)
	if err != nil {
		t.Fatalf("创建共识1失败: %v", err)
	}

	network1, err := New(cfg1, consensus1)
	if err != nil {
		t.Fatalf("创建网络1失败: %v", err)
	}
	defer network1.Stop()

	// 创建第二个网络实例
	cfg2 := config.NetworkConfig{
		Port:     26681,
		Host:     "127.0.0.1",
		MaxPeers: 10,
	}

	storage2, err := storage.New(config.StorageConfig{
		DataDir:     "./test_data_trusted2",
		MaxSize:     1024 * 1024,
		CacheSize:   100,
		Compression: false,
	})
	if err != nil {
		t.Fatalf("创建存储2失败: %v", err)
	}
	defer storage2.Stop()

	exec2, err := execution.New(config.ExecutionConfig{
		MaxThreads: 4,
		BatchSize:  100,
		Timeout:    5000,
	}, storage2)
	if err != nil {
		t.Fatalf("创建执行引擎2失败: %v", err)
	}

	consensus2, err := consensus.New(config.ConsensusConfig{
		Algorithm: "pbft",
		MaxFaulty: 1,
		BlockTime: 1000,
		BatchSize: 100,
	}, exec2, storage2)
	if err != nil {
		t.Fatalf("创建共识2失败: %v", err)
	}

	network2, err := New(cfg2, consensus2)
	if err != nil {
		t.Fatalf("创建网络2失败: %v", err)
	}
	defer network2.Stop()

	// 启动两个网络
	err = network1.Start()
	if err != nil {
		t.Fatalf("启动网络1失败: %v", err)
	}

	err = network2.Start()
	if err != nil {
		t.Fatalf("启动网络2失败: %v", err)
	}

	// 等待网络启动
	time.Sleep(500 * time.Millisecond)

	// 获取网络2的地址信息
	host2ID := network2.GetHost().ID()
	host2Addrs := network2.GetHost().Addrs()

	// 构建网络2的完整地址
	var network2Addr string
	for _, addr := range host2Addrs {
		if addr.Protocols()[0].Name == "ip4" {
			network2Addr = fmt.Sprintf("%s/p2p/%s", addr.String(), host2ID.String())
			break
		}
	}

	if network2Addr == "" {
		t.Fatal("无法构建网络2的地址")
	}

	t.Logf("网络2地址: %s", network2Addr)

	// 网络1和网络2通过 mDNS 自动发现和连接
	// 等待连接建立
	time.Sleep(2 * time.Second)

	// 验证连接
	peers1 := network1.GetPeers()
	peers2 := network2.GetPeers()

	t.Logf("网络1的节点: %v", peers1)
	t.Logf("网络2的节点: %v", peers2)

	// 检查网络1是否连接到网络2
	connected := false
	for _, peer := range peers1 {
		if peer == host2ID {
			connected = true
			break
		}
	}

	if !connected {
		t.Error("网络1应该连接到网络2")
	}

	// 检查网络2是否连接到网络1
	host1ID := network1.GetHost().ID()
	connected = false
	for _, peer := range peers2 {
		if peer == host1ID {
			connected = true
			break
		}
	}

	if !connected {
		t.Error("网络2应该连接到网络1")
	}

	// 测试网络断开（通过停止网络）
	network1.Stop()

	// 等待连接断开
	time.Sleep(500 * time.Millisecond)

	// 验证连接已断开
	peers2 = network2.GetPeers()
	connected = false
	for _, peer := range peers2 {
		if peer == host1ID {
			connected = true
			break
		}
	}

	if connected {
		t.Error("网络1停止后，网络2应该断开连接")
	}

	t.Log("可信节点连接测试完成")
}

// TestPrivateKeyFileManagement 测试私钥文件管理功能
func TestPrivateKeyFileManagement(t *testing.T) {
	// 测试文件路径
	testKeyPath := "./test_private_key.pem"

	// 清理测试文件
	defer func() {
		if err := os.Remove(testKeyPath); err != nil && !os.IsNotExist(err) {
			t.Logf("清理测试文件失败: %v", err)
		}
	}()

	// 确保文件不存在
	os.Remove(testKeyPath)

	// 创建配置，指定私钥文件路径
	cfg := config.NetworkConfig{
		Port:           26690,
		Host:           "127.0.0.1",
		MaxPeers:       10,
		PrivateKeyPath: testKeyPath,
	}

	// 创建存储实例
	storage, err := storage.New(config.StorageConfig{
		DataDir:     "./test_data_private_key",
		MaxSize:     1024 * 1024,
		CacheSize:   100,
		Compression: false,
	})
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer storage.Stop()

	// 创建执行引擎
	exec, err := execution.New(config.ExecutionConfig{
		MaxThreads: 4,
		BatchSize:  100,
		Timeout:    5000,
	}, storage)
	if err != nil {
		t.Fatalf("创建执行引擎失败: %v", err)
	}

	// 创建共识实例
	consensus, err := consensus.New(config.ConsensusConfig{
		Algorithm: "pbft",
		MaxFaulty: 1,
		BlockTime: 1000,
		BatchSize: 100,
	}, exec, storage)
	if err != nil {
		t.Fatalf("创建共识失败: %v", err)
	}

	// 创建网络实例（应该自动生成私钥文件）
	network1, err := New(cfg, consensus)
	if err != nil {
		t.Fatalf("创建网络1失败: %v", err)
	}
	defer network1.Stop()

	// 验证私钥文件已创建
	if _, err := os.Stat(testKeyPath); os.IsNotExist(err) {
		t.Error("私钥文件应该被创建")
	}

	// 检查文件权限
	if info, err := os.Stat(testKeyPath); err == nil {
		mode := info.Mode()
		if mode&0077 != 0 {
			t.Errorf("私钥文件权限不正确，期望0600，实际: %v", mode)
		}
		t.Logf("私钥文件权限正确: %v", mode)
	}

	// 检查文件内容
	if data, err := os.ReadFile(testKeyPath); err == nil {
		if len(data) == 0 {
			t.Error("私钥文件内容为空")
		}
		t.Logf("私钥文件大小: %d 字节", len(data))
		// 检查是否为PEM格式
		if !bytes.Contains(data, []byte("-----BEGIN LIBP2P PRIVATE KEY-----")) {
			t.Error("私钥文件不是有效的PEM格式")
		}
		t.Log("私钥文件格式正确")
	}

	// 获取第一个网络的节点ID
	host1ID := network1.GetHost().ID()

	// 创建第二个网络实例（应该加载相同的私钥文件）
	cfg.Port = 26691
	network2, err := New(cfg, consensus)
	if err != nil {
		t.Fatalf("创建网络2失败: %v", err)
	}
	defer network2.Stop()

	// 获取第二个网络的节点ID
	host2ID := network2.GetHost().ID()

	// 验证两个网络使用相同的私钥（节点ID应该相同）
	if host1ID != host2ID {
		t.Errorf("两个网络应该使用相同的私钥，但节点ID不同: %s vs %s", host1ID, host2ID)
	} else {
		t.Logf("两个网络成功使用相同的私钥，节点ID: %s", host1ID)
	}

	// 测试没有指定私钥文件路径的情况
	cfg.PrivateKeyPath = ""
	cfg.Port = 26692
	network3, err := New(cfg, consensus)
	if err != nil {
		t.Fatalf("创建网络3失败: %v", err)
	}
	defer network3.Stop()

	host3ID := network3.GetHost().ID()

	// 验证临时私钥生成的节点ID不同
	if host3ID == host1ID {
		t.Error("临时私钥应该生成不同的节点ID")
	} else {
		t.Logf("临时私钥生成了不同的节点ID: %s", host3ID)
	}
}

// TestConnectionManager 测试连接管理器功能
func TestConnectionManager(t *testing.T) {
	// 创建配置，设置较小的最大节点数以便测试
	cfg := config.NetworkConfig{
		Port:           26700,
		Host:           "127.0.0.1",
		MaxPeers:       2, // 设置较小的最大节点数
		PrivateKeyPath: "",
	}

	// 创建存储实例
	storage, err := storage.New(config.StorageConfig{
		DataDir:     "./test_data_conn_mgr",
		MaxSize:     1024 * 1024,
		CacheSize:   100,
		Compression: false,
	})
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer storage.Stop()

	// 创建执行引擎
	exec, err := execution.New(config.ExecutionConfig{
		MaxThreads: 4,
		BatchSize:  100,
		Timeout:    5000,
	}, storage)
	if err != nil {
		t.Fatalf("创建执行引擎失败: %v", err)
	}

	// 创建共识实例
	consensus, err := consensus.New(config.ConsensusConfig{
		Algorithm: "pbft",
		MaxFaulty: 1,
		BlockTime: 1000,
		BatchSize: 100,
	}, exec, storage)
	if err != nil {
		t.Fatalf("创建共识失败: %v", err)
	}

	// 创建网络实例
	network, err := New(cfg, consensus)
	if err != nil {
		t.Fatalf("创建网络失败: %v", err)
	}
	defer network.Stop()

	// 启动网络
	err = network.Start()
	if err != nil {
		t.Fatalf("启动网络失败: %v", err)
	}

	// 等待网络启动
	time.Sleep(100 * time.Millisecond)

	// 测试连接统计信息
	connStats := network.GetConnectionStats()

	// 验证统计信息包含所有必要字段
	expectedKeys := []string{"current_peers", "max_peers", "usage_percentage", "bootstrap_peers"}
	for _, key := range expectedKeys {
		if _, exists := connStats[key]; !exists {
			t.Errorf("连接统计信息缺少键: %s", key)
		}
	}

	// 验证最大节点数设置正确
	if connStats["max_peers"] != 2 {
		t.Errorf("期望最大节点数为 2，实际为 %v", connStats["max_peers"])
	}

	// 验证当前节点数应该为0（初始状态）
	if connStats["current_peers"] != 0 {
		t.Errorf("期望当前节点数为 0，实际为 %v", connStats["current_peers"])
	}

	// 验证使用率应该为0
	if connStats["usage_percentage"] != 0.0 {
		t.Errorf("期望使用率为 0.0，实际为 %v", connStats["usage_percentage"])
	}

	t.Logf("连接统计信息: %+v", connStats)
}
