package crypto

import (
	"errors"
)

// Algorithm 签名算法类型
type Algorithm string

const (
	SECP256K1 Algorithm = "secp256k1"
	ED25519   Algorithm = "ed25519"
)

// 自定义错误类型
var (
	ErrInvalidPrivateKey    = errors.New("invalid private key")
	ErrInvalidPublicKey     = errors.New("invalid public key")
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrInvalidMnemonic      = errors.New("invalid mnemonic")
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")
)

// KeyPair 密钥对接口
type KeyPair interface {
	// Algorithm 获取算法类型
	Algorithm() Algorithm

	// PublicKey 获取公钥字节
	PublicKey() []byte

	// PrivateKey 获取私钥字节
	PrivateKey() []byte

	// Address 获取地址
	Address() string

	// Sign 签名数据
	Sign(data []byte) ([]byte, error)
}

// CryptoProvider 加密提供者接口
type CryptoProvider interface {
	// GenerateKeyPair 生成新密钥对
	GenerateKeyPair(algorithm Algorithm) (KeyPair, error)

	// GenerateKeyPairFromMnemonic 从助记词生成密钥对
	GenerateKeyPairFromMnemonic(mnemonic string, algorithm Algorithm) (KeyPair, error)

	// LoadKeyPair 从私钥字节加载密钥对
	LoadKeyPair(privateKeyBytes []byte, algorithm Algorithm) (KeyPair, error)

	// Hash 计算哈希
	Hash(data []byte) []byte

	// VerifySignature 验证签名
	VerifySignature(publicKey, data, signature []byte, algorithm Algorithm) (bool, error)
}
