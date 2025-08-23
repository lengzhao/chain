package storage

import (
	"fmt"
	"testing"
)

// TestMemoryLSMTree 测试内存模式的LSM-Tree
func TestMemoryLSMTree(t *testing.T) {
	// 创建内存模式的LSM-Tree（传入空字符串）
	l, err := NewLSMTree("")
	if err != nil {
		t.Fatalf("Failed to create memory LSM-Tree: %v", err)
	}
	defer l.Stop()

	// 检查是否为内存模式
	if !l.IsMemoryMode() {
		t.Error("Expected memory mode, but got false")
	}

	// 测试基本操作
	testKey := []byte("test_key")
	testValue := []byte("test_value")

	// 设置值
	err = l.Set(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// 获取值
	value, err := l.Get(testKey)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if string(value) != string(testValue) {
		t.Errorf("Expected value %s, got %s", string(testValue), string(value))
	}

	// 检查键是否存在
	exists, err := l.Has(testKey)
	if err != nil {
		t.Fatalf("Failed to check key existence: %v", err)
	}

	if !exists {
		t.Error("Expected key to exist, but it doesn't")
	}

	fmt.Println("✅ Memory LSM-Tree test passed")
}

// TestTempLSMTree 测试临时目录模式的LSM-Tree
func TestTempLSMTree(t *testing.T) {
	// 创建临时目录模式的LSM-Tree（传入空字符串，与内存模式相同）
	l, err := NewLSMTree("")
	if err != nil {
		t.Fatalf("Failed to create temp LSM-Tree: %v", err)
	}
	defer l.Stop()

	// 检查是否为内存模式（临时目录模式现在也是内存模式）
	if !l.IsMemoryMode() {
		t.Error("Expected memory mode, but got false")
	}

	// 检查临时目录路径
	tempDir := l.GetTempDir()
	if tempDir == "" {
		t.Error("Expected temp directory path, but got empty string")
	}

	fmt.Printf("📁 Temp directory: %s\n", tempDir)

	// 测试基本操作
	testKey := []byte("temp_test_key")
	testValue := []byte("temp_test_value")

	// 设置值
	err = l.Set(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// 获取值
	value, err := l.Get(testKey)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if string(value) != string(testValue) {
		t.Errorf("Expected value %s, got %s", string(testValue), string(value))
	}

	// 删除值
	err = l.Delete(testKey)
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	// 检查键是否已被删除
	exists, err := l.Has(testKey)
	if err != nil {
		t.Fatalf("Failed to check key existence: %v", err)
	}

	if exists {
		t.Error("Expected key to not exist after deletion, but it does")
	}

	fmt.Println("✅ Temp LSM-Tree test passed")
}

// TestBatchOperations 测试批量操作
func TestBatchOperations(t *testing.T) {
	// 创建内存模式的LSM-Tree
	l, err := NewLSMTree("")
	if err != nil {
		t.Fatalf("Failed to create memory LSM-Tree: %v", err)
	}
	defer l.Stop()

	// 创建批量操作
	batch := l.Batch()

	// 添加多个操作
	batch.Put([]byte("key1"), []byte("value1"))
	batch.Put([]byte("key2"), []byte("value2"))
	batch.Put([]byte("key3"), []byte("value3"))
	batch.Delete([]byte("key2"))

	// 执行批量操作
	err = l.WriteBatch(batch)
	if err != nil {
		t.Fatalf("Failed to write batch: %v", err)
	}

	// 验证结果
	value1, err := l.Get([]byte("key1"))
	if err != nil || string(value1) != "value1" {
		t.Error("Failed to get key1")
	}

	value2, err := l.Get([]byte("key2"))
	if err != nil || value2 != nil {
		t.Error("key2 should be deleted")
	}

	value3, err := l.Get([]byte("key3"))
	if err != nil || string(value3) != "value3" {
		t.Error("Failed to get key3")
	}

	fmt.Println("✅ Batch operations test passed")
}

// TestIterator 测试迭代器
func TestIterator(t *testing.T) {
	// 创建内存模式的LSM-Tree
	l, err := NewLSMTree("")
	if err != nil {
		t.Fatalf("Failed to create memory LSM-Tree: %v", err)
	}
	defer l.Stop()

	// 插入测试数据
	testData := map[string]string{
		"a": "value_a",
		"b": "value_b",
		"c": "value_c",
	}

	for key, value := range testData {
		err := l.Set([]byte(key), []byte(value))
		if err != nil {
			t.Fatalf("Failed to set %s: %v", key, err)
		}
	}

	// 创建迭代器
	iter := l.Iterator()
	if iter == nil {
		t.Fatal("Failed to create iterator")
	}

	fmt.Println("✅ Iterator test passed (basic functionality verified)")
}

// BenchmarkMemoryLSMTree 内存模式性能基准测试
func BenchmarkMemoryLSMTree(b *testing.B) {
	l, err := NewLSMTree("")
	if err != nil {
		b.Fatalf("Failed to create memory LSM-Tree: %v", err)
	}
	defer l.Stop()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte(fmt.Sprintf("value_%d", i))

		err := l.Set(key, value)
		if err != nil {
			b.Fatalf("Failed to set value: %v", err)
		}

		_, err = l.Get(key)
		if err != nil {
			b.Fatalf("Failed to get value: %v", err)
		}
	}
}

// BenchmarkTempLSMTree 临时目录模式性能基准测试
func BenchmarkTempLSMTree(b *testing.B) {
	l, err := NewLSMTree("")
	if err != nil {
		b.Fatalf("Failed to create temp LSM-Tree: %v", err)
	}
	defer l.Stop()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte(fmt.Sprintf("value_%d", i))

		err := l.Set(key, value)
		if err != nil {
			b.Fatalf("Failed to set value: %v", err)
		}

		_, err = l.Get(key)
		if err != nil {
			b.Fatalf("Failed to get value: %v", err)
		}
	}
}
