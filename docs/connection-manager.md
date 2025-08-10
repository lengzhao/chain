# 连接管理器 (Connection Manager)

## 概述

网络模块现在使用 libp2p 的 `ConnectionManager` 来管理节点连接数量，确保网络连接数不会超过配置的限制。

## 功能特性

### 1. 连接数限制
- 通过 `MaxPeers` 配置项设置最大连接节点数
- 自动管理连接池，防止连接数过多
- 支持优雅期，避免频繁断开重连

### 2. 连接统计
- 实时监控当前连接节点数
- 计算连接使用率
- 显示可信节点数量
- 提供详细的连接状态信息

## 配置

在 `config.yaml` 中配置：

```yaml
network:
  port: 26656
  host: "0.0.0.0"
  max_peers: 50  # 最大连接节点数
  private_key_path: "./private_key.pem"
  trusted_peers:
    # 可信节点列表
```

## API 使用

### 获取连接统计信息

```go
// 获取连接统计信息
connStats := network.GetConnectionStats()

// 统计信息包含以下字段：
// - current_peers: 当前连接节点数
// - max_peers: 最大节点数
// - usage_percentage: 连接使用率
// - trusted_peers: 可信节点数

fmt.Printf("当前连接节点数: %v\n", connStats["current_peers"])
fmt.Printf("最大节点数: %v\n", connStats["max_peers"])
fmt.Printf("连接使用率: %.2f%%\n", connStats["usage_percentage"])
fmt.Printf("可信节点数: %v\n", connStats["trusted_peers"])
```

## 工作原理

### 连接管理器配置

```go
// 创建连接管理器
connManager, err := connmgr.NewConnManager(
    20,                              // 最小连接数（默认值）
    cfg.MaxPeers,                    // 最大连接数
    connmgr.WithGracePeriod(time.Minute*5), // 优雅期
)
```

### 自动连接管理

1. **最小连接数**: 系统会尝试保持至少20个连接，确保网络连通性
2. **最大连接数**: 当连接数达到 `MaxPeers` 时，新的连接请求会被拒绝
3. **优雅期**: 连接断开后有5分钟的优雅期，避免频繁重连
4. **DHT Bootstrap**: 可信节点作为 DHT bootstrap 节点，提供网络发现和路由
5. **自动清理**: 长时间无活动的连接会被自动清理

## 示例输出

```
=== 连接统计信息 ===
当前连接节点数: 3
最大节点数: 50
连接使用率: 6.00%
可信节点数: 2

当前连接的节点数量: 3
节点 1: 12D3KooW... (已连接)
节点 2: 12D3KooW... (已连接)
节点 3: 12D3KooW... (已连接)
```

## 最佳实践

### 1. 合理设置 MaxPeers
- 根据网络环境和资源情况设置合适的最大连接数
- 系统会自动保持至少20个连接以确保网络连通性
- 建议值：50-200 个节点（考虑到最小连接数为20）

### 2. 监控连接状态
- 定期检查连接统计信息
- 关注连接使用率，避免接近100%

### 3. 可信节点配置
- 将重要的节点配置为可信节点
- 可信节点作为 DHT bootstrap 节点，提供网络发现和路由功能
- DHT 会自动管理与 bootstrap 节点的连接

### 4. 网络稳定性
- 连接管理器会自动处理连接断开和重连
- 无需手动干预连接管理

## 测试

运行连接管理器测试：

```bash
go test ./network -v -run "TestConnectionManager"
```

## 注意事项

1. **性能影响**: 连接管理器会消耗少量CPU和内存资源
2. **网络延迟**: 连接限制可能导致某些连接请求被延迟
3. **配置变更**: 修改 `MaxPeers` 需要重启网络服务
4. **监控建议**: 建议在生产环境中监控连接统计信息 