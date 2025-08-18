package storage

import (
	"fmt"

	"github.com/govm-net/chain/types"
)

// ObjectKeyBuilder Object存储键构建器
type ObjectKeyBuilder struct{}

// NewObjectKeyBuilder 创建新的ObjectKeyBuilder实例
func NewObjectKeyBuilder() *ObjectKeyBuilder {
	return &ObjectKeyBuilder{}
}

// BuildObjectMetaKey 构建Object元数据键
// 格式: object:meta:objectID
func (okb *ObjectKeyBuilder) BuildObjectMetaKey(objectID types.Hash) []byte {
	return []byte(fmt.Sprintf("object:meta:%s", objectID.String()))
}

// BuildObjectDataKey 构建Object数据键
// 格式: object:data:objectID:key
func (okb *ObjectKeyBuilder) BuildObjectDataKey(objectID types.Hash, key []byte) []byte {
	prefix := fmt.Sprintf("object:data:%s:", objectID.String())
	return append([]byte(prefix), key...)
}

// BuildOwnerIndexKey 构建owner索引键
// 格式: object:index:owner:ownerAddress:objectID
func (okb *ObjectKeyBuilder) BuildOwnerIndexKey(owner []byte, objectID types.Hash) []byte {
	return []byte(fmt.Sprintf("object:index:owner:%s:%s", string(owner), objectID.String()))
}

// BuildContractIndexKey 构建contract索引键
// 格式: object:index:contract:contractAddress:objectID
func (okb *ObjectKeyBuilder) BuildContractIndexKey(contract []byte, objectID types.Hash) []byte {
	return []byte(fmt.Sprintf("object:index:contract:%s:%s", string(contract), objectID.String()))
}

// BuildExpirationIndexKey 构建过期时间索引键
// 格式: object:index:expiration:timestamp:objectID
func (okb *ObjectKeyBuilder) BuildExpirationIndexKey(expiresAt int64, objectID types.Hash) []byte {
	return []byte(fmt.Sprintf("object:index:expiration:%d:%s", expiresAt, objectID.String()))
}

// ParseObjectDataKey 解析Object数据键，提取key部分
func (okb *ObjectKeyBuilder) ParseObjectDataKey(dataKey []byte) (types.Hash, []byte, error) {
	// 格式: object:data:objectID:key
	keyStr := string(dataKey)
	prefix := "object:data:"

	if len(keyStr) <= len(prefix) {
		return types.Hash{}, nil, fmt.Errorf("invalid object data key format")
	}

	// 移除前缀
	remaining := keyStr[len(prefix):]

	// 查找第一个冒号分隔符
	colonIndex := -1
	for i, char := range remaining {
		if char == ':' {
			colonIndex = i
			break
		}
	}

	if colonIndex == -1 {
		return types.Hash{}, nil, fmt.Errorf("invalid object data key format: missing separator")
	}

	// 提取objectID和key
	_ = remaining[:colonIndex] // objectIDStr，暂时未使用
	keyStr = remaining[colonIndex+1:]

	// 解析objectID - 这里简化处理，实际项目中需要实现从字符串解析Hash的方法
	// TODO: 实现从字符串解析Hash的方法
	var objectID types.Hash
	// 临时使用空Hash，实际实现时需要从objectIDStr解析

	return objectID, []byte(keyStr), nil
}

// IsObjectMetaKey 判断是否为Object元数据键
func (okb *ObjectKeyBuilder) IsObjectMetaKey(key []byte) bool {
	keyStr := string(key)
	return len(keyStr) > 12 && keyStr[:12] == "object:meta:"
}

// IsObjectDataKey 判断是否为Object数据键
func (okb *ObjectKeyBuilder) IsObjectDataKey(key []byte) bool {
	keyStr := string(key)
	return len(keyStr) > 12 && keyStr[:12] == "object:data:"
}

// IsObjectIndexKey 判断是否为Object索引键
func (okb *ObjectKeyBuilder) IsObjectIndexKey(key []byte) bool {
	keyStr := string(key)
	return len(keyStr) > 13 && keyStr[:13] == "object:index:"
}
