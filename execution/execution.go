package execution

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/govm-net/chain/config"
	"github.com/govm-net/chain/governance"
	"github.com/govm-net/chain/storage"
	"github.com/govm-net/chain/types"
)

// Execution 执行引擎
type Execution struct {
	config *config.ExecutionConfig
	storage *storage.Storage
	governance *governance.DPOSGovernance

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 交易池
	txPool    map[types.Hash]*types.Transaction
	txPoolMu  sync.RWMutex
	maxTxPool int

	// 状态管理
	state     map[string][]byte
	stateMu   sync.RWMutex
	stateRoot types.Hash
}

// New 创建新的执行引擎
func New(cfg *config.ExecutionConfig, storage *storage.Storage, gov *governance.DPOSGovernance) (*Execution, error) {
	ctx, cancel := context.WithCancel(context.Background())

	execution := &Execution{
		config:     cfg,
		storage:    storage,
		governance: gov,
		ctx:        ctx,
		cancel:     cancel,
		txPool:     make(map[types.Hash]*types.Transaction),
		maxTxPool:  10000,
		state:      make(map[string][]byte),
	}

	return execution, nil
}

// Start 启动执行引擎
func (e *Execution) Start() error {
	log := slog.With("module", "execution")
	log.Info("starting execution engine")

	// 启动交易处理
	e.wg.Add(1)
	go e.transactionProcessor()

	log.Info("execution engine started")
	return nil
}

// Stop 停止执行引擎
func (e *Execution) Stop() {
	log := slog.With("module", "execution")
	log.Info("stopping execution engine")

	e.cancel()
	e.wg.Wait()

	log.Info("execution engine stopped")
}

// AddTransaction 添加交易到池
func (e *Execution) AddTransaction(tx *types.Transaction) error {
	e.txPoolMu.Lock()
	defer e.txPoolMu.Unlock()

	// 检查池大小
	if len(e.txPool) >= e.maxTxPool {
		return fmt.Errorf("transaction pool is full")
	}

	// 计算交易哈希
	txHash := tx.CalculateHash()
	e.txPool[txHash] = tx

	return nil
}

// ExecuteTransaction 执行单个交易
func (e *Execution) ExecuteTransaction(tx *types.Transaction) error {
	// 检查是否为治理交易
	if tx.IsGovernanceTransaction() {
		return e.executeGovernanceTransaction(tx)
	}

	// 执行普通交易
	return e.executeTransferTransaction(tx)
}

// executeGovernanceTransaction 执行治理交易
func (e *Execution) executeGovernanceTransaction(tx *types.Transaction) error {
	if e.governance == nil {
		return fmt.Errorf("governance module not initialized")
	}

	// 委托给治理模块处理
	return e.governance.ProcessTransaction(tx)
}

// executeTransferTransaction 执行转账交易
func (e *Execution) executeTransferTransaction(tx *types.Transaction) error {
	// 这里实现普通的转账逻辑
	// 暂时只是占位符
	return nil
}

// ExecuteBlock 执行区块
func (e *Execution) ExecuteBlock(block *types.Block) error {
	log := slog.With("module", "execution", "block_height", block.Header.Height)
	log.Info("executing block")

	// 执行区块中的所有交易
	for _, txHash := range block.Transactions {
		tx, exists := e.txPool[txHash]
		if !exists {
			log.Warn("transaction not found in pool", "tx_hash", txHash.String())
			continue
		}

		if err := e.ExecuteTransaction(tx); err != nil {
			log.Error("failed to execute transaction", "tx_hash", txHash.String(), "error", err)
			return fmt.Errorf("failed to execute transaction %s: %w", txHash.String(), err)
		}

		// 从池中移除已执行的交易
		e.txPoolMu.Lock()
		delete(e.txPool, txHash)
		e.txPoolMu.Unlock()
	}

	// 更新状态根
	e.updateStateRoot()

	log.Info("block executed successfully")
	return nil
}

// transactionProcessor 交易处理器
func (e *Execution) transactionProcessor() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			// 处理交易池中的交易
			e.processTransactionPool()
		}
	}
}

// processTransactionPool 处理交易池
func (e *Execution) processTransactionPool() {
	e.txPoolMu.RLock()
	txCount := len(e.txPool)
	e.txPoolMu.RUnlock()

	if txCount == 0 {
		return
	}

	// 这里可以实现批量处理逻辑
	// 暂时只是占位符
}

// updateStateRoot 更新状态根
func (e *Execution) updateStateRoot() {
	// 计算新的状态根
	// 这里暂时使用简单的哈希计算
	e.stateMu.Lock()
	defer e.stateMu.Unlock()

	// 合并所有状态数据
	var data []byte
	for key, value := range e.state {
		data = append(data, []byte(key)...)
		data = append(data, value...)
	}

	// 计算哈希
	e.stateRoot = e.calculateHash(data)
}

// calculateHash 计算哈希
func (e *Execution) calculateHash(data []byte) types.Hash {
	// 这里使用简单的哈希计算
	// 实际应该使用更安全的哈希算法
	var hash types.Hash
	copy(hash[:], data[:32])
	return hash
}

// GetStateRoot 获取状态根
func (e *Execution) GetStateRoot() types.Hash {
	e.stateMu.RLock()
	defer e.stateMu.RUnlock()
	return e.stateRoot
}

// GetState 获取状态
func (e *Execution) GetState(key string) ([]byte, bool) {
	e.stateMu.RLock()
	defer e.stateMu.RUnlock()
	value, exists := e.state[key]
	return value, exists
}

// SetState 设置状态
func (e *Execution) SetState(key string, value []byte) {
	e.stateMu.Lock()
	defer e.stateMu.Unlock()
	e.state[key] = value
}
