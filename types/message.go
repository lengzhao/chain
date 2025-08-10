package types

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"
)

// Message 网络消息
type Message struct {
	Type string    `json:"type"`
	Data []byte    `json:"data"`
	From string    `json:"from"`
	To   string    `json:"to"`
	Time time.Time `json:"time"`
}

// Serialize 序列化消息
func (m *Message) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// DeserializeMessage 反序列化消息
func DeserializeMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// MessageType 消息类型常量
const (
	MessageTypeBlock       = "block"
	MessageTypeTransaction = "transaction"
	MessageTypeConsensus   = "consensus"
	MessageTypePing        = "ping"
	MessageTypePong        = "pong"
)

// Request 点对点请求
type Request struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

// Response 点对点响应
type Response = Request

// Serialize 序列化请求 - 使用二进制格式: typeLen(int16) + dataLen(int16) + type + data
func (r *Request) Serialize() ([]byte, error) {
	typeLen := len(r.Type)
	dataLen := len(r.Data)

	// 检查长度限制 (int16 最大值 65535)
	if typeLen > 65535 || dataLen > 65535 {
		return nil, fmt.Errorf("type or data length exceeds int16 limit")
	}

	// 计算总长度: 2 bytes (typeLen) + 2 bytes (dataLen) + typeLen + dataLen
	totalLen := 4 + typeLen + dataLen
	result := make([]byte, totalLen)

	// 写入 typeLen (int16, 小端序)
	binary.LittleEndian.PutUint16(result[0:2], uint16(typeLen))

	// 写入 dataLen (int16, 小端序)
	binary.LittleEndian.PutUint16(result[2:4], uint16(dataLen))

	// 写入 type
	copy(result[4:4+typeLen], []byte(r.Type))

	// 写入 data
	if dataLen > 0 {
		copy(result[4+typeLen:], r.Data)
	}

	return result, nil
}

// Deserialize 反序列化请求 - 解析二进制格式
func (r *Request) Deserialize(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("data too short, need at least 4 bytes")
	}

	// 读取 typeLen (int16, 小端序)
	typeLen := int(binary.LittleEndian.Uint16(data[0:2]))

	// 读取 dataLen (int16, 小端序)
	dataLen := int(binary.LittleEndian.Uint16(data[2:4]))

	// 验证数据长度
	expectedLen := 4 + typeLen + dataLen
	if len(data) < expectedLen {
		return fmt.Errorf("data length mismatch, expected %d, got %d", expectedLen, len(data))
	}

	// 读取 type
	r.Type = string(data[4 : 4+typeLen])

	// 读取 data
	if dataLen > 0 {
		r.Data = make([]byte, dataLen)
		copy(r.Data, data[4+typeLen:4+typeLen+dataLen])
	} else {
		r.Data = nil
	}

	return nil
}

// 请求类型常量
const (
	RequestTypePing     = "ping"
	RequestTypeGetBlock = "get_block"
	RequestTypeGetChain = "get_chain"
	RequestTypeSync     = "sync"
)
