package storage

import (
	"testing"

	"github.com/govm-net/chain/config"
	"github.com/govm-net/chain/types"
)

func TestNewObjectStorage(t *testing.T) {
	// 创建测试存储
	cfg := config.StorageConfig{
		DataDir:     "./test_data",
		MaxSize:     1024 * 1024, // 1MB
		CacheSize:   100,
		Compression: false,
	}

	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Stop()

	objectID := types.NewHash([]byte("test_object_id_123456789012345678901234567890"))
	objectStorage := NewObjectStorage(storage, objectID)

	if objectStorage == nil {
		t.Fatal("NewObjectStorage returned nil")
	}

	if objectStorage.Storage != storage {
		t.Error("Storage reference mismatch")
	}

	if objectStorage.objectID != objectID {
		t.Errorf("ObjectID mismatch: got %s, want %s", objectStorage.objectID.String(), objectID.String())
	}
}

func TestObjectStorageGetSetDelete(t *testing.T) {
	// 创建测试存储
	cfg := config.StorageConfig{
		DataDir:     "./test_data",
		MaxSize:     1024 * 1024, // 1MB
		CacheSize:   100,
		Compression: false,
	}

	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Stop()

	objectID := types.NewHash([]byte("test_object_id_123456789012345678901234567890"))
	objectStorage := NewObjectStorage(storage, objectID)

	// 测试数据
	key := []byte("test_key")
	value := []byte("test_value")

	// 测试Set
	err = objectStorage.Set(key, value)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// 测试Get
	retrievedValue, err := objectStorage.Get(key)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if string(retrievedValue) != string(value) {
		t.Errorf("Value mismatch: got %s, want %s", string(retrievedValue), string(value))
	}

	// 测试Get不存在的键
	nonExistentKey := []byte("non_existent_key")
	retrievedValue, err = objectStorage.Get(nonExistentKey)
	if err != nil {
		t.Fatalf("Failed to get non-existent key: %v", err)
	}

	if retrievedValue != nil {
		t.Errorf("Expected nil for non-existent key, got %v", retrievedValue)
	}

	// 测试Delete
	err = objectStorage.Delete(key)
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	// 验证删除后获取返回nil
	retrievedValue, err = objectStorage.Get(key)
	if err != nil {
		t.Fatalf("Failed to get deleted key: %v", err)
	}

	if retrievedValue != nil {
		t.Errorf("Expected nil for deleted key, got %v", retrievedValue)
	}
}

func TestObjectStorageDataIsolation(t *testing.T) {
	// 创建测试存储
	cfg := config.StorageConfig{
		DataDir:     "./test_data",
		MaxSize:     1024 * 1024, // 1MB
		CacheSize:   100,
		Compression: false,
	}

	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Stop()

	// 创建两个不同的ObjectStorage实例
	objectID1 := types.NewHash([]byte("object_id_1_123456789012345678901234567890"))
	objectID2 := types.NewHash([]byte("object_id_2_123456789012345678901234567890"))

	objectStorage1 := NewObjectStorage(storage, objectID1)
	objectStorage2 := NewObjectStorage(storage, objectID2)

	key := []byte("same_key")
	value1 := []byte("value_for_object1")
	value2 := []byte("value_for_object2")

	// 在两个ObjectStorage中设置相同的键但不同的值
	err = objectStorage1.Set(key, value1)
	if err != nil {
		t.Fatalf("Failed to set value in objectStorage1: %v", err)
	}

	err = objectStorage2.Set(key, value2)
	if err != nil {
		t.Fatalf("Failed to set value in objectStorage2: %v", err)
	}

	// 验证数据隔离
	retrievedValue1, err := objectStorage1.Get(key)
	if err != nil {
		t.Fatalf("Failed to get value from objectStorage1: %v", err)
	}

	if string(retrievedValue1) != string(value1) {
		t.Errorf("Value mismatch in objectStorage1: got %s, want %s", string(retrievedValue1), string(value1))
	}

	retrievedValue2, err := objectStorage2.Get(key)
	if err != nil {
		t.Fatalf("Failed to get value from objectStorage2: %v", err)
	}

	if string(retrievedValue2) != string(value2) {
		t.Errorf("Value mismatch in objectStorage2: got %s, want %s", string(retrievedValue2), string(value2))
	}
}

func TestObjectStorageGetObjectID(t *testing.T) {
	// 创建测试存储
	cfg := config.StorageConfig{
		DataDir:     "./test_data",
		MaxSize:     1024 * 1024, // 1MB
		CacheSize:   100,
		Compression: false,
	}

	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Stop()

	objectID := types.NewHash([]byte("test_object_id_123456789012345678901234567890"))
	objectStorage := NewObjectStorage(storage, objectID)

	retrievedObjectID := objectStorage.GetObjectID()
	if retrievedObjectID != objectID {
		t.Errorf("ObjectID mismatch: got %s, want %s", retrievedObjectID.String(), objectID.String())
	}
}

func TestObjectStorageBuildObjectDataKey(t *testing.T) {
	// 创建测试存储
	cfg := config.StorageConfig{
		DataDir:     "./test_data",
		MaxSize:     1024 * 1024, // 1MB
		CacheSize:   100,
		Compression: false,
	}

	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Stop()

	objectID := types.NewHash([]byte("test_object_id_123456789012345678901234567890"))
	objectStorage := NewObjectStorage(storage, objectID)

	key := []byte("test_key")
	objectKey := objectStorage.buildObjectDataKey(key)

	expectedPrefix := "object:data:" + objectID.String() + ":"
	if !contains(objectKey, []byte(expectedPrefix)) {
		t.Errorf("Object key should contain prefix %s, got %s", expectedPrefix, string(objectKey))
	}

	if !contains(objectKey, key) {
		t.Errorf("Object key should contain original key %s, got %s", string(key), string(objectKey))
	}
}

// 辅助函数：检查字节切片是否包含子切片
func contains(slice, subslice []byte) bool {
	for i := 0; i <= len(slice)-len(subslice); i++ {
		if bytesEqual(slice[i:i+len(subslice)], subslice) {
			return true
		}
	}
	return false
}

// 辅助函数：比较两个字节切片是否相等
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
