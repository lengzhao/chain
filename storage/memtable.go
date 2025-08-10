package storage

import (
	"sync"
)

// MemTable 内存表
type MemTable struct {
	data map[string][]byte
	mu   sync.RWMutex
	size int
}

// NewMemTable 创建新的内存表
func NewMemTable() *MemTable {
	return &MemTable{
		data: make(map[string][]byte),
	}
}

// Get 获取数据
func (m *MemTable) Get(key []byte) []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[string(key)]
}

// Set 设置数据
func (m *MemTable) Set(key, value []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	keyStr := string(key)
	oldValue := m.data[keyStr]
	
	// 更新大小
	m.size -= len(oldValue)
	m.size += len(value)
	
	m.data[keyStr] = value
}

// Delete 删除数据
func (m *MemTable) Delete(key []byte) {
	m.Set(key, nil)
}

// Size 获取内存表大小
func (m *MemTable) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.size
}

// Clear 清空内存表
func (m *MemTable) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string][]byte)
	m.size = 0
} 