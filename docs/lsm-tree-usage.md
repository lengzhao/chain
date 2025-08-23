# LSM树存储使用指南

## 概述

我们的LSM树实现提供了统一的接口，通过参数来控制存储模式：

- **内存模式**：传入空字符串 `""`，使用临时目录实现快速存储
- **持久化模式**：传入目录路径，数据持久化到指定目录

## 统一接口

```go
// 统一的创建接口
func NewLSMTree(dataDir string) (*LSMTree, error)
```

## 使用方式

### 1. 内存模式（临时存储）

```go
package main

import (
    "fmt"
    "github.com/govm-net/chain/storage"
)

func main() {
    // 创建内存模式的LSM-Tree（传入空字符串）
    l, err := storage.NewLSMTree("")
    if err != nil {
        panic(err)
    }
    defer l.Stop() // 自动清理临时目录

    // 检查是否为内存模式
    if l.IsMemoryMode() {
        fmt.Println("✅ 运行在内存模式")
        fmt.Printf("📁 临时目录: %s\n", l.GetTempDir())
    }

    // 基本操作
    key := []byte("test_key")
    value := []byte("test_value")

    // 设置值
    err = l.Set(key, value)
    if err != nil {
        panic(err)
    }

    // 获取值
    retrievedValue, err := l.Get(key)
    if err != nil {
        panic(err)
    }
    fmt.Printf("获取到的值: %s\n", string(retrievedValue))
}
```

### 2. 持久化模式（永久存储）

```go
package main

import (
    "fmt"
    "github.com/govm-net/chain/storage"
)

func main() {
    // 创建持久化模式的LSM-Tree（传入目录路径）
    dataDir := "./data/blockchain"
    l, err := storage.NewLSMTree(dataDir)
    if err != nil {
        panic(err)
    }
    defer l.Stop()

    // 检查是否为持久化模式
    if !l.IsMemoryMode() {
        fmt.Println("✅ 运行在持久化模式")
        fmt.Printf("📁 数据目录: %s\n", dataDir)
    }

    // 基本操作
    key := []byte("blockchain_key")
    value := []byte("blockchain_value")

    err = l.Set(key, value)
    if err != nil {
        panic(err)
    }

    retrievedValue, err := l.Get(key)
    if err != nil {
        panic(err)
    }
    fmt.Printf("获取到的值: %s\n", string(retrievedValue))
}
```

### 3. 批量操作

```go
package main

import (
    "fmt"
    "github.com/govm-net/chain/storage"
)

func main() {
    // 创建LSM-Tree（可以是内存模式或持久化模式）
    l, err := storage.NewLSMTree("") // 内存模式
    if err != nil {
        panic(err)
    }
    defer l.Stop()

    // 创建批量操作
    batch := l.Batch()
    
    // 添加多个操作
    batch.Put([]byte("key1"), []byte("value1"))
    batch.Put([]byte("key2"), []byte("value2"))
    batch.Put([]byte("key3"), []byte("value3"))
    batch.Delete([]byte("key2")) // 删除key2

    // 执行批量操作
    err = l.WriteBatch(batch)
    if err != nil {
        panic(err)
    }

    // 验证结果
    value1, _ := l.Get([]byte("key1"))
    fmt.Printf("key1: %s\n", string(value1))

    value2, _ := l.Get([]byte("key2"))
    if value2 == nil {
        fmt.Println("key2: 已删除")
    }

    value3, _ := l.Get([]byte("key3"))
    fmt.Printf("key3: %s\n", string(value3))
}
```

## 模式对比

| 特性 | 内存模式 (`""`) | 持久化模式 (`"path"`) |
|------|----------------|---------------------|
| **数据持久性** | 程序退出后丢失 | 数据永久保存 |
| **性能** | 最快 | 较快 |
| **存储位置** | 临时目录 | 指定目录 |
| **适用场景** | 测试、缓存、临时数据 | 生产环境、重要数据 |
| **资源管理** | 自动清理 | 手动管理 |

## 使用场景

### 1. 测试环境
```go
// 在单元测试中使用内存模式
func TestMyFunction(t *testing.T) {
    l, err := storage.NewLSMTree("") // 内存模式
    if err != nil {
        t.Fatal(err)
    }
    defer l.Stop()
    
    // 执行测试...
}
```

### 2. 生产环境
```go
// 在生产环境中使用持久化模式
func NewBlockchainStorage() *storage.LSMTree {
    dataDir := "/var/blockchain/data"
    l, err := storage.NewLSMTree(dataDir)
    if err != nil {
        panic(err)
    }
    return l
}
```

### 3. 缓存层
```go
// 作为高性能缓存使用
func NewCache() *storage.LSMTree {
    l, err := storage.NewLSMTree("") // 内存模式
    if err != nil {
        panic(err)
    }
    return l
}
```

### 4. 开发环境
```go
// 开发环境可以选择内存模式或本地目录
func NewDevStorage() *storage.LSMTree {
    var dataDir string
    if os.Getenv("DEV_MEMORY_MODE") == "true" {
        dataDir = "" // 内存模式
    } else {
        dataDir = "./dev_data" // 本地目录
    }
    
    l, err := storage.NewLSMTree(dataDir)
    if err != nil {
        panic(err)
    }
    return l
}
```

## 性能基准测试

运行基准测试来比较不同模式的性能：

```bash
# 内存模式基准测试
go test ./storage -bench=BenchmarkMemoryLSMTree

# 持久化模式基准测试
go test ./storage -bench=BenchmarkPersistentLSMTree
```

## 注意事项

1. **资源清理**：记得调用`Stop()`方法释放资源
2. **内存模式**：程序退出后数据丢失，适合临时数据
3. **持久化模式**：数据永久保存，适合重要数据
4. **并发安全**：LSM树实现是线程安全的
5. **目录权限**：确保有足够的权限创建和访问目录

## 配置选项

LSM树使用以下默认配置：

```go
opts := &opt.Options{
    WriteBuffer: 64 * 1024 * 1024,        // 64MB写缓冲区
    Compression: opt.SnappyCompression,    // Snappy压缩
    BlockCacheCapacity: 8 * 1024 * 1024,  // 8MB块缓存
    OpenFilesCacheCapacity: 1000,          // 1000个打开文件缓存
    WriteL0SlowdownTrigger: 8,             // L0层慢速写入触发
    WriteL0PauseTrigger: 12,               // L0层暂停写入触发
}
```

## 总结

统一的LSM树接口提供了灵活性和易用性：

- **简单易用**：一个接口支持两种模式
- **性能优化**：内存模式提供最快速度
- **数据安全**：持久化模式确保数据不丢失
- **自动管理**：内存模式自动清理临时文件

根据你的具体需求选择合适的模式：测试和临时数据使用内存模式，生产环境使用持久化模式。
