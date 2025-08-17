package crypto

import (
	"log/slog"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/go-bip39"
)

// CosmosProvider 基于cosmos sdk的加密提供者实现
type CosmosProvider struct {
	logger *slog.Logger
}

// NewCosmosProvider 创建新的CosmosProvider实例
func NewCosmosProvider() *CosmosProvider {
	return &CosmosProvider{
		logger: slog.Default(),
	}
}

// NewCosmosProviderWithLogger 创建带有自定义日志器的CosmosProvider实例
func NewCosmosProviderWithLogger(logger *slog.Logger) *CosmosProvider {
	return &CosmosProvider{
		logger: logger,
	}
}

// GenerateKeyPair 生成新密钥对
func (p *CosmosProvider) GenerateKeyPair(algorithm Algorithm) (KeyPair, error) {
	p.logger.Debug("Generating new key pair", "algorithm", algorithm)

	// 生成助记词
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		p.logger.Error("Failed to generate entropy", "error", err)
		return nil, err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		p.logger.Error("Failed to generate mnemonic", "error", err)
		return nil, err
	}

	return p.GenerateKeyPairFromMnemonic(mnemonic, algorithm)
}

// GenerateKeyPairFromMnemonic 从助记词生成密钥对
func (p *CosmosProvider) GenerateKeyPairFromMnemonic(mnemonic string, algorithm Algorithm) (KeyPair, error) {
	p.logger.Debug("Generating key pair from mnemonic", "algorithm", algorithm)

	// 验证助记词
	if !bip39.IsMnemonicValid(mnemonic) {
		p.logger.Error("Invalid mnemonic")
		return nil, ErrInvalidMnemonic
	}

	// 从助记词生成种子
	seed := bip39.NewSeed(mnemonic, "")

	// 根据算法创建私钥
	var privKey cryptotypes.PrivKey

	switch algorithm {
	case SECP256K1:
		// 使用HD钱包派生路径生成私钥
		masterPriv, ch := hd.ComputeMastersFromSeed(seed)
		derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, "m/44'/118'/0'/0/0")
		if err != nil {
			p.logger.Error("Failed to derive private key", "error", err)
			return nil, err
		}
		privKey = &secp256k1.PrivKey{Key: derivedPriv}

	case ED25519:
		// Ed25519使用不同的派生路径
		masterPriv, ch := hd.ComputeMastersFromSeed(seed)
		derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, "m/44'/118'/0'/0/0")
		if err != nil {
			p.logger.Error("Failed to derive private key", "error", err)
			return nil, err
		}
		privKey = ed25519.GenPrivKeyFromSecret(derivedPriv)

	default:
		p.logger.Error("Unsupported algorithm", "algorithm", algorithm)
		return nil, ErrUnsupportedAlgorithm
	}

	return p.createKeyPairWithAlgorithm(privKey, algorithm)
}

// LoadKeyPair 从私钥字节加载密钥对
func (p *CosmosProvider) LoadKeyPair(privateKeyBytes []byte, algorithm Algorithm) (KeyPair, error) {
	p.logger.Debug("Loading key pair from private key bytes", "algorithm", algorithm)

	var privKey cryptotypes.PrivKey

	switch algorithm {
	case SECP256K1:
		if len(privateKeyBytes) != 32 {
			p.logger.Error("Invalid secp256k1 private key length", "length", len(privateKeyBytes))
			return nil, ErrInvalidPrivateKey
		}
		privKey = &secp256k1.PrivKey{Key: privateKeyBytes}

	case ED25519:
		if len(privateKeyBytes) != 64 {
			p.logger.Error("Invalid ed25519 private key length", "length", len(privateKeyBytes))
			return nil, ErrInvalidPrivateKey
		}
		privKey = ed25519.GenPrivKeyFromSecret(privateKeyBytes)

	default:
		p.logger.Error("Unsupported algorithm", "algorithm", algorithm)
		return nil, ErrUnsupportedAlgorithm
	}

	return p.createKeyPairWithAlgorithm(privKey, algorithm)
}

// VerifySignature 验证签名
func (p *CosmosProvider) VerifySignature(publicKey, data, signature []byte, algorithm Algorithm) (bool, error) {
	p.logger.Debug("Verifying signature", "data_length", len(data), "signature_length", len(signature), "algorithm", algorithm)

	var pubKey cryptotypes.PubKey
	var expectedLength int

	switch algorithm {
	case SECP256K1:
		expectedLength = 33
		if len(publicKey) != expectedLength {
			p.logger.Error("Invalid secp256k1 public key length", "length", len(publicKey))
			return false, ErrInvalidPublicKey
		}
		pubKey = &secp256k1.PubKey{Key: publicKey}

	case ED25519:
		expectedLength = 32
		if len(publicKey) != expectedLength {
			p.logger.Error("Invalid ed25519 public key length", "length", len(publicKey))
			return false, ErrInvalidPublicKey
		}
		pubKey = &ed25519.PubKey{Key: publicKey}

	default:
		p.logger.Error("Unsupported algorithm", "algorithm", algorithm)
		return false, ErrUnsupportedAlgorithm
	}

	// 验证签名
	valid := pubKey.VerifySignature(data, signature)
	if !valid {
		p.logger.Debug("Signature verification failed")
		return false, nil // 返回false而不是错误
	}

	p.logger.Debug("Signature verification successful")
	return true, nil
}

// createKeyPairWithAlgorithm 创建KeyPair实例
func (p *CosmosProvider) createKeyPairWithAlgorithm(privKey cryptotypes.PrivKey, algorithm Algorithm) (KeyPair, error) {
	pubKey := privKey.PubKey()

	// 生成地址
	address, err := p.generateAddress(pubKey)
	if err != nil {
		p.logger.Error("Failed to generate address", "error", err)
		return nil, err
	}

	return &CosmosKeyPair{
		algorithm: algorithm,
		privKey:   privKey,
		pubKey:    pubKey,
		address:   address,
	}, nil
}

// generateAddress 生成地址
func (p *CosmosProvider) generateAddress(pubKey cryptotypes.PubKey) (string, error) {
	// 这里使用简化的地址生成，实际项目中可能需要更复杂的地址格式
	// 暂时返回公钥的十六进制表示作为地址
	address := pubKey.Address().String()
	return address, nil
}
