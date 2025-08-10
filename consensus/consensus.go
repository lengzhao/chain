package consensus

import (
	"sync"

	"github.com/govm-net/chain/config"
	"github.com/govm-net/chain/execution"
	"github.com/govm-net/chain/storage"
	"github.com/govm-net/chain/types"
)

// Consensus 共识接口
type Consensus struct {
	config    *config.ConsensusConfig
	execution *execution.Execution
	storage   *storage.Storage
	mu        sync.RWMutex
}

// New 创建新的共识实例
func New(cfg config.ConsensusConfig, execution *execution.Execution, storage *storage.Storage) (*Consensus, error) {
	return &Consensus{
		config:    &cfg,
		execution: execution,
		storage:   storage,
	}, nil
}

// Start 启动共识
func (c *Consensus) Start() error {
	// TODO: 启动PBFT共识
	return nil
}

// Stop 停止共识
func (c *Consensus) Stop() {
	// TODO: 停止共识
}

// ProposeBlock 提议新区块
func (c *Consensus) ProposeBlock(transactions []*types.Transaction) (*types.Block, error) {
	// TODO: 实现区块提议
	return &types.Block{}, nil
}

// ValidateBlock 验证区块
func (c *Consensus) ValidateBlock(block *types.Block) error {
	// TODO: 实现区块验证
	return nil
}
