package types

import (
	"testing"
)

func TestNewObject(t *testing.T) {
	owner := []byte("owner123")
	contract := []byte("contract456")

	object := NewObject(owner, contract)

	if object == nil {
		t.Fatal("NewObject returned nil")
	}

	if string(object.Owner) != string(owner) {
		t.Errorf("Owner mismatch: got %s, want %s", string(object.Owner), string(owner))
	}

	if string(object.Contract) != string(contract) {
		t.Errorf("Contract mismatch: got %s, want %s", string(object.Contract), string(contract))
	}

	// 检查ID是否为空（应该由ObjectManager设置）
	var emptyHash Hash
	if object.ID != emptyHash {
		t.Errorf("ID should be empty initially, got %s", object.ID.String())
	}
}

func TestObjectSerializeDeserialize(t *testing.T) {
	owner := []byte("test_owner")
	contract := []byte("test_contract")

	object := NewObject(owner, contract)
	object.ID = NewHash([]byte("test_id_123456789012345678901234567890"))

	// 序列化
	data, err := object.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize object: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Serialized data is empty")
	}

	// 反序列化
	newObject := &Object{}
	err = newObject.Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize object: %v", err)
	}

	// 验证数据
	if string(newObject.Owner) != string(owner) {
		t.Errorf("Owner mismatch after deserialization: got %s, want %s", string(newObject.Owner), string(owner))
	}

	if string(newObject.Contract) != string(contract) {
		t.Errorf("Contract mismatch after deserialization: got %s, want %s", string(newObject.Contract), string(contract))
	}

	if newObject.ID != object.ID {
		t.Errorf("ID mismatch after deserialization: got %s, want %s", newObject.ID.String(), object.ID.String())
	}
}

func TestObjectGetExpiresAt(t *testing.T) {
	object := NewObject([]byte("owner"), []byte("contract"))

	// 测试已废弃的方法返回0
	expiresAt := object.GetExpiresAt()
	if expiresAt != 0 {
		t.Errorf("GetExpiresAt should return 0 (deprecated), got %d", expiresAt)
	}
}

func TestObjectWithEmptyData(t *testing.T) {
	// 测试空数据
	object := NewObject(nil, nil)

	if object.Owner != nil {
		t.Errorf("Owner should be nil, got %v", object.Owner)
	}

	if object.Contract != nil {
		t.Errorf("Contract should be nil, got %v", object.Contract)
	}
}

func TestObjectSerializeEmptyObject(t *testing.T) {
	object := &Object{}

	// 测试序列化空对象
	data, err := object.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize empty object: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Serialized data should not be empty")
	}

	// 测试反序列化
	newObject := &Object{}
	err = newObject.Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize empty object: %v", err)
	}
}

func TestObjectDeserializeInvalidData(t *testing.T) {
	object := &Object{}

	// 测试反序列化无效数据
	invalidData := []byte("invalid data")
	err := object.Deserialize(invalidData)

	if err == nil {
		t.Error("Expected error when deserializing invalid data")
	}
}
