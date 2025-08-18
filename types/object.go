package types

import (
	"bytes"
	"encoding/gob"
)

// Object 表示一个智能合约创建的数据对象
type Object struct {
	ID       Hash   `json:"id"`       // Object唯一标识
	Owner    []byte `json:"owner"`    // Object所有者
	Contract []byte `json:"contract"` // 创建Object的合约
}

// NewObject 创建新的Object
func NewObject(owner, contract []byte) *Object {
	return &Object{
		Owner:    owner,
		Contract: contract,
	}
}

// GetExpiresAt 获取Object的过期时间（已移除，使用独立的生命周期表）
// 这个方法已废弃，请使用ObjectLifecycleTable.GetLifecycle()获取生命周期信息
func (o *Object) GetExpiresAt() int64 {
	// 返回0表示需要从生命周期表获取
	return 0
}

// Serialize 序列化Object
func (o *Object) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(o)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize 反序列化Object
func (o *Object) Deserialize(data []byte) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	return decoder.Decode(o)
}
