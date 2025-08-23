package chain

import (
	"context"
	"fmt"
	"sync"

	"github.com/govm-net/chain/config"
	"github.com/govm-net/chain/consensus"
	"github.com/govm-net/chain/execution"
	"github.com/govm-net/chain/governance"
	"github.com/govm-net/chain/network"
	"github.com/govm-net/chain/storage"
)

// Chain 区块链核心结构
type Chain struct {
	config    *config.Config
	network   *network.Network
	consensus *consensus.Consensus
	execution *execution.Execution
	storage   *storage.Storage

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New 创建新的区块链实例
func New(cfg *config.Config) (*Chain, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建存储实例
	storage, err := storage.New(cfg.Storage)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// 创建治理模块
	governance, err := governance.NewDPOSGovernance(&cfg.Governance, storage)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create governance module: %w", err)
	}

	// 创建执行引擎
	execution, err := execution.New(&cfg.Execution, storage, governance)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create execution engine: %w", err)
	}

	// 创建共识引擎
	consensus, err := consensus.New(cfg.Consensus, execution, storage)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create consensus engine: %w", err)
	}

	// 创建网络层
	network, err := network.New(cfg.Network)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create network layer: %w", err)
	}

	return &Chain{
		config:    cfg,
		network:   network,
		consensus: consensus,
		execution: execution,
		storage:   storage,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Start 启动区块链
func (c *Chain) Start() error {
	// 启动存储
	if err := c.storage.Start(); err != nil {
		return fmt.Errorf("启动存储失败: %w", err)
	}

	// 启动执行引擎
	if err := c.execution.Start(); err != nil {
		return fmt.Errorf("启动执行引擎失败: %w", err)
	}

	// 启动共识引擎
	if err := c.consensus.Start(); err != nil {
		return fmt.Errorf("启动共识引擎失败: %w", err)
	}

	// 启动网络层
	if err := c.network.Start(); err != nil {
		return fmt.Errorf("启动网络层失败: %w", err)
	}

	return nil
}

// Stop 停止区块链
func (c *Chain) Stop() {
	c.cancel()

	// 停止网络层
	c.network.Stop()

	// 停止共识引擎
	c.consensus.Stop()

	// 停止执行引擎
	c.execution.Stop()

	// 停止存储
	c.storage.Stop()

	// 等待所有goroutine结束
	c.wg.Wait()
}
