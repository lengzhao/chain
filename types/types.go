package types

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// Hash 32字节Hash类型
type Hash [32]byte

// NewHash 从字节数组创建Hash
func NewHash(data []byte) Hash {
	var h Hash
	copy(h[:], data)
	return h
}

// String 返回Hash的十六进制字符串表示
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

// Bytes 返回Hash的字节数组
func (b Hash) Bytes() []byte {
	return b[:]
}

// Block 区块结构
type Block struct {
	Header       BlockHeader `json:"header"`
	Transactions []Transaction `json:"transactions"`
}

// BlockHeader 区块头
type BlockHeader struct {
	Height     uint64    `json:"height"`
	Timestamp  time.Time `json:"timestamp"`
	PrevHash   Hash      `json:"prev_hash"`
	StateRoot  Hash      `json:"state_root"`
	TxRoot     Hash      `json:"tx_root"`
	Consensus  []byte    `json:"consensus"`
}

// Transaction 交易结构
type Transaction struct {
	Hash        Hash      `json:"hash"`
	From        []byte    `json:"from"`
	To          []byte    `json:"to"`
	Data        []byte    `json:"data"`
	Nonce       uint64    `json:"nonce"`
	Timestamp   time.Time `json:"timestamp"`
	AccessList  AccessList `json:"access_list"`
	Signature   []byte    `json:"signature"`
}

// AccessList 访问列表
type AccessList struct {
	Reads  []Hash `json:"reads"`
	Writes []Hash `json:"writes"`
}

// CalculateHash 计算交易的Hash
func (tx *Transaction) CalculateHash() Hash {
	data := append(tx.From, tx.To...)
	data = append(data, tx.Data...)
	data = append(data, byte(tx.Nonce))
	
	hash := sha256.Sum256(data)
	return NewHash(hash[:])
}

// CalculateBlockHash 计算区块的Hash
func (b *Block) CalculateBlockHash() Hash {
	data := append([]byte{}, byte(b.Header.Height))
	data = append(data, b.Header.PrevHash.Bytes()...)
	data = append(data, b.Header.StateRoot.Bytes()...)
	data = append(data, b.Header.TxRoot.Bytes()...)
	
	hash := sha256.Sum256(data)
	return NewHash(hash[:])
} 