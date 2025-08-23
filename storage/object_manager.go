package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/govm-net/chain/types"
)

// ObjectManager 负责Object的创建、删除、转移等生命周期管理
type ObjectManager struct {
	storage *Storage
	ctx     context.Context
}

type objectNewCount struct{}

// NewObjectManager 创建新的ObjectManager实例
func NewObjectManager(storage *Storage, ctx context.Context) *ObjectManager {
	var count int = 1
	ctx = context.WithValue(ctx, objectNewCount{}, count)
	return &ObjectManager{
		storage: storage,
		ctx:     ctx,
	}
}

// CreateObject 创建Object
func (om *ObjectManager) CreateObject(owner, contract []byte) (*types.Object, error) {
	// 创建新的Object
	object := types.NewObject(owner, contract)

	// 生成Object ID（基于owner、contract和tx hash）
	object.ID = om.generateObjectID(owner, contract)

	// 存储Object元数据
	err := om.storeObjectMeta(object)
	if err != nil {
		return nil, fmt.Errorf("failed to store object meta: %w", err)
	}

	// 创建索引
	err = om.createIndexes(object)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	// 创建生命周期索引
	err = om.createLifecycleIndex(object.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create lifecycle index: %w", err)
	}

	return object, nil
}

// GetObject 获取Object
func (om *ObjectManager) GetObject(id types.Hash) (*types.Object, error) {
	// 构建Object元数据键
	metaKey := om.buildObjectMetaKey(id)

	// 获取Object元数据
	data, err := om.storage.Get(metaKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get object meta: %w", err)
	}
	if data == nil {
		return nil, fmt.Errorf("object not found: %s", id.String())
	}

	// 反序列化Object
	object := &types.Object{}
	err = object.Deserialize(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object: %w", err)
	}

	return object, nil
}

// TransferObject 转移Object
func (om *ObjectManager) TransferObject(id types.Hash, newOwner []byte) error {
	// 获取Object
	object, err := om.GetObject(id)
	if err != nil {
		return err
	}

	// 检查Object是否过期（由业务层判断）
	// 这里只提供获取过期时间的接口，业务层自行判断

	// 删除旧索引
	err = om.deleteIndexes(object)
	if err != nil {
		return fmt.Errorf("failed to delete old indexes: %w", err)
	}

	// 更新owner
	object.Owner = newOwner

	// 更新Object元数据
	err = om.storeObjectMeta(object)
	if err != nil {
		return fmt.Errorf("failed to update object meta: %w", err)
	}

	// 创建新索引
	err = om.createIndexes(object)
	if err != nil {
		return fmt.Errorf("failed to create new indexes: %w", err)
	}

	return nil
}

// DeleteObject 删除Object
func (om *ObjectManager) DeleteObject(id types.Hash) error {
	// 获取Object
	object, err := om.GetObject(id)
	if err != nil {
		return err
	}

	// 删除索引
	err = om.deleteIndexes(object)
	if err != nil {
		return fmt.Errorf("failed to delete indexes: %w", err)
	}

	// 删除Object元数据
	metaKey := om.buildObjectMetaKey(id)
	err = om.storage.Delete(metaKey)
	if err != nil {
		return fmt.Errorf("failed to delete object meta: %w", err)
	}

	// 删除生命周期索引
	err = om.deleteLifecycleIndex(id)
	if err != nil {
		return fmt.Errorf("failed to delete lifecycle index: %w", err)
	}

	// TODO: 删除Object的所有数据（需要遍历所有以objectID为前缀的键）

	return nil
}

// GetObjectLifecycle 获取Object的生命周期
func (om *ObjectManager) GetObjectLifecycle(id types.Hash) (int64, error) {
	return om.getLifecycleData(id)
}

// GetObjectStorage 获取Object的Storage接口
func (om *ObjectManager) GetObjectStorage(id types.Hash) (*ObjectStorage, error) {
	// 检查Object是否存在
	_, err := om.GetObject(id)
	if err != nil {
		return nil, err
	}

	return NewObjectStorage(om.storage, id), nil
}

// generateObjectID 生成Object ID
func (om *ObjectManager) generateObjectID(owner, contract []byte) types.Hash {
	data := append(owner, contract...)
	txHash := om.ctx.Value(types.CtxTxHash{})

	count := om.ctx.Value(objectNewCount{}).(int)
	om.ctx = context.WithValue(om.ctx, objectNewCount{}, count+1)
	hashStr := fmt.Sprintf("%s:%d", txHash, count)
	data = append(data, []byte(hashStr)...)

	hash := sha256.Sum256(data)
	return types.NewHash(hash[:])
}

// storeObjectMeta 存储Object元数据
func (om *ObjectManager) storeObjectMeta(object *types.Object) error {
	metaKey := om.buildObjectMetaKey(object.ID)
	data, err := object.Serialize()
	if err != nil {
		return err
	}
	return om.storage.Set(metaKey, data)
}

// buildObjectMetaKey 构建Object元数据键
func (om *ObjectManager) buildObjectMetaKey(id types.Hash) []byte {
	return []byte(fmt.Sprintf("object:meta:%s", id.String()))
}

// createIndexes 创建索引
func (om *ObjectManager) createIndexes(object *types.Object) error {
	// 创建owner索引
	ownerKey := om.buildOwnerIndexKey(object.Owner, object.ID)
	err := om.storage.Set(ownerKey, []byte{1}) // 使用1表示存在
	if err != nil {
		return err
	}

	// 创建contract索引
	contractKey := om.buildContractIndexKey(object.Contract, object.ID)
	return om.storage.Set(contractKey, []byte{1})
}

// deleteIndexes 删除索引
func (om *ObjectManager) deleteIndexes(object *types.Object) error {
	// 删除owner索引
	ownerKey := om.buildOwnerIndexKey(object.Owner, object.ID)
	err := om.storage.Delete(ownerKey)
	if err != nil {
		return err
	}

	// 删除contract索引
	contractKey := om.buildContractIndexKey(object.Contract, object.ID)
	return om.storage.Delete(contractKey)
}

// buildOwnerIndexKey 构建owner索引键
func (om *ObjectManager) buildOwnerIndexKey(owner []byte, id types.Hash) []byte {
	return []byte(fmt.Sprintf("object:index:owner:%s:%s", id.String(), string(owner)))
}

// buildContractIndexKey 构建contract索引键
func (om *ObjectManager) buildContractIndexKey(contract []byte, id types.Hash) []byte {
	return []byte(fmt.Sprintf("object:index:contract:%s:%s", string(contract), id.String()))
}

// createLifecycleIndex 创建生命周期索引
func (om *ObjectManager) createLifecycleIndex(objectID types.Hash) error {
	// 从ctx中获取block time
	blockTime := om.getBlockTime()

	// 构建生命周期索引键
	lifecycleKey := om.buildLifecycleIndexKey(objectID)
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, blockTime)

	return om.storage.Set(lifecycleKey, buf.Bytes())
}

// deleteLifecycleIndex 删除生命周期索引
func (om *ObjectManager) deleteLifecycleIndex(objectID types.Hash) error {
	// 构建生命周期索引键
	lifecycleKey := om.buildLifecycleIndexKey(objectID)

	// 删除生命周期索引
	return om.storage.Delete(lifecycleKey)
}

// getLifecycleData 获取生命周期数据
func (om *ObjectManager) getLifecycleData(objectID types.Hash) (int64, error) {
	// TODO: 实现从生命周期索引中获取数据的逻辑
	// 由于当前没有范围查询功能，这里简化实现
	// 实际实现时需要遍历生命周期索引找到对应的数据
	lifecycleKey := om.buildLifecycleIndexKey(objectID)
	data, err := om.storage.Get(lifecycleKey)
	if err != nil {
		return 0, err
	}
	if data == nil {
		return 0, fmt.Errorf("lifecycle data not found: %s", objectID.String())
	}
	var lifecycle int64
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &lifecycle)
	return lifecycle, nil
}

// buildLifecycleIndexKey 构建生命周期索引键
func (om *ObjectManager) buildLifecycleIndexKey(objectID types.Hash) []byte {
	return []byte(fmt.Sprintf("object:index:lifecycle:%s", objectID.String()))
}

// getBlockTime 从ctx中获取block time
func (om *ObjectManager) getBlockTime() int64 {
	if blockTime, ok := om.ctx.Value("block_time").(int64); ok {
		return blockTime
	}
	// 如果没有block time，使用当前时间
	return time.Now().Unix()
}

// ExtendObjectLifecycle 延长Object生命周期
func (om *ObjectManager) ExtendObjectLifecycle(objectID types.Hash, additionalSeconds int64) error {
	// 获取当前生命周期数据
	currentLifecycle, err := om.getLifecycleData(objectID)
	if err != nil {
		return err
	}
	blockTime := om.getBlockTime()
	if currentLifecycle < blockTime {
		return fmt.Errorf("object lifecycle is expired")
	}

	if additionalSeconds <= 0 {
		currentLifecycle = blockTime
	} else {
		currentLifecycle += additionalSeconds
	}

	// 创建新的生命周期索引
	lifecycleKey := om.buildLifecycleIndexKey(objectID)

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, currentLifecycle)

	return om.storage.Set(lifecycleKey, buf.Bytes())
}
