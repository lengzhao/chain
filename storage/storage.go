package storage

import (
	"fmt"
	"sync"

	"github.com/govm-net/chain/config"
	"github.com/govm-net/chain/types"
)

// Storage 存储接口
type Storage struct {
	config *config.StorageConfig
	db     *LSMTree
	cache  *Cache
	mu     sync.RWMutex
}

// New 创建新的存储实例
func New(cfg config.StorageConfig) (*Storage, error) {
	// 创建LSM-Tree存储引擎
	db, err := NewLSMTree(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("创建LSM-Tree失败: %w", err)
	}

	// 创建缓存层
	cache := NewCache(cfg.CacheSize)

	return &Storage{
		config: &cfg,
		db:     db,
		cache:  cache,
	}, nil
}

// Start 启动存储
func (s *Storage) Start() error {
	return s.db.Start()
}

// Stop 停止存储
func (s *Storage) Stop() {
	s.db.Stop()
}

// Get 获取数据
func (s *Storage) Get(key []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 先查缓存
	if data := s.cache.Get(key); data != nil {
		return data, nil
	}

	// 查数据库
	data, err := s.db.Get(key)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	if data != nil {
		s.cache.Set(key, data)
	}

	return data, nil
}

// Set 设置数据
func (s *Storage) Set(key, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 更新缓存
	s.cache.Set(key, value)

	// 写入数据库
	return s.db.Set(key, value)
}

// Delete 删除数据
func (s *Storage) Delete(key []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 删除缓存
	s.cache.Delete(key)

	// 删除数据库
	return s.db.Delete(key)
}

// GetBlock 获取区块
func (s *Storage) GetBlock(height uint64) (*types.Block, error) {
	key := []byte(fmt.Sprintf("block:%d", height))
	data, err := s.Get(key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("区块不存在: %d", height)
	}

	// TODO: 反序列化区块
	return &types.Block{}, nil
}

// SetBlock 设置区块
func (s *Storage) SetBlock(block *types.Block) error {
	key := []byte(fmt.Sprintf("block:%d", block.Header.Height))
	// TODO: 序列化区块
	data := []byte{}
	return s.Set(key, data)
}
