package storage

import (
	"container/list"
	"sync"
)

// Cache LRU缓存
type Cache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex
}

// cacheEntry 缓存条目
type cacheEntry struct {
	key   string
	value []byte
}

// NewCache 创建新的缓存
func NewCache(capacity int) *Cache {
	return &Cache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Get 获取缓存数据
func (c *Cache) Get(key []byte) []byte {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, exists := c.cache[string(key)]; exists {
		c.list.MoveToFront(element)
		return element.Value.(*cacheEntry).value
	}
	return nil
}

// Set 设置缓存数据
func (c *Cache) Set(key, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keyStr := string(key)
	if element, exists := c.cache[keyStr]; exists {
		c.list.MoveToFront(element)
		element.Value.(*cacheEntry).value = value
		return
	}

	entry := &cacheEntry{key: keyStr, value: value}
	element := c.list.PushFront(entry)
	c.cache[keyStr] = element

	if c.list.Len() > c.capacity {
		c.evict()
	}
}

// Delete 删除缓存数据
func (c *Cache) Delete(key []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keyStr := string(key)
	if element, exists := c.cache[keyStr]; exists {
		c.list.Remove(element)
		delete(c.cache, keyStr)
	}
}

// evict 淘汰最久未使用的数据
func (c *Cache) evict() {
	element := c.list.Back()
	if element != nil {
		c.list.Remove(element)
		entry := element.Value.(*cacheEntry)
		delete(c.cache, entry.key)
	}
} 