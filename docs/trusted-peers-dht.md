# 可信节点作为 DHT Bootstrap 节点

## 概述

网络模块现在将 bootstrap 节点（BootstrapPeers）配置为 DHT（分布式哈希表）的 bootstrap 节点，这样可以更好地利用 bootstrap 节点进行网络发现和路由。

## 功能特性

### 1. DHT Bootstrap 功能
- 可信节点自动成为 DHT 的 bootstrap 节点
- 提供网络发现和路由功能
- DHT 自动管理与 bootstrap 节点的连接

### 2. 网络发现增强
- 通过 DHT 进行节点发现
- 提高网络连通性和稳定性
- 支持大规模网络部署

### 3. 自动连接管理
- DHT 自动处理与 bootstrap 节点的连接
- 无需手动管理连接状态
- 支持连接重试和故障恢复

## 配置

在 `config.yaml` 中配置可信节点：

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

## 工作原理

### DHT Bootstrap 配置

```go
// 解析可信节点为 bootstrap 节点
var bootstrapPeers []peer.AddrInfo
for _, peerAddr := range cfg.BootstrapPeers {
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

### 自动连接管理

1. **Bootstrap 连接**: DHT 自动连接到配置的可信节点
2. **网络发现**: 通过 DHT 发现网络中的其他节点
3. **路由功能**: 提供节点查找和消息路由功能
4. **故障恢复**: 自动处理连接断开和重连

## API 使用

### 动态添加可信节点

```go
// 动态添加可信节点作为 DHT bootstrap 节点
err := network.AddTrustedPeer("/ip4/192.168.1.102/tcp/26656/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
if err != nil {
    log.Printf("添加可信节点失败: %v", err)
} else {
    fmt.Println("可信节点已添加为 DHT bootstrap 节点")
}
```

### 检查可信节点状态

```go
// 获取可信节点列表
configBootstrapPeers := len(network.config.BootstrapPeers)
fmt.Printf("Bootstrap 节点数量: %d\n", configBootstrapPeers)

// 检查特定节点是否为 bootstrap 节点
peerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
for _, peerAddr := range network.config.BootstrapPeers {
    maddr, _ := multiaddr.NewMultiaddr(peerAddr)
    addrInfo, _ := peer.AddrInfoFromP2pAddr(maddr)
    if addrInfo.ID == peerID {
        fmt.Println("该节点是 bootstrap 节点")
        break
    }
}
```

## 示例输出

```
Bootstrap 节点已添加为 DHT bootstrap 节点: QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N
网络已启动，Bootstrap 节点作为 DHT bootstrap 节点

=== 连接统计信息 ===
当前连接节点数: 3
最大节点数: 50
连接使用率: 6.00%
Bootstrap 节点数: 2
```

## 最佳实践

### 1. 选择 Bootstrap 节点
- 选择稳定、可靠的节点作为 bootstrap 节点
- 建议选择地理位置分散的节点
- 确保 bootstrap 节点有足够的带宽和计算资源

### 2. 配置建议
- 建议配置 3-10 个 bootstrap 节点
- 避免配置过多 bootstrap 节点（可能影响性能）
- 定期检查和更新 bootstrap 节点列表

### 3. 网络拓扑
- Bootstrap 节点应该分布在不同的网络区域
- 考虑网络延迟和带宽限制
- 确保网络连通性

### 4. 监控和维护
- 定期监控 bootstrap 节点的可用性
- 及时替换不可用的 bootstrap 节点
- 监控 DHT 路由性能

## 优势

### 1. 网络稳定性
- DHT 提供可靠的网络发现机制
- 自动处理节点加入和离开
- 支持网络分区恢复

### 2. 扩展性
- 支持大规模网络部署
- 自动负载均衡
- 高效的节点查找

### 3. 容错性
- 支持节点故障自动恢复
- 网络分区自动修复
- 连接重试机制

## 注意事项

1. **性能影响**: DHT 操作会消耗少量 CPU 和内存资源
2. **网络延迟**: DHT 查找可能增加少量网络延迟
3. **配置变更**: 修改可信节点列表需要重启网络服务
4. **依赖关系**: 依赖可信节点的可用性

## 测试

运行 bootstrap 节点测试：

```bash
go test ./network -v -run "TestTrustedPeers"
```

## 故障排除

### 常见问题

1. **Bootstrap 节点连接失败**
   - 检查网络连通性
   - 验证节点地址格式
   - 确认节点服务状态

2. **DHT 查找失败**
   - 检查 bootstrap 节点配置
   - 验证网络防火墙设置
   - 确认 DHT 服务状态

3. **网络发现缓慢**
   - 增加 bootstrap 节点数量
   - 优化网络配置
   - 检查网络延迟 