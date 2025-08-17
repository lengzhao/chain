package types

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequest_Serialize_Deserialize(t *testing.T) {
	// 创建测试请求
	originalReq := &Request{
		Type:      "test_request",
		Data:      []byte("request data"),
		Timestamp: time.Now(),
	}

	// 序列化
	data, err := originalReq.Serialize()
	require.NoError(t, err)
	assert.NotNil(t, data)

	// 反序列化
	var newReq Request
	err = newReq.Deserialize(data)
	require.NoError(t, err)

	// 验证数据
	assert.Equal(t, originalReq.Type, newReq.Type)
	assert.Equal(t, originalReq.Data, newReq.Data)
	assert.Equal(t, originalReq.Timestamp.Unix(), newReq.Timestamp.Unix())
}

func TestResponse_Serialize_Deserialize(t *testing.T) {
	// 创建测试响应
	originalResp := &Response{
		Type:      "test_response",
		Data:      []byte("response data"),
		Error:     "test error",
		Timestamp: time.Now(),
	}

	// 序列化
	data, err := originalResp.Serialize()
	require.NoError(t, err)
	assert.NotNil(t, data)

	// 反序列化
	var newResp Response
	err = newResp.Deserialize(data)
	require.NoError(t, err)

	// 验证数据
	assert.Equal(t, originalResp.Type, newResp.Type)
	assert.Equal(t, originalResp.Data, newResp.Data)
	assert.Equal(t, originalResp.Error, newResp.Error)
	assert.Equal(t, originalResp.Timestamp.Unix(), newResp.Timestamp.Unix())
}

func TestVoteMessage_Serialize_Deserialize(t *testing.T) {
	// 创建测试投票消息
	originalVote := &VoteMessage{
		Voter:      "voter1",
		Candidate:  "candidate1",
		VoteAmount: 1000000,
		Round:      1,
	}

	// 序列化
	data, err := originalVote.Serialize()
	require.NoError(t, err)
	assert.NotNil(t, data)

	// 反序列化
	var newVote VoteMessage
	err = newVote.Deserialize(data)
	require.NoError(t, err)

	// 验证数据
	assert.Equal(t, originalVote.Voter, newVote.Voter)
	assert.Equal(t, originalVote.Candidate, newVote.Candidate)
	assert.Equal(t, originalVote.VoteAmount, newVote.VoteAmount)
	assert.Equal(t, originalVote.Round, newVote.Round)
}

func TestBlockProposalMessage_Serialize_Deserialize(t *testing.T) {
	// 创建测试区块
	block := &Block{
		Header: BlockHeader{
			Height:    1,
			Timestamp: time.Now(),
			PrevHash:  NewHash([]byte("prev_hash")),
		},
		Transactions: []Hash{NewHash([]byte("tx1")), NewHash([]byte("tx2"))},
	}

	// 创建测试区块提议消息
	originalProposal := &BlockProposalMessage{
		Validator: "validator1",
		Round:     1,
		Slot:      0,
		Block:     block,
	}

	// 序列化
	data, err := originalProposal.Serialize()
	require.NoError(t, err)
	assert.NotNil(t, data)

	// 反序列化
	var newProposal BlockProposalMessage
	err = newProposal.Deserialize(data)
	require.NoError(t, err)

	// 验证数据
	assert.Equal(t, originalProposal.Validator, newProposal.Validator)
	assert.Equal(t, originalProposal.Round, newProposal.Round)
	assert.Equal(t, originalProposal.Slot, newProposal.Slot)
	assert.Equal(t, originalProposal.Block.Header.Height, newProposal.Block.Header.Height)
	assert.Equal(t, len(originalProposal.Block.Transactions), len(newProposal.Block.Transactions))
}

func TestBlock_Serialize_Deserialize(t *testing.T) {
	// 创建测试区块
	originalBlock := &Block{
		Header: BlockHeader{
			ChainID:   NewHash([]byte("chain_id")),
			Height:    100,
			Timestamp: time.Now(),
			PrevHash:  NewHash([]byte("prev_hash")),
			StateRoot: NewHash([]byte("state_root")),
			TxRoot:    NewHash([]byte("tx_root")),
		},
		Transactions: []Hash{
			NewHash([]byte("transaction1")),
			NewHash([]byte("transaction2")),
		},
	}

	// 序列化
	data, err := originalBlock.Serialize()
	require.NoError(t, err)
	assert.NotNil(t, data)

	// 反序列化
	var newBlock Block
	err = newBlock.Deserialize(data)
	require.NoError(t, err)

	// 验证数据
	assert.Equal(t, originalBlock.Header.Height, newBlock.Header.Height)
	assert.Equal(t, originalBlock.Header.ChainID, newBlock.Header.ChainID)
	assert.Equal(t, originalBlock.Header.PrevHash, newBlock.Header.PrevHash)
	assert.Equal(t, originalBlock.Header.StateRoot, newBlock.Header.StateRoot)
	assert.Equal(t, originalBlock.Header.TxRoot, newBlock.Header.TxRoot)
	assert.Equal(t, len(originalBlock.Transactions), len(newBlock.Transactions))
}

func TestTransaction_Serialize_Deserialize(t *testing.T) {
	// 创建测试交易
	originalTx := &Transaction{
		ChainID: NewHash([]byte("chain_id")),
		From:    []byte("from_address"),
		To:      []byte("to_address"),
		Data:    []byte("transaction_data"),
		Nonce:   123,
		AccessList: AccessList{
			Reads:  []Hash{NewHash([]byte("read1")), NewHash([]byte("read2"))},
			Writes: []Hash{NewHash([]byte("write1"))},
		},
		Signature: []byte("signature_data"),
	}

	// 序列化
	data, err := originalTx.Serialize()
	require.NoError(t, err)
	assert.NotNil(t, data)

	// 反序列化
	var newTx Transaction
	err = newTx.Deserialize(data)
	require.NoError(t, err)

	// 验证数据
	assert.Equal(t, originalTx.ChainID, newTx.ChainID)
	assert.Equal(t, originalTx.From, newTx.From)
	assert.Equal(t, originalTx.To, newTx.To)
	assert.Equal(t, originalTx.Data, newTx.Data)
	assert.Equal(t, originalTx.Nonce, newTx.Nonce)
	assert.Equal(t, originalTx.Signature, newTx.Signature)
	assert.Equal(t, len(originalTx.AccessList.Reads), len(newTx.AccessList.Reads))
	assert.Equal(t, len(originalTx.AccessList.Writes), len(newTx.AccessList.Writes))
}

func TestPerformance_JSON_vs_GOB(t *testing.T) {
	// 创建大量测试数据
	block := &Block{
		Header: BlockHeader{
			Height:    1000,
			Timestamp: time.Now(),
			PrevHash:  NewHash([]byte("prev_hash")),
		},
		Transactions: make([]Hash, 1000),
	}

	// 填充交易数据
	for i := 0; i < 1000; i++ {
		block.Transactions[i] = NewHash([]byte(fmt.Sprintf("transaction_%d", i)))
	}

	// 测试gob序列化性能
	start := time.Now()
	data, err := block.Serialize()
	serializeTime := time.Since(start)
	require.NoError(t, err)

	// 测试gob反序列化性能
	start = time.Now()
	var newBlock Block
	err = newBlock.Deserialize(data)
	deserializeTime := time.Since(start)
	require.NoError(t, err)

	t.Logf("GOB序列化时间: %v", serializeTime)
	t.Logf("GOB反序列化时间: %v", deserializeTime)
	t.Logf("数据大小: %d bytes", len(data))

	// 验证数据完整性
	assert.Equal(t, block.Header.Height, newBlock.Header.Height)
	assert.Equal(t, len(block.Transactions), len(newBlock.Transactions))
}
