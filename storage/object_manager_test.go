package storage

import (
	"context"
	"testing"

	"github.com/govm-net/chain/config"
	"github.com/govm-net/chain/types"
)

func TestNewObjectManager(t *testing.T) {
	cfg := config.StorageConfig{
		DataDir:     "./test_data",
		MaxSize:     1024 * 1024,
		CacheSize:   100,
		Compression: false,
	}

	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Stop()

	ctx := context.Background()
	objectManager := NewObjectManager(storage, ctx)

	if objectManager == nil {
		t.Fatal("NewObjectManager returned nil")
	}

	if objectManager.storage != storage {
		t.Error("Storage reference mismatch")
	}
}

func TestObjectManagerCreateObject(t *testing.T) {
	cfg := config.StorageConfig{
		DataDir:     "./test_data",
		MaxSize:     1024 * 1024,
		CacheSize:   100,
		Compression: false,
	}

	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Stop()

	ctx := context.WithValue(context.Background(), types.CtxTxHash{}, types.NewHash([]byte("test_tx_hash_123456789012345678901234567890")))
	objectManager := NewObjectManager(storage, ctx)

	owner := []byte("test_owner")
	contract := []byte("test_contract")

	object, err := objectManager.CreateObject(owner, contract)
	if err != nil {
		t.Fatalf("Failed to create object: %v", err)
	}

	if object == nil {
		t.Fatal("CreateObject returned nil")
	}

	if string(object.Owner) != string(owner) {
		t.Errorf("Owner mismatch: got %s, want %s", string(object.Owner), string(owner))
	}

	var emptyHash types.Hash
	if object.ID == emptyHash {
		t.Error("Object ID should be set after creation")
	}
}

func TestObjectManagerGetObject(t *testing.T) {
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

	ctx := context.WithValue(context.Background(), types.CtxTxHash{}, types.NewHash([]byte("test_tx_hash_123456789012345678901234567890")))
	objectManager := NewObjectManager(storage, ctx)

	owner := []byte("test_owner")
	contract := []byte("test_contract")

	// 创建Object
	createdObject, err := objectManager.CreateObject(owner, contract)
	if err != nil {
		t.Fatalf("Failed to create object: %v", err)
	}

	// 获取Object
	retrievedObject, err := objectManager.GetObject(createdObject.ID)
	if err != nil {
		t.Fatalf("Failed to get object: %v", err)
	}

	if retrievedObject == nil {
		t.Fatal("GetObject returned nil")
	}

	if retrievedObject.ID != createdObject.ID {
		t.Errorf("Object ID mismatch: got %s, want %s", retrievedObject.ID.String(), createdObject.ID.String())
	}

	if string(retrievedObject.Owner) != string(owner) {
		t.Errorf("Owner mismatch: got %s, want %s", string(retrievedObject.Owner), string(owner))
	}

	if string(retrievedObject.Contract) != string(contract) {
		t.Errorf("Contract mismatch: got %s, want %s", string(retrievedObject.Contract), string(contract))
	}
}

func TestObjectManagerGetObjectNotFound(t *testing.T) {
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

	ctx := context.Background()
	objectManager := NewObjectManager(storage, ctx)

	// 尝试获取不存在的Object
	nonExistentID := types.NewHash([]byte("non_existent_id_123456789012345678901234567890"))
	_, err = objectManager.GetObject(nonExistentID)

	if err == nil {
		t.Error("Expected error when getting non-existent object")
	}
}

func TestObjectManagerTransferObject(t *testing.T) {
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

	ctx := context.WithValue(context.Background(), types.CtxTxHash{}, types.NewHash([]byte("test_tx_hash_123456789012345678901234567890")))
	objectManager := NewObjectManager(storage, ctx)

	owner := []byte("original_owner")
	contract := []byte("test_contract")

	// 创建Object
	object, err := objectManager.CreateObject(owner, contract)
	if err != nil {
		t.Fatalf("Failed to create object: %v", err)
	}

	newOwner := []byte("new_owner")

	// 转移Object
	err = objectManager.TransferObject(object.ID, newOwner)
	if err != nil {
		t.Fatalf("Failed to transfer object: %v", err)
	}

	// 验证转移结果
	transferredObject, err := objectManager.GetObject(object.ID)
	if err != nil {
		t.Fatalf("Failed to get transferred object: %v", err)
	}

	if string(transferredObject.Owner) != string(newOwner) {
		t.Errorf("Owner mismatch after transfer: got %s, want %s", string(transferredObject.Owner), string(newOwner))
	}
}

func TestObjectManagerDeleteObject(t *testing.T) {
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

	ctx := context.WithValue(context.Background(), types.CtxTxHash{}, types.NewHash([]byte("test_tx_hash_123456789012345678901234567890")))
	objectManager := NewObjectManager(storage, ctx)

	owner := []byte("test_owner")
	contract := []byte("test_contract")

	// 创建Object
	object, err := objectManager.CreateObject(owner, contract)
	if err != nil {
		t.Fatalf("Failed to create object: %v", err)
	}

	// 删除Object
	err = objectManager.DeleteObject(object.ID)
	if err != nil {
		t.Fatalf("Failed to delete object: %v", err)
	}

	// 验证删除结果
	_, err = objectManager.GetObject(object.ID)
	if err == nil {
		t.Error("Expected error when getting deleted object")
	}
}

func TestObjectManagerGetObjectStorage(t *testing.T) {
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

	ctx := context.WithValue(context.Background(), types.CtxTxHash{}, types.NewHash([]byte("test_tx_hash_123456789012345678901234567890")))
	objectManager := NewObjectManager(storage, ctx)

	owner := []byte("test_owner")
	contract := []byte("test_contract")

	// 创建Object
	object, err := objectManager.CreateObject(owner, contract)
	if err != nil {
		t.Fatalf("Failed to create object: %v", err)
	}

	// 获取ObjectStorage
	objectStorage, err := objectManager.GetObjectStorage(object.ID)
	if err != nil {
		t.Fatalf("Failed to get object storage: %v", err)
	}

	if objectStorage == nil {
		t.Fatal("GetObjectStorage returned nil")
	}

	if objectStorage.GetObjectID() != object.ID {
		t.Errorf("ObjectStorage ObjectID mismatch: got %s, want %s", objectStorage.GetObjectID().String(), object.ID.String())
	}
}

func TestObjectManagerGetObjectStorageNotFound(t *testing.T) {
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

	ctx := context.Background()
	objectManager := NewObjectManager(storage, ctx)

	// 尝试获取不存在的Object的ObjectStorage
	nonExistentID := types.NewHash([]byte("non_existent_id_123456789012345678901234567890"))
	_, err = objectManager.GetObjectStorage(nonExistentID)

	if err == nil {
		t.Error("Expected error when getting ObjectStorage for non-existent object")
	}
}
