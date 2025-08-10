package execution

import (
	"sync"

	"github.com/govm-net/chain/config"
	"github.com/govm-net/chain/storage"
	"github.com/govm-net/chain/types"
)

// Execution 执行引擎
type Execution struct {
	config  *config.ExecutionConfig
	storage *storage.Storage
	mu      sync.RWMutex
}

// New 创建新的执行引擎
func New(cfg config.ExecutionConfig, storage *storage.Storage) (*Execution, error) {
	return &Execution{
		config:  &cfg,
		storage: storage,
	}, nil
}

// Start 启动执行引擎
func (e *Execution) Start() error {
	// TODO: 启动并行执行引擎
	return nil
}

// Stop 停止执行引擎
func (e *Execution) Stop() {
	// TODO: 停止执行引擎
}

// ExecuteTransaction 执行交易
func (e *Execution) ExecuteTransaction(tx *types.Transaction) error {
	// TODO: 实现交易执行
	return nil
}

// ExecuteBlock 执行区块
func (e *Execution) ExecuteBlock(block *types.Block) error {
	// TODO: 实现区块执行
	return nil
}
