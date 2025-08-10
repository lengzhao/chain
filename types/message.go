package types

import (
	"encoding/json"
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
	MessageTypeBlock      = "block"
	MessageTypeTransaction = "transaction"
	MessageTypeConsensus  = "consensus"
	MessageTypePing       = "ping"
	MessageTypePong       = "pong"
) 