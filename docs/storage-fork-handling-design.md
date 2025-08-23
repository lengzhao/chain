# 存储方案和区块分叉处理设计

## 1. 存储方案设计

### 1.1 当前存储架构
- **LSM树存储引擎**：基于goleveldb的高性能键值存储
- **Object存储**：智能合约的键值非关系型存储
- **缓存层**：内存缓存提升访问性能
- **分层存储**：热数据在内存，冷数据在磁盘

### 1.2 参考主流项目的优化建议

#### 比特币模式（链式存储）
```go
// 区块存储结构
type BlockStorage struct {
    BlockHash    types.Hash
    Transactions []types.Transaction
    StateRoot    types.Hash
    ParentHash   types.Hash
    Timestamp    int64
}
```

#### 以太坊模式（状态存储）
```go
// 状态存储结构
type StateStorage struct {
    AccountStates map[types.Address]*AccountState
    ContractCode  map[types.Address][]byte
    StorageTrie   *Trie // Merkle Patricia Trie
}
```

#### EOS/波场模式（混合存储）
```go
// 混合存储结构
type HybridStorage struct {
    // 图数据库：处理复杂关系
    GraphDB *GraphDatabase
    
    // 键值存储：处理简单数据
    KVStore *LSMTree
    
    // 对象存储：智能合约数据
    ObjectStore *ObjectStorage
}
```

## 2. 区块分叉处理机制

### 2.1 分叉检测
```go
// 分叉检测器
type ForkDetector struct {
    currentChain []types.Block
    forkChains   map[types.Hash][]types.Block
    consensus    *consensus.Consensus
}

func (fd *ForkDetector) DetectFork(newBlock *types.Block) *ForkInfo {
    // 检查新区块是否与当前链冲突
    if fd.isConflicting(newBlock) {
        return &ForkInfo{
            Type:        ForkType,
            BlockHeight: newBlock.Header.Height,
            ConflictingBlocks: []types.Block{*newBlock},
        }
    }
    return nil
}
```

### 2.2 分叉处理策略

#### 策略1：最长链原则（PoW风格）
```go
func (fd *ForkDetector) HandleForkByLongestChain() {
    chains := fd.getAllChains()
    longestChain := fd.findLongestChain(chains)
    
    // 切换到最长链
    fd.switchToChain(longestChain)
    
    // 回滚短链上的交易
    fd.rollbackTransactions(longestChain)
}
```

#### 策略2：DPoS投票机制（EOS/波场风格）
```go
func (fd *ForkDetector) HandleForkByDPoSVoting() {
    // 收集验证者投票
    votes := fd.collectValidatorVotes()
    
    // 根据投票权重选择主链
    mainChain := fd.selectChainByVotes(votes)
    
    // 执行链切换
    fd.switchToChain(mainChain)
}
```

#### 策略3：混合策略（推荐）
```go
func (fd *ForkDetector) HandleForkHybrid() {
    // 1. 首先尝试DPoS投票解决
    if fd.canResolveByVoting() {
        fd.HandleForkByDPoSVoting()
        return
    }
    
    // 2. 如果投票无法解决，使用最长链原则
    fd.HandleForkByLongestChain()
}
```

### 2.3 状态回滚机制
```go
// 状态回滚管理器
type StateRollbackManager struct {
    storage *storage.Storage
    snapshots map[int64]*StateSnapshot
}

type StateSnapshot struct {
    BlockHeight int64
    StateRoot   types.Hash
    Changes     []StateChange
}

func (srm *StateRollbackManager) RollbackToHeight(height int64) error {
    snapshot := srm.snapshots[height]
    if snapshot == nil {
        return fmt.Errorf("snapshot not found for height %d", height)
    }
    
    // 应用状态变更
    for _, change := range snapshot.Changes {
        srm.applyStateChange(change)
    }
    
    return nil
}
```

## 3. 实现优先级

### 3.1 第一阶段：基础分叉检测
- [ ] 实现分叉检测逻辑
- [ ] 添加区块冲突检查
- [ ] 实现简单的链切换机制

### 3.2 第二阶段：状态管理
- [ ] 实现状态快照机制
- [ ] 添加状态回滚功能
- [ ] 优化存储性能

### 3.3 第三阶段：高级功能
- [ ] 实现DPoS投票机制
- [ ] 添加分叉预测
- [ ] 实现自动恢复机制

## 4. 性能考虑

### 4.1 存储优化
- 使用LSM树提供高性能写入
- 实现数据压缩减少存储空间
- 添加缓存层提升读取性能

### 4.2 分叉处理优化
- 异步处理分叉检测
- 批量处理状态回滚
- 实现增量快照减少存储开销

## 5. 存储回滚机制

### 5.1 事务处理机制
```go
// 核心组件
type RollbackManager struct {
    storage      *Storage
    transactions map[string]*Transaction
    snapshots    map[int64]*StateSnapshot
    mu           sync.RWMutex
}

// 事务对象
type Transaction struct {
    ID        string
    BlockHash types.Hash
    Changes   []StateChange
    Status    TransactionStatus
    CreatedAt int64
}

// 状态变更
type StateChange struct {
    Key       []byte
    OldValue  []byte
    NewValue  []byte
    Operation ChangeOperation
}
```

### 5.2 回滚处理流程
1. **开始事务**：为每个区块创建事务
2. **记录变更**：记录所有状态变更（旧值和新值）
3. **提交事务**：应用变更到存储并创建快照
4. **回滚事务**：按相反顺序恢复旧值

### 5.3 支持的操作类型
- **OpSet**：设置值（回滚时恢复旧值）
- **OpDelete**：删除值（回滚时恢复被删除的值）
- **OpObjectCreate**：创建Object（回滚时删除Object）
- **OpObjectDelete**：删除Object（回滚时恢复Object）
- **OpObjectTransfer**：转移Object（回滚时恢复所有权）

### 5.4 性能优化
- **批量处理**：批量应用和回滚变更
- **并发安全**：使用读写锁保证并发安全
- **内存管理**：定期清理过期的事务和快照

## 6. 安全考虑

### 6.1 数据一致性
- 确保状态回滚的原子性
- 防止部分回滚导致的数据不一致
- 实现事务机制保证数据完整性

### 6.2 攻击防护
- 防止恶意分叉攻击
- 实现分叉深度限制
- 添加验证者权重验证

### 6.3 回滚安全
- 验证回滚操作的合法性
- 防止无限回滚攻击
- 实现回滚深度限制
