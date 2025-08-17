package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
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
	Transactions []Hash      `json:"transactions"`
}

// BlockHeader 区块头
type BlockHeader struct {
	ChainID   Hash      `json:"chain_id"`
	Height    uint64    `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	PrevHash  Hash      `json:"prev_hash"`
	StateRoot Hash      `json:"state_root"`
	TxRoot    Hash      `json:"tx_root"`
}

// Transaction 交易结构
type Transaction struct {
	ChainID    Hash       `json:"chain_id"`
	From       []byte     `json:"from"`
	To         []byte     `json:"to"`
	Data       []byte     `json:"data"`
	Nonce      uint64     `json:"nonce"`
	AccessList AccessList `json:"access_list"`
	Signature  []byte     `json:"signature"`
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
	data = append(data, []byte(fmt.Sprintf("%d", tx.Nonce))...)
	for _, read := range tx.AccessList.Reads {
		data = append(data, read.Bytes()...)
	}
	for _, write := range tx.AccessList.Writes {
		data = append(data, write.Bytes()...)
	}

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

// Serialize 序列化区块
func (b *Block) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(b)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize 反序列化区块
func (b *Block) Deserialize(data []byte) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	return decoder.Decode(b)
}

// Serialize 序列化交易
func (tx *Transaction) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(tx)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize 反序列化交易
func (tx *Transaction) Deserialize(data []byte) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	return decoder.Decode(tx)
}

// Serialize 序列化区块头
func (bh *BlockHeader) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(bh)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize 反序列化区块头
func (bh *BlockHeader) Deserialize(data []byte) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	return decoder.Decode(bh)
}

// Serialize 序列化访问列表
func (al *AccessList) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(al)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize 反序列化访问列表
func (al *AccessList) Deserialize(data []byte) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	return decoder.Decode(al)
}

// DPOS治理交易数据

// ValidatorRegisterData 验证者注册数据
type ValidatorRegisterData struct {
	PublicKey []byte `json:"public_key"` // 公钥
	Stake     int64  `json:"stake"`      // 质押数量
}

// ValidatorExitData 验证者退出数据
type ValidatorExitData struct {
	// 退出时不需要额外数据，使用交易发送者地址
}

// VoteData 投票数据
type VoteData struct {
	Candidate string `json:"candidate"` // 候选人地址
	Amount    int64  `json:"amount"`    // 投票数量
}

// UnvoteData 取消投票数据
type UnvoteData struct {
	Candidate string `json:"candidate"` // 候选人地址
	Amount    int64  `json:"amount"`    // 取消投票数量
}

// UnlockData 解锁数据
type UnlockData struct {
	Candidate string `json:"candidate"` // 候选人地址
	Amount    int64  `json:"amount"`    // 解锁数量
}

// 系统预留地址（前1024个地址）
const (
	// 治理相关地址
	AddressVote              = "0x0000000000000000000000000000000000000001" // 投票
	AddressValidatorRegister = "0x0000000000000000000000000000000000000002" // 验证者注册
	AddressValidatorExit     = "0x0000000000000000000000000000000000000003" // 验证者退出
	AddressUnvote            = "0x0000000000000000000000000000000000000004" // 取消投票
	AddressUnlock            = "0x0000000000000000000000000000000000000005" // 请求解锁
)

// GetTransactionType 获取交易类型
func (tx *Transaction) GetTransactionType() string {
	toAddr := string(tx.To)

	switch toAddr {
	case AddressVote:
		return "vote"
	case AddressValidatorRegister:
		return "validator_register"
	case AddressValidatorExit:
		return "validator_exit"
	case AddressUnvote:
		return "unvote"
	case AddressUnlock:
		return "unlock"
	default:
		return "transfer"
	}
}

// IsGovernanceTransaction 判断是否为治理交易
func (tx *Transaction) IsGovernanceTransaction() bool {
	toAddr := string(tx.To)
	return toAddr == AddressVote ||
		toAddr == AddressValidatorRegister ||
		toAddr == AddressValidatorExit ||
		toAddr == AddressUnvote ||
		toAddr == AddressUnlock
}
