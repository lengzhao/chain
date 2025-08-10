# 可信节点管理简化

## 概述

根据用户反馈，网络模块已经简化了 bootstrap 节点的管理方式。Bootstrap 节点现在作为 DHT bootstrap 节点，由 DHT 自动管理，无需网络模块手动管理连接状态。

## 主要改进

### 1. 移除手动可信节点管理

**之前**：
- 网络模块维护 `trustedPeers` 映射
- 手动管理 bootstrap 节点的连接和断开
- 提供 `AddTrustedPeer`、`RemoveTrustedPeer`、`IsTrustedPeer` 等方法

**现在**：
- Bootstrap 节点直接配置为 DHT bootstrap 节点
- DHT 自动处理与 bootstrap 节点的连接
- 网络模块不再维护 bootstrap 节点状态

### 2. DHT Bootstrap 集成

```go
// 解析可信节点为 bootstrap 节点
var bootstrapPeers []peer.AddrInfo
for _, peerAddr := range cfg.TrustedPeers {
    maddr, err := multiaddr.NewMultiaddr(peerAddr)
    if err != nil {
        continue
    }

    addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
    if err != nil {
        continue
    }

    bootstrapPeers = append(bootstrapPeers, *addrInfo)
}

// 创建 DHT 实例，设置 bootstrap 节点
dhtInstance, err := dht.New(context.Background(), h, dht.BootstrapPeers(bootstrapPeers...))
```

### 3. 简化的网络启动

**之前**：
```go
func (n *Network) Start() error {
    // 1. 连接可信节点
    if err := n.connectToTrustedPeers(); err != nil {
        return fmt.Errorf("连接可信节点失败: %w", err)
    }

    // 2. 启动mDNS发现
    n.startMDNSDiscovery()

    return nil
}
```

**现在**：
```go
func (n *Network) Start() error {
    // 1. 启动mDNS发现
    n.startMDNSDiscovery()

    	// 2. Bootstrap 节点现在作为 DHT bootstrap 节点，DHT 会自动处理连接
	fmt.Printf("网络已启动，Bootstrap 节点作为 DHT bootstrap 节点\n")

    return nil
}
```

## 移除的组件

### 1. 结构体字段
```go
// 移除的字段
trustedPeers map[peer.ID]bool
trustedMu     sync.RWMutex
```

### 2. 方法
- `initializeTrustedPeers()` - 初始化可信节点
- `connectToTrustedPeers()` - 连接可信节点
- `AddTrustedPeer()` - 添加可信节点
- `RemoveTrustedPeer()` - 移除可信节点
- `IsTrustedPeer()` - 检查是否为可信节点
- `GetTrustedPeers()` - 获取可信节点列表

### 3. 测试更新
- 更新了所有相关测试，使用配置中的可信节点信息
- 创建了 `createTestNetworkWithConfig()` 辅助函数
- 移除了对已删除方法的调用

## 配置方式

### 配置文件
```yaml
network:
  port: 26656
  host: "0.0.0.0"
  max_peers: 50
  private_key_path: "./private_key.pem"
  trusted_peers:
    # 可信节点列表（作为 DHT bootstrap 节点）
    - "/ip4/192.168.1.100/tcp/26656/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N"
    - "/ip4/192.168.1.101/tcp/26656/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N"
```

### 代码配置
```go
cfg := config.NetworkConfig{
    Port:           26656,
    Host:           "127.0.0.1",
    MaxPeers:       50,
    PrivateKeyPath: "./private_key.pem",
    	BootstrapPeers: []string{
        "/ip4/192.168.1.100/tcp/26656/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
        "/ip4/192.168.1.101/tcp/26656/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
    },
}
```

## 优势

### 1. 简化架构
- 减少了网络模块的复杂性
- 移除了重复的连接管理逻辑
- 代码更加简洁和易于维护

### 2. 更好的网络发现
- DHT 提供更可靠的网络发现机制
- 自动处理节点加入和离开
- 支持网络分区恢复

### 3. 自动连接管理
- DHT 自动处理与 bootstrap 节点的连接
- 支持连接重试和故障恢复
- 无需手动干预

### 4. 扩展性
- 支持大规模网络部署
- 自动负载均衡
- 高效的节点查找

## 兼容性

### 配置兼容性
- 现有的 `trusted_peers` 配置仍然有效
- 配置格式保持不变
- 无需修改现有配置文件

### API 兼容性
- 移除了可信节点管理相关的 API
- 保留了连接统计信息中的可信节点数量
- 其他 API 保持不变

## 测试验证

所有测试都已更新并通过：

```bash
go test ./network -v
```

测试覆盖：
- 可信节点配置验证
- DHT bootstrap 功能
- 连接统计信息
- mDNS 发现功能
- 网络连接管理

## 示例更新

示例文件已更新，移除了对已删除方法的调用：

```go
// 显示配置的可信节点信息
configBootstrapPeers := len(cfg.BootstrapPeers)
fmt.Printf("配置的 bootstrap 节点数量: %d\n", configBootstrapPeers)
for i, peerAddr := range cfg.BootstrapPeers {
	fmt.Printf("Bootstrap 节点 %d: %s\n", i+1, peerAddr)
}

	// 示例：Bootstrap 节点由 DHT 自动管理
	fmt.Println("\n示例：Bootstrap 节点由 DHT 自动管理...")
	fmt.Println("Bootstrap 节点现在由 DHT 自动管理，无需手动添加或检查")
	fmt.Println("DHT 会自动处理与 bootstrap 节点的连接和路由")
```

## 总结

这次简化大大减少了网络模块的复杂性，同时提高了网络发现和连接的可靠性。可信节点现在作为 DHT bootstrap 节点，由 DHT 自动管理，这是一个更加优雅和高效的解决方案。 