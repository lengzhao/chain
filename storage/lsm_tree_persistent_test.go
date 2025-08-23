package storage

import (
	"fmt"
	"os"
	"testing"
)

// TestPersistentLSMTree 测试持久化模式的LSM-Tree
func TestPersistentLSMTree(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "persistent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建持久化模式的LSM-Tree
	l, err := NewLSMTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create persistent LSM-Tree: %v", err)
	}
	defer l.Stop()

	// 检查是否为持久化模式
	if l.IsMemoryMode() {
		t.Error("Expected persistent mode, but got memory mode")
	}

	// 测试基本操作
	testKey := []byte("persistent_test_key")
	testValue := []byte("persistent_test_value")

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

	// 验证数据确实写入了磁盘
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected files to be created in the directory")
	}

	fmt.Printf("📁 Persistent directory: %s\n", tempDir)
	fmt.Printf("📄 Files created: %d\n", len(files))
	fmt.Println("✅ Persistent LSM-Tree test passed")
}

// TestDataPersistence 测试数据持久化
func TestDataPersistence(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "persistence-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 第一次创建LSM-Tree并写入数据
	l1, err := NewLSMTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create first LSM-Tree: %v", err)
	}

	testKey := []byte("persistence_key")
	testValue := []byte("persistence_value")

	err = l1.Set(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// 关闭第一个实例
	l1.Stop()

	// 第二次创建LSM-Tree并读取数据
	l2, err := NewLSMTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create second LSM-Tree: %v", err)
	}
	defer l2.Stop()

	// 读取之前写入的数据
	value, err := l2.Get(testKey)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if string(value) != string(testValue) {
		t.Errorf("Expected value %s, got %s", string(testValue), string(value))
	}

	fmt.Println("✅ Data persistence test passed")
}

// TestMemoryVsPersistent 比较内存模式和持久化模式
func TestMemoryVsPersistent(t *testing.T) {
	// 创建临时目录用于持久化测试
	tempDir, err := os.MkdirTemp("", "compare-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建内存模式的LSM-Tree
	memoryL, err := NewLSMTree("")
	if err != nil {
		t.Fatalf("Failed to create memory LSM-Tree: %v", err)
	}
	defer memoryL.Stop()

	// 创建持久化模式的LSM-Tree
	persistentL, err := NewLSMTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create persistent LSM-Tree: %v", err)
	}
	defer persistentL.Stop()

	// 验证模式
	if !memoryL.IsMemoryMode() {
		t.Error("Memory LSM-Tree should be in memory mode")
	}

	if persistentL.IsMemoryMode() {
		t.Error("Persistent LSM-Tree should not be in memory mode")
	}

	// 验证临时目录
	memoryTempDir := memoryL.GetTempDir()
	persistentTempDir := persistentL.GetTempDir()

	if memoryTempDir == "" {
		t.Error("Memory LSM-Tree should have a temp directory")
	}

	if persistentTempDir != "" {
		t.Error("Persistent LSM-Tree should not have a temp directory")
	}

	fmt.Printf("🧠 Memory mode temp dir: %s\n", memoryTempDir)
	fmt.Printf("💾 Persistent mode dir: %s\n", tempDir)
	fmt.Println("✅ Memory vs Persistent comparison test passed")
}

// BenchmarkPersistentLSMTree 持久化模式性能基准测试
func BenchmarkPersistentLSMTree(b *testing.B) {
	// 创建临时目录用于基准测试
	tempDir, err := os.MkdirTemp("", "benchmark-persistent-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	l, err := NewLSMTree(tempDir)
	if err != nil {
		b.Fatalf("Failed to create persistent LSM-Tree: %v", err)
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
