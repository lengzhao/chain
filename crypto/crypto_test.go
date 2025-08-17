package crypto

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/cosmos/go-bip39"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCosmosProvider_GenerateKeyPair(t *testing.T) {
	provider := NewCosmosProvider()

	// 测试secp256k1
	keyPair, err := provider.GenerateKeyPair(SECP256K1)
	require.NoError(t, err)
	require.NotNil(t, keyPair)
	assert.Equal(t, SECP256K1, keyPair.Algorithm())
	assert.Equal(t, 33, len(keyPair.PublicKey()))
	assert.Equal(t, 32, len(keyPair.PrivateKey()))

	// 测试ed25519
	keyPair2, err := provider.GenerateKeyPair(ED25519)
	require.NoError(t, err)
	require.NotNil(t, keyPair2)
	assert.Equal(t, ED25519, keyPair2.Algorithm())
	assert.Equal(t, 32, len(keyPair2.PublicKey()))
	assert.Equal(t, 64, len(keyPair2.PrivateKey()))

	// 测试不支持的算法
	_, err = provider.GenerateKeyPair("unsupported")
	assert.Error(t, err)
	assert.Equal(t, ErrUnsupportedAlgorithm, err)
}

func TestCosmosProvider_GenerateKeyPairFromMnemonic(t *testing.T) {
	provider := NewCosmosProvider()

	entropy, err := bip39.NewEntropy(256)
	require.NoError(t, err)
	mnemonic, err := bip39.NewMnemonic(entropy)
	require.NoError(t, err)

	// 测试secp256k1
	keyPair, err := provider.GenerateKeyPairFromMnemonic(mnemonic, SECP256K1)
	require.NoError(t, err)
	require.NotNil(t, keyPair)
	assert.Equal(t, SECP256K1, keyPair.Algorithm())

	// 测试ed25519
	keyPair2, err := provider.GenerateKeyPairFromMnemonic(mnemonic, ED25519)
	require.NoError(t, err)
	require.NotNil(t, keyPair2)
	assert.Equal(t, ED25519, keyPair2.Algorithm())

	// 测试不支持的算法
	_, err = provider.GenerateKeyPairFromMnemonic(mnemonic, "unsupported")
	assert.Error(t, err)
	assert.Equal(t, ErrUnsupportedAlgorithm, err)

	// 测试无效助记词
	_, err = provider.GenerateKeyPairFromMnemonic("invalid mnemonic", SECP256K1)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidMnemonic, err)
}

func TestCosmosProvider_LoadKeyPair(t *testing.T) {
	provider := NewCosmosProvider()

	// 测试secp256k1
	originalKeyPair, err := provider.GenerateKeyPair(SECP256K1)
	require.NoError(t, err)

	loadedKeyPair, err := provider.LoadKeyPair(originalKeyPair.PrivateKey(), SECP256K1)
	require.NoError(t, err)
	require.NotNil(t, loadedKeyPair)
	assert.Equal(t, SECP256K1, loadedKeyPair.Algorithm())

	// 测试ed25519
	originalKeyPair2, err := provider.GenerateKeyPair(ED25519)
	require.NoError(t, err)

	loadedKeyPair2, err := provider.LoadKeyPair(originalKeyPair2.PrivateKey(), ED25519)
	require.NoError(t, err)
	require.NotNil(t, loadedKeyPair2)
	assert.Equal(t, ED25519, loadedKeyPair2.Algorithm())

	// 测试不支持的算法
	_, err = provider.LoadKeyPair([]byte{1, 2, 3}, "unsupported")
	assert.Error(t, err)
	assert.Equal(t, ErrUnsupportedAlgorithm, err)

	// 测试无效私钥
	_, err = provider.LoadKeyPair([]byte{1, 2, 3}, SECP256K1)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPrivateKey, err)
}

func TestCosmosProvider_VerifySignature(t *testing.T) {
	provider := NewCosmosProvider()

	// 测试secp256k1
	keyPair, err := provider.GenerateKeyPair(SECP256K1)
	require.NoError(t, err)

	data := []byte("Test data for signing")
	signature, err := keyPair.Sign(data)
	require.NoError(t, err)

	valid, err := provider.VerifySignature(keyPair.PublicKey(), data, signature, SECP256K1)
	require.NoError(t, err)
	assert.True(t, valid)

	modifiedData := []byte("Modified test data")
	valid, err = provider.VerifySignature(keyPair.PublicKey(), modifiedData, signature, SECP256K1)
	require.NoError(t, err)
	assert.False(t, valid)

	// 测试ed25519
	keyPair2, err := provider.GenerateKeyPair(ED25519)
	require.NoError(t, err)

	signature2, err := keyPair2.Sign(data)
	require.NoError(t, err)

	valid, err = provider.VerifySignature(keyPair2.PublicKey(), data, signature2, ED25519)
	require.NoError(t, err)
	assert.True(t, valid)

	// 测试不支持的算法
	_, err = provider.VerifySignature(keyPair.PublicKey(), data, signature, "unsupported")
	assert.Error(t, err)
	assert.Equal(t, ErrUnsupportedAlgorithm, err)

	// 测试无效公钥
	_, err = provider.VerifySignature([]byte{1, 2, 3}, data, signature, SECP256K1)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPublicKey, err)
}

func TestCosmosKeyPair_Sign(t *testing.T) {
	provider := NewCosmosProvider()
	keyPair, err := provider.GenerateKeyPair(SECP256K1)
	require.NoError(t, err)

	data := []byte("Data to sign")
	signature, err := keyPair.Sign(data)
	require.NoError(t, err)
	assert.NotEmpty(t, signature)

	signature2, err := keyPair.Sign(data)
	require.NoError(t, err)
	assert.Equal(t, signature, signature2)
}

func TestCosmosProvider_WithCustomLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	provider := NewCosmosProviderWithLogger(logger)

	keyPair, err := provider.GenerateKeyPair(SECP256K1)
	require.NoError(t, err)
	assert.NotNil(t, keyPair)
}
