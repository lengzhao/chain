package types

// ConsensusNetworkInterface 共识模块需要的网络接口
// 这是NetworkInterface的子集，只包含共识需要的功能
type ConsensusNetworkInterface interface {
	// 消息广播和发送
	BroadcastMessage(topic string, data []byte) error
	SendRequest(peerID string, requestType string, data []byte) ([]byte, error)

	// 消息接收
	RegisterMessageHandler(topic string, handler MessageHandler)

	// 节点管理
	GetPeers() []string
}

// NetworkMessage 网络消息
type NetworkMessage struct {
	From  string
	To    string
	Topic string
	Data  []byte
}

// PBFT 共识接口
type PBFT interface {
	Start() error
	Stop() error
	HandleClientRequest(req interface{}) error
	HandleConsensusMessage(msg interface{}) error
	GetState() interface{}
}

// MessageValidator 消息验证器接口
type MessageValidator interface {
	ValidateClientRequest(req interface{}) error
	ValidatePrePrepare(msg interface{}) error
	ValidatePrepare(msg interface{}) error
	ValidateCommit(msg interface{}) error
	ValidateViewChange(msg interface{}) error
	ValidateNewView(msg interface{}) error
}

// ExecutionEngine 执行引擎接口
type ExecutionEngine interface {
	Execute(operation []byte) ([]byte, error)
	GetState() map[string]interface{}
	Reset()
}

// StorageInterface 存储接口
type StorageInterface interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	Has(key []byte) (bool, error)
	Close() error
}

// NetworkInterface 网络模块接口
type NetworkInterface interface {
	BroadcastMessage(topic string, data []byte) error
	RegisterMessageHandler(topic string, handler MessageHandler)
	RegisterRequestHandler(requestType string, handler RequestHandler)
	SendRequest(peerID string, requestType string, data []byte) ([]byte, error)
	GetPeers() []string
	ConnectToPeer(addr string) error

	// 节点地址查询接口
	GetLocalAddresses() []string
	GetLocalPeerID() string
}

// MessageHandler 消息处理器
type MessageHandler func(peerID string, msg Message) error

// RequestHandler 请求处理器
type RequestHandler func(peerID string, msg Request) ([]byte, error)

// ConfigInterface 配置接口
type ConfigInterface interface {
	GetNetworkConfig() NetworkConfig
	GetConsensusConfig() ConsensusConfig
	GetStorageConfig() StorageConfig
	GetExecutionConfig() ExecutionConfig
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	Port           int      `yaml:"port"`
	Host           string   `yaml:"host"`
	MaxPeers       int      `yaml:"max_peers"`
	BootstrapPeers []string `yaml:"bootstrap_peers"`
	PrivateKeyPath string   `yaml:"private_key_path"`
}

// ConsensusConfig 共识配置
type ConsensusConfig struct {
	Algorithm string `yaml:"algorithm"`
	MaxFaulty int    `yaml:"max_faulty"`
	BlockTime int    `yaml:"block_time"`
	BatchSize int    `yaml:"batch_size"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	DataDir     string `yaml:"data_dir"`
	MaxSize     int64  `yaml:"max_size"`
	CacheSize   int    `yaml:"cache_size"`
	Compression bool   `yaml:"compression"`
}

// ExecutionConfig 执行配置
type ExecutionConfig struct {
	MaxThreads int `yaml:"max_threads"`
	BatchSize  int `yaml:"batch_size"`
	Timeout    int `yaml:"timeout"`
}
