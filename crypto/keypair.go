package crypto

import (
	"log/slog"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// CosmosKeyPair 基于cosmos sdk的密钥对实现
type CosmosKeyPair struct {
	algorithm Algorithm
	privKey   cryptotypes.PrivKey
	pubKey    cryptotypes.PubKey
	address   string
}

// Algorithm 获取算法类型
func (kp *CosmosKeyPair) Algorithm() Algorithm {
	return kp.algorithm
}

// PublicKey 获取公钥字节
func (kp *CosmosKeyPair) PublicKey() []byte {
	return kp.pubKey.Bytes()
}

// PrivateKey 获取私钥字节
func (kp *CosmosKeyPair) PrivateKey() []byte {
	return kp.privKey.Bytes()
}

// Address 获取地址
func (kp *CosmosKeyPair) Address() string {
	return kp.address
}

// Sign 签名数据
func (kp *CosmosKeyPair) Sign(data []byte) ([]byte, error) {
	slog.Debug("Signing data", "data_length", len(data))

	signature, err := kp.privKey.Sign(data)
	if err != nil {
		slog.Error("Failed to sign data", "error", err)
		return nil, err
	}

	slog.Debug("Data signed successfully", "signature_length", len(signature))
	return signature, nil
}
