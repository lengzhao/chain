# Storage模块Object概念设计 - 共识文档

## 明确的需求描述和验收标准

### 核心需求
1. **Object概念**: 支持智能合约创建Object，Object是key/value的非关系型存储模块
2. **Object属性**: Object包含owner和contract属性
3. **权限控制**: 只有合约可以读写object；如果要将object添加为Writes，要求owner为交易的发起人或合约
4. **Object转移**: owner可以将Object转移给其他人，由交易控制，支持单个Object转移
5. **AccessList集成**: AccessList里的Reads/Writes就是Object的ID，直接使用现有的Hash数组
6. **生命周期**: Object默认生命周期有2年

### 验收标准
1. **功能完整性**: 
   - Object的创建、读取、修改、删除功能正常
   - Object转移功能正常
   - AccessList与Object ID关联正确
   - 权限控制机制正确

2. **性能要求**:
   - Object操作性能不低于现有存储操作
   - 支持高并发访问
   - 内存使用合理

3. **兼容性要求**:
   - 与现有存储架构完全兼容
   - 不影响现有功能
   - 向后兼容现有AccessList机制

4. **可靠性要求**:
   - 数据一致性保证
   - 故障恢复能力
   - 事务原子性

## 技术实现方案和技术约束

### 技术实现方案

#### 1. Object数据结构设计
```go
type Object struct {
    ID        Hash   `json:"id"`         // Object唯一标识
    Owner     []byte `json:"owner"`      // Object所有者
    Contract  []byte `json:"contract"`   // 创建Object的合约
    ExpiresAt int64  `json:"expires_at"` // 过期时间戳
}

// ObjectStorage 继承Storage功能，提供Object语义的存储操作
type ObjectStorage struct {
    *Storage
    objectID Hash
}
```

#### 2. 存储结构设计
- **Object元数据存储**: 使用前缀 `object:meta:` + ObjectID
- **Object数据存储**: 使用前缀 `object:data:` + ObjectID + `:` + key
- **Object索引存储**: 使用前缀 `object:index:` + Owner/Contract

#### 3. 权限控制机制
- **读取权限**: 只有合约可以读取Object数据
- **写入权限**: 只有合约可以修改Object数据
- **转移权限**: owner可以将Object转移给其他人
- **AccessList验证**: 写入AccessList需要验证owner权限

#### 4. 生命周期管理
- **创建**: 设置过期时间（当前时间+2年）
- **过期检查**: 定期检查过期Object
- **清理机制**: 自动清理过期Object

#### 5. 接口设计
- **Get接口**: 与Storage.Get接口一致，参数为key []byte，返回value []byte和error
- **Set接口**: 与Storage.Set接口一致，参数为key []byte和value []byte，返回error
- **继承设计**: ObjectStorage继承Storage功能，通过objectID前缀提供Object语义的存储操作
- **数据隔离**: 每个Object的数据通过objectID前缀进行隔离

### 技术约束

#### 1. 架构约束
- 必须基于现有LSM-Tree存储引擎
- 必须保持与现有存储接口的兼容性
- 必须使用现有的Hash类型作为Object ID

#### 2. 性能约束
- Object操作延迟不超过现有存储操作的1.5倍
- 内存使用增长不超过20%
- 支持至少1000个并发Object操作

#### 3. 兼容性约束
- 不能修改现有AccessList结构
- 不能影响现有交易处理流程
- 必须保持现有API接口不变

## 任务边界限制和验收标准

### 任务边界限制

#### 包含范围
- Object数据结构和存储设计
- Object Get/Set接口实现（与Storage接口一致）
- Object转移功能实现
- 权限控制机制实现
- 生命周期管理实现
- 与AccessList的集成

#### 不包含范围
- 智能合约执行引擎
- Object权限验证逻辑（由上层处理）
- Object序列化格式优化
- Object版本控制机制
- 复杂的权限模型

### 验收标准

#### 功能验收
1. **Object创建**: 合约可以成功创建Object，设置正确的owner和contract
2. **Object读取**: 合约可以通过Get接口读取Object数据
3. **Object修改**: 合约可以通过Set接口修改Object数据
4. **Object转移**: owner可以成功转移Object给其他人
5. **权限控制**: 权限控制机制正确工作
6. **生命周期**: Object在2年后正确过期
7. **接口兼容**: Object的Get/Set接口与Storage接口完全兼容

#### 性能验收
1. **延迟测试**: Object操作延迟在可接受范围内
2. **并发测试**: 支持高并发Object操作
3. **内存测试**: 内存使用在合理范围内

#### 兼容性验收
1. **向后兼容**: 现有功能不受影响
2. **接口兼容**: 现有API接口正常工作
3. **数据兼容**: 现有数据格式保持不变

#### 可靠性验收
1. **一致性测试**: 数据一致性得到保证
2. **故障恢复**: 系统故障后能正确恢复
3. **事务测试**: 事务原子性得到保证

## 确认所有不确定性已解决

### 已确认的技术决策
1. ✅ Object ID使用Hash类型
2. ✅ Object数据存储在现有LSM-Tree中
3. ✅ Object元数据与数据分开存储
4. ✅ 使用gob格式进行序列化
5. ✅ 权限模型：只有合约可以读写object
6. ✅ Object转移：支持单个Object转移
7. ✅ 不实现版本控制
8. ✅ Object生命周期：2年
9. ✅ AccessList使用现有Hash数组

### 已确认的架构决策
1. ✅ 在storage模块内创建Object子模块
2. ✅ 扩展现有存储接口
3. ✅ Object元数据使用专门缓存策略
4. ✅ 利用现有存储机制保证原子性

### 风险控制
1. **性能风险**: 通过缓存和索引优化控制
2. **兼容性风险**: 通过严格的接口设计控制
3. **数据一致性风险**: 通过事务机制控制
4. **权限安全风险**: 通过严格的权限验证控制
