package storage

import (
	"fmt"
	"os"
	"sync"
)

// LSMTree LSM-Tree存储引擎
type LSMTree struct {
	dataDir  string
	memTable *MemTable
	sstables []*SSTable
	mu       sync.RWMutex
}

// NewLSMTree 创建新的LSM-Tree实例
func NewLSMTree(dataDir string) (*LSMTree, error) {
	// 创建数据目录
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 创建内存表
	memTable := NewMemTable()

	// 加载现有的SSTable文件
	sstables, err := loadSSTables(dataDir)
	if err != nil {
		return nil, fmt.Errorf("加载SSTable失败: %w", err)
	}

	return &LSMTree{
		dataDir:  dataDir,
		memTable: memTable,
		sstables: sstables,
	}, nil
}

// Start 启动LSM-Tree
func (l *LSMTree) Start() error {
	// TODO: 启动后台合并任务
	return nil
}

// Stop 停止LSM-Tree
func (l *LSMTree) Stop() {
	// TODO: 停止后台任务
}

// Get 获取数据
func (l *LSMTree) Get(key []byte) ([]byte, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// 先查内存表
	if value := l.memTable.Get(key); value != nil {
		return value, nil
	}

	// 查SSTable
	for _, sstable := range l.sstables {
		if value := sstable.Get(key); value != nil {
			return value, nil
		}
	}

	return nil, nil
}

// Set 设置数据
func (l *LSMTree) Set(key, value []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 写入内存表
	l.memTable.Set(key, value)

	// 检查内存表是否需要刷新
	if l.memTable.Size() > 64*1024*1024 { // 64MB
		if err := l.flushMemTable(); err != nil {
			return fmt.Errorf("刷新内存表失败: %w", err)
		}
	}

	return nil
}

// Delete 删除数据
func (l *LSMTree) Delete(key []byte) error {
	return l.Set(key, nil) // 标记删除
}

// flushMemTable 刷新内存表到SSTable
func (l *LSMTree) flushMemTable() error {
	// TODO: 实现内存表刷新逻辑
	return nil
}

// loadSSTables 加载现有的SSTable文件
func loadSSTables(dataDir string) ([]*SSTable, error) {
	// TODO: 实现SSTable加载逻辑
	return []*SSTable{}, nil
}
