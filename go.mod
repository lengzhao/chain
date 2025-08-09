module github.com/govm-net/chain

go 1.24.2

require (
	// 共识引擎
	github.com/tendermint/tendermint v0.37.0
	github.com/cosmos/cosmos-sdk v0.50.0
	
	// 网络层
	github.com/libp2p/go-libp2p v0.32.0
	github.com/libp2p/go-libp2p-pubsub v0.10.0
	
	// 密码学
	github.com/herumi/bls-eth-go-binary v1.33.0
	golang.org/x/crypto v0.17.0
	
	// 性能优化
	github.com/panjf2000/ants/v2 v2.8.2
	github.com/valyala/bytebufferpool v1.0.0
	
	// 存储
	github.com/dgraph-io/badger/v3 v3.2103.5
	github.com/syndtr/goleveldb v1.0.0
	
	// 通信
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
	
	// 日志
	go.uber.org/zap v1.26.0
)
