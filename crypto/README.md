# Crypto Module

区块链项目的加密模块，提供密钥管理、数字签名和验证功能。

## 设计原则

### 签名与验证分离

- **签名功能**：需要私钥，通过 `KeyPair` 接口提供
- **验证功能**：只需要公钥/地址，通过 `CryptoProvider` 接口提供

这种设计确保了：
- 验证逻辑的独立性，不需要访问私钥
- 更清晰的职责分离
- 更好的安全性（验证时不会暴露私钥）

## 核心接口

### KeyPair 接口

用于管理密钥对和签名操作：

```go
type KeyPair interface {
    Algorithm() Algorithm           // 获取算法类型
    PublicKey() []byte             // 获取公钥字节
    PrivateKey() []byte            // 获取私钥字节
    Address() string               // 获取地址
    Sign(data []byte) ([]byte, error) // 签名数据
}
```

### CryptoProvider 接口

提供加密服务和验证功能：

```go
type CryptoProvider interface {
    GenerateKeyPair(algorithm Algorithm) (KeyPair, error)
    GenerateKeyPairFromMnemonic(mnemonic string, algorithm Algorithm) (KeyPair, error)
    LoadKeyPair(privateKeyBytes []byte, algorithm Algorithm) (KeyPair, error)
    Hash(data []byte) []byte
    VerifySignature(publicKey, data, signature []byte, algorithm Algorithm) (bool, error)
}
```

## 支持的算法

- `SECP256K1`：比特币和以太坊使用的椭圆曲线算法
- `ED25519`：现代高效的椭圆曲线算法

## 使用示例

### 基本使用

```go
// 创建提供者
provider := crypto.NewCosmosProvider()

// 生成密钥对
keyPair, err := provider.GenerateKeyPair(crypto.SECP256K1)
if err != nil {
    log.Fatal(err)
}

// 签名数据
data := []byte("Hello, Blockchain!")
signature, err := keyPair.Sign(data)
if err != nil {
    log.Fatal(err)
}

// 验证签名（只需要公钥，不需要私钥）
valid, err := provider.VerifySignature(keyPair.PublicKey(), data, signature, crypto.SECP256K1)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Signature valid: %t\n", valid)
```

### 从助记词生成

```go
mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
keyPair, err := provider.GenerateKeyPairFromMnemonic(mnemonic, crypto.ED25519)
if err != nil {
    log.Fatal(err)
}
```

### 从私钥加载

```go
privateKeyBytes := []byte{...} // 32字节的私钥
keyPair, err := provider.LoadKeyPair(privateKeyBytes, crypto.SECP256K1)
if err != nil {
    log.Fatal(err)
}
```

## 安全考虑

1. **私钥保护**：私钥只在 `KeyPair` 内部使用，验证时不需要访问私钥
2. **算法验证**：验证时会检查公钥长度和算法匹配
3. **错误处理**：提供详细的错误信息，但不泄露敏感数据

## 测试

运行测试：

```bash
go test ./crypto/... -v
```

运行示例：

```bash
go run examples/crypto_demo/main.go
```
