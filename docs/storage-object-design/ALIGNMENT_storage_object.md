# Storage模块Object概念设计 - 对齐文档

## 项目上下文分析

### 现有项目结构
- **技术栈**: Go语言，使用LSM-Tree存储引擎
- **架构模式**: 模块化设计，storage模块独立
- **现有组件**: 
  - LSM-Tree存储引擎 (storage/lsm_tree.go)
  - 缓存层 (storage/cache.go)
  - 基础存储接口 (storage/storage.go)
  - 类型定义 (types/types.go)

### 现有代码模式
- 使用Hash类型作为唯一标识符
- AccessList包含Reads/Writes字段，目前是Hash数组
- 存储接口提供基础的Get/Set/Delete操作
- 使用gob进行序列化

### 业务域和数据模型
- 区块链系统，支持智能合约
- 交易包含AccessList，用于访问控制
- 支持DPOS治理机制
- 需要高性能的key-value存储

## 需求理解确认

### 原始需求
1. 支持Object概念：智能合约可以创建Object
2. Object有owner/Contract属性
3. Object是key/value的非关系型存储模块
4. owner可以将Object转移给其他人
5. Contract可以修改object里的数据
6. AccessList里的Reads/Writes就是Object的ID

### 边界确认
**明确任务范围**:
- 在现有storage模块基础上扩展Object概念
- 设计Object的数据结构和存储方式
- 实现Object的创建、转移、修改功能
- 将AccessList与Object ID关联
- 保持与现有存储架构的兼容性

**任务边界限制**:
- 不涉及智能合约执行引擎的具体实现
- 不涉及Object的权限验证逻辑（由上层处理）
- 不涉及Object的序列化格式优化（使用现有gob格式）
- 不涉及Object的版本控制机制

### 需求理解
基于对现有项目的理解，这个需求是要在现有的key-value存储基础上，增加Object的概念，使智能合约能够：
1. 创建具有所有权的数据对象
2. 管理Object的访问权限
3. 支持Object的转移操作
4. 通过AccessList控制对Object的读写访问

### 疑问澄清

#### 技术实现疑问
1. **Object ID生成方式**: Object的ID是否需要特殊的生成规则，还是直接使用Hash？
2. **Object元数据存储**: Object的owner、contract等元数据如何存储？是否与Object数据分开存储？
3. **Object转移机制**: Object转移是否需要原子性操作？如何保证转移过程中的一致性？
4. **AccessList扩展**: 现有的AccessList.Reads/Writes字段是否需要扩展以支持Object的读写权限？

#### 架构设计疑问
1. **存储结构**: Object数据是否仍然存储在现有的LSM-Tree中，还是需要新的存储结构？
2. **缓存策略**: Object的元数据是否需要特殊的缓存策略？
3. **事务支持**: Object操作是否需要事务支持？
4. **性能考虑**: Object的创建、转移、修改操作对性能的影响如何控制？

#### 接口设计疑问
1. **API设计**: 是否需要为Object操作提供专门的API接口？
2. **错误处理**: Object操作失败时的错误处理策略？
3. **并发控制**: 多个合约同时操作同一个Object时的并发控制？

#### 集成疑问
1. **与现有代码集成**: 如何最小化对现有代码的影响？
2. **向后兼容**: 如何保证与现有AccessList机制的兼容性？
3. **测试策略**: 如何测试Object功能的正确性？

## 智能决策策略

基于现有项目内容和行业知识，我倾向于以下决策：

### 技术决策
1. **Object ID**: 使用Hash类型作为Object ID，保持与现有系统的一致性
2. **存储结构**: Object数据存储在现有LSM-Tree中，使用特殊前缀区分
3. **元数据存储**: Object元数据与数据分开存储，便于查询和管理
4. **序列化**: 继续使用gob格式，保持一致性

### 架构决策
1. **模块化设计**: 在storage模块内创建专门的Object子模块
2. **接口扩展**: 扩展现有存储接口，增加Object相关方法
3. **缓存策略**: Object元数据使用专门的缓存策略
4. **事务支持**: 利用现有的存储机制保证原子性

### 用户确认的关键决策点
1. **权限模型**: 只有合约可以读写object；如果要将object添加为Writes，要求owner为交易的发起人或合约
2. **Object转移**: 由交易控制，只需要支持单个Object转移
3. **版本控制**: 不实现Object的版本控制
4. **生命周期**: Object默认生命周期有2年
5. **AccessList**: 直接使用现有的Hash数组
