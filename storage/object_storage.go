package storage

import (
	"fmt"

	"github.com/govm-net/chain/types"
)

// ObjectStorage 继承Storage功能，提供Object语义的存储操作
type ObjectStorage struct {
	*Storage
	objectID types.Hash
}

// NewObjectStorage 创建新的ObjectStorage实例
func NewObjectStorage(storage *Storage, objectID types.Hash) *ObjectStorage {
	return &ObjectStorage{
		Storage:  storage,
		objectID: objectID,
	}
}

// Get 获取Object内的数据
func (os *ObjectStorage) Get(key []byte) ([]byte, error) {
	// 构建Object数据键：object:data:objectID:key
	objectKey := os.buildObjectDataKey(key)
	return os.Storage.Get(objectKey)
}

// Set 设置Object内的数据
func (os *ObjectStorage) Set(key, value []byte) error {
	// 构建Object数据键：object:data:objectID:key
	objectKey := os.buildObjectDataKey(key)
	return os.Storage.Set(objectKey, value)
}

// Delete 删除Object内的数据
func (os *ObjectStorage) Delete(key []byte) error {
	// 构建Object数据键：object:data:objectID:key
	objectKey := os.buildObjectDataKey(key)
	return os.Storage.Delete(objectKey)
}

// buildObjectDataKey 构建Object数据存储键
func (os *ObjectStorage) buildObjectDataKey(key []byte) []byte {
	prefix := fmt.Sprintf("object:data:%s:", os.objectID.String())
	return append([]byte(prefix), key...)
}

// GetObjectID 获取Object ID
func (os *ObjectStorage) GetObjectID() types.Hash {
	return os.objectID
}
