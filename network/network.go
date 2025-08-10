package network

import (
	"context"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/govm-net/chain/config"
	"github.com/govm-net/chain/consensus"
	"github.com/govm-net/chain/types"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/multiformats/go-multiaddr"
)

// Network 网络层 - 基于libp2p的现代化P2P网络
type Network struct {
	config    *config.NetworkConfig
	consensus *consensus.Consensus
	host      host.Host
	ctx       context.Context
	cancel    context.CancelFunc

	// DHT自动管理（由libp2p.Routing自动处理）

	// Gossipsub消息传播
	pubsub   *pubsub.PubSub
	topics   map[string]*pubsub.Topic
	topicsMu sync.RWMutex

	// 消息处理
	messageHandlers map[string]MessageHandler
	handlersMu      sync.RWMutex

	// 点对点请求处理
	requestHandlers map[string]RequestHandler
	requestMu       sync.RWMutex

	// 节点发现
	mdnsService mdns.Service
}

// MessageHandler 消息处理器
type MessageHandler func(peer.ID, []byte) error

// RequestHandler 点对点请求处理器
type RequestHandler func(peer.ID, []byte) ([]byte, error)

// New 创建新的网络实例
func New(cfg config.NetworkConfig, consensus *consensus.Consensus) (*Network, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 获取私钥
	priv, err := getOrGeneratePrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("获取私钥失败: %w", err)
	}

	// 创建连接管理器
	connManager, err := connmgr.NewConnManager(
		20,                                     // 最小连接数（默认值）
		cfg.MaxPeers,                           // 最大连接数
		connmgr.WithGracePeriod(time.Minute*5), // 优雅期
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("创建连接管理器失败: %w", err)
	}

	// 解析 bootstrap 节点
	var bootstrapPeers []peer.AddrInfo
	for _, peerAddr := range cfg.BootstrapPeers {
		maddr, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			slog.Error("failed to parse bootstrap peer address", "peer_addr", peerAddr, "error", err)
			continue
		}

		addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			slog.Error("failed to parse bootstrap peer info", "peer_addr", peerAddr, "error", err)
			continue
		}

		bootstrapPeers = append(bootstrapPeers, *addrInfo)
	}
	if len(bootstrapPeers) == 0 {
		bootstrapPeers = dht.GetDefaultBootstrapPeerAddrInfos()
	}

	// 创建libp2p主机
	host, err := libp2p.New(
		// 使用生成的密钥对
		libp2p.Identity(priv),
		// 监听地址
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/%s/tcp/%d", cfg.Host, cfg.Port),         // TCP连接
			fmt.Sprintf("/ip4/%s/udp/%d/quic-v1", cfg.Host, cfg.Port), // QUIC传输
		),
		// 支持TLS连接
		// libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// 支持noise连接
		libp2p.Security(noise.ID, noise.New),
		// 支持默认传输协议
		libp2p.DefaultTransports,
		// 尝试使用uPNP为NAT主机开放端口
		libp2p.NATPortMap(),
		// 让此主机使用DHT查找其他主机，并设置bootstrap节点
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			dhtInstance, err := dht.New(context.Background(), h, dht.BootstrapPeers(bootstrapPeers...))
			return dhtInstance, err
		}),
		// 启用NAT服务以帮助其他节点检测NAT
		libp2p.EnableNATService(),
		// 启用自动中继
		libp2p.EnableAutoRelayWithStaticRelays([]peer.AddrInfo{}),
		// 启用洞穿
		libp2p.EnableHolePunching(),
		// 使用连接管理器
		libp2p.ConnectionManager(connManager),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("创建libp2p主机失败: %w", err)
	}

	network := &Network{
		config:          &cfg,
		consensus:       consensus,
		host:            host,
		ctx:             ctx,
		cancel:          cancel,
		messageHandlers: make(map[string]MessageHandler),
		requestHandlers: make(map[string]RequestHandler),
		topics:          make(map[string]*pubsub.Topic),
	}

	// DHT由libp2p.Routing自动初始化

	// 初始化Gossipsub
	if err := network.initializeGossipsub(); err != nil {
		cancel()
		return nil, fmt.Errorf("初始化Gossipsub失败: %w", err)
	}

	network.host.SetStreamHandler("/chain/1.0.0/request", network.handleRequest)

	// 设置网络通知
	host.Network().Notify(network)

	return network, nil
}

// initializeGossipsub 初始化Gossipsub
func (n *Network) initializeGossipsub() error {
	// 创建Gossipsub实例
	pubsubInstance, err := pubsub.NewGossipSub(n.ctx, n.host)
	if err != nil {
		return fmt.Errorf("创建Gossipsub失败: %w", err)
	}

	n.pubsub = pubsubInstance
	return nil
}

// handleRequest 处理点对点请求
func (n *Network) handleRequest(stream network.Stream) {
	// 不要在这里立即关闭流，让客户端先读取响应

	slog.Info("handling request", "remote_peer", stream.Conn().RemotePeer().String())

	defer func() {
		time.Sleep(100 * time.Millisecond)
		stream.Close()
	}()

	// 读取请求数据 - 支持大数据量
	var data []byte
	buffer := make([]byte, 1024) // 1KB缓冲区

	// 读取所有可用数据
	bytesRead, err := stream.Read(buffer)
	if err != nil {
		slog.Error("failed to read request data", "error", err)
		return
	}

	if bytesRead == 0 {
		slog.Error("received empty request data")
		return
	}

	data = buffer[:bytesRead]

	if len(data) == 0 {
		slog.Error("received empty request data")
		return
	}
	slog.Info("received request data", "bytes_read", len(data), "data", string(data))

	// 解析请求
	var req types.Request
	if err := req.Deserialize(data); err != nil {
		slog.Error("failed to deserialize request", "error", err)
		return
	}

	slog.Info("parsed request", "type", req.Type, "data_len", len(req.Data))

	// 调用处理器
	handler, exists := n.requestHandlers[req.Type]
	if !exists {
		return
	}

	slog.Info("calling handler", "type", req.Type, "peer", stream.Conn().RemotePeer().String())
	respData, err := handler(stream.Conn().RemotePeer(), req.Data)
	if err != nil {
		slog.Error("failed to process request", "error", err)
		// 发送错误响应
		resp := types.Response{
			Type: "error",
			Data: []byte(err.Error()),
		}
		respBytes, _ := resp.Serialize()
		if _, err := stream.Write(respBytes); err != nil {
			slog.Error("failed to send error response", "error", err)
		}
		return
	}

	slog.Info("handler completed", "type", req.Type, "response_len", len(respData))

	// 发送响应
	resp := types.Response{
		Type: req.Type,
		Data: respData,
	}
	respBytes, _ := resp.Serialize()
	slog.Info("sending response", "type", req.Type, "response_bytes", len(respBytes))
	if _, err := stream.Write(respBytes); err != nil {
		slog.Error("failed to send response", "error", err)
	} else {
		slog.Info("response sent successfully")
	}
}

// Start 启动网络
func (n *Network) Start() error {
	// 1. 启动mDNS发现
	n.startMDNSDiscovery()

	// 2. 可信节点现在作为 DHT bootstrap 节点，DHT 会自动处理连接
	slog.Info("network started, bootstrap peers configured as DHT bootstrap nodes")

	return nil
}

// Stop 停止网络
func (n *Network) Stop() {
	n.cancel()

	// 停止mDNS服务
	if n.mdnsService != nil {
		n.mdnsService.Close()
	}

	// 关闭主机
	if n.host != nil {
		n.host.Close()
	}
}

// startMDNSDiscovery 启动mDNS发现
func (n *Network) startMDNSDiscovery() {
	n.mdnsService = mdns.NewMdnsService(n.host, "chain-network", n)

	// 启动mDNS服务
	if err := n.mdnsService.Start(); err != nil {
		slog.Error("failed to start mDNS service", "error", err)
		return
	}

	slog.Info("mDNS service started", "service_name", "chain-network")
}

// BroadcastMessage 广播消息
func (n *Network) BroadcastMessage(topic string, data []byte) error {
	if n.pubsub == nil {
		return fmt.Errorf("Gossipsub未初始化")
	}

	topicObj, err := n.getOrCreateTopic(topic)
	if err != nil {
		return fmt.Errorf("获取主题失败: %w", err)
	}

	message := &types.Message{
		Type: topic,
		Data: data,
		From: n.host.ID().String(),
		To:   "",
		Time: time.Now(),
	}

	messageData, err := message.Serialize()
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	if err := topicObj.Publish(n.ctx, messageData); err != nil {
		return fmt.Errorf("发布消息失败: %w", err)
	}

	return nil
}

// getOrCreateTopic 获取或创建主题
func (n *Network) getOrCreateTopic(topicName string) (*pubsub.Topic, error) {
	n.topicsMu.Lock()
	defer n.topicsMu.Unlock()

	if topic, exists := n.topics[topicName]; exists {
		return topic, nil
	}

	topic, err := n.pubsub.Join(topicName)
	if err != nil {
		return nil, fmt.Errorf("加入主题失败: %w", err)
	}

	n.topics[topicName] = topic
	return topic, nil
}

// SubscribeToTopic 订阅主题
func (n *Network) SubscribeToTopic(topicName string) error {
	if n.pubsub == nil {
		return fmt.Errorf("Gossipsub未初始化")
	}

	topic, err := n.getOrCreateTopic(topicName)
	if err != nil {
		return fmt.Errorf("获取主题失败: %w", err)
	}

	subscription, err := topic.Subscribe()
	if err != nil {
		return fmt.Errorf("订阅主题失败: %w", err)
	}

	go n.handleSubscription(topicName, subscription)
	return nil
}

// handleSubscription 处理订阅
func (n *Network) handleSubscription(topicName string, subscription *pubsub.Subscription) {
	defer subscription.Cancel()

	for {
		msg, err := subscription.Next(n.ctx)
		if err != nil {
			if err == context.Canceled {
				return
			}
			slog.Error("failed to receive message", "error", err)
			continue
		}

		if msg.ReceivedFrom == n.host.ID() {
			continue // 忽略自己发送的消息
		}

		go n.processPubsubMessage(topicName, msg)
	}
}

// processPubsubMessage 处理pubsub消息
func (n *Network) processPubsubMessage(topicName string, msg *pubsub.Message) {
	n.handlersMu.RLock()
	handler, exists := n.messageHandlers[topicName]
	n.handlersMu.RUnlock()

	if !exists {
		return
	}

	if err := handler(msg.ReceivedFrom, msg.Data); err != nil {
		slog.Error("failed to process message", "error", err)
	}
}

// RegisterMessageHandler 注册消息处理器
func (n *Network) RegisterMessageHandler(topic string, handler MessageHandler) {
	n.handlersMu.Lock()
	defer n.handlersMu.Unlock()
	n.messageHandlers[topic] = handler
}

// RegisterRequestHandler 注册点对点请求处理器
func (n *Network) RegisterRequestHandler(requestType string, handler RequestHandler) {
	n.requestMu.Lock()
	defer n.requestMu.Unlock()
	n.requestHandlers[requestType] = handler
}

// SendRequest 发送点对点请求
func (n *Network) SendRequest(peerID peer.ID, requestType string, data []byte) ([]byte, error) {
	// 创建流
	stream, err := n.host.NewStream(n.ctx, peerID, "/chain/1.0.0/request")
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// 构造请求
	req := types.Request{
		Type: requestType,
		Data: data,
	}

	// 序列化请求
	reqData, err := req.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	// 发送请求
	if _, err := stream.Write(reqData); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 读取响应
	var responseData []byte
	buffer := make([]byte, 4096)
	for {
		bytesRead, err := stream.Read(buffer)
		if bytesRead > 0 {
			responseData = append(responseData, buffer[:bytesRead]...)
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			slog.Error("failed to read response", "error", err, "peer", peerID.String())
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		if bytesRead == 0 {
			break
		}
	}

	slog.Info("read response", "bytes_read", len(responseData), "peer", peerID.String())

	// 解析响应
	var resp types.Response
	if err := resp.Deserialize(responseData); err != nil {
		return nil, fmt.Errorf("failed to deserialize response: %w", err)
	}

	// 检查是否是错误响应
	if resp.Type == "error" {
		return nil, fmt.Errorf("remote error: %s", string(resp.Data))
	}

	return resp.Data, nil
}

// RequestBlock 请求特定高度的区块
func (n *Network) RequestBlock(peerID peer.ID, height uint64) ([]byte, error) {
	// 序列化区块高度
	heightData := []byte(fmt.Sprintf("%d", height))
	return n.SendRequest(peerID, types.RequestTypeGetBlock, heightData)
}

// RequestChainSync 请求链同步
func (n *Network) RequestChainSync(peerID peer.ID, fromHeight uint64) ([]byte, error) {
	// 序列化起始高度
	fromData := []byte(fmt.Sprintf("%d", fromHeight))
	return n.SendRequest(peerID, types.RequestTypeSync, fromData)
}

// GetPeers 获取连接的节点
func (n *Network) GetPeers() []peer.ID {
	return n.host.Network().Peers()
}

// ConnectToPeer 连接到指定节点（外部API）
func (n *Network) ConnectToPeer(addr string) error {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return fmt.Errorf("解析地址失败: %w", err)
	}

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return fmt.Errorf("解析节点信息失败: %w", err)
	}

	if err := n.host.Connect(n.ctx, *info); err != nil {
		return fmt.Errorf("连接节点失败: %w", err)
	}

	return nil
}

// IsPeerConnected 检查节点是否已连接
func (n *Network) IsPeerConnected(peerID peer.ID) bool {
	conns := n.host.Network().ConnsToPeer(peerID)
	return len(conns) > 0
}

// GetPubsub 获取Gossipsub实例
func (n *Network) GetPubsub() *pubsub.PubSub {
	return n.pubsub
}

// ListTopics 列出所有主题
func (n *Network) ListTopics() []string {
	if n.pubsub == nil {
		return []string{}
	}
	return n.pubsub.GetTopics()
}

// GetTopicPeers 获取主题中的节点
func (n *Network) GetTopicPeers(topicName string) []peer.ID {
	if n.pubsub == nil {
		return []peer.ID{}
	}
	topic, err := n.pubsub.Join(topicName)
	if err != nil {
		return []peer.ID{}
	}
	return topic.ListPeers()
}

// GetHost 获取libp2p主机
func (n *Network) GetHost() host.Host {
	return n.host
}

// IsMDNSEnabled 检查mDNS是否启用
func (n *Network) IsMDNSEnabled() bool {
	return n.mdnsService != nil
}

// GetMDNSStatus 获取mDNS状态
func (n *Network) GetMDNSStatus() map[string]interface{} {
	return map[string]interface{}{
		"enabled":      n.mdnsService != nil,
		"service":      n.mdnsService != nil,
		"service_name": "chain-network",
	}
}

// GetConnectionStats 获取连接统计信息
func (n *Network) GetConnectionStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 获取当前连接的节点数量
	peers := n.GetPeers()
	stats["current_peers"] = len(peers)

	// 获取配置的最大节点数
	stats["max_peers"] = n.config.MaxPeers

	// 计算连接使用率
	if n.config.MaxPeers > 0 {
		usage := float64(len(peers)) / float64(n.config.MaxPeers) * 100
		stats["usage_percentage"] = usage
	} else {
		stats["usage_percentage"] = 0.0
	}

	// 获取配置的 bootstrap 节点数量
	stats["bootstrap_peers"] = len(n.config.BootstrapPeers)

	return stats
}

// 实现network.Notifiee接口
func (n *Network) Listen(network.Network, multiaddr.Multiaddr)      {}
func (n *Network) ListenClose(network.Network, multiaddr.Multiaddr) {}

func (n *Network) Connected(net network.Network, conn network.Conn) {
	peerID := conn.RemotePeer()
	slog.Info("peer connected", "peer_id", peerID.String())
}

func (n *Network) Disconnected(net network.Network, conn network.Conn) {
	peerID := conn.RemotePeer()
	slog.Info("peer disconnected", "peer_id", peerID.String())
}

func (n *Network) OpenedStream(network.Network, network.Stream) {}
func (n *Network) ClosedStream(network.Network, network.Stream) {}

// 实现mdns.Notifee接口
func (n *Network) HandlePeerFound(pi peer.AddrInfo) {
	// 避免连接自己
	if pi.ID == n.host.ID() {
		return
	}

	slog.Info("mDNS peer discovered", "peer_id", pi.ID.String(), "addresses", pi.Addrs)

	if err := n.host.Connect(n.ctx, pi); err != nil {
		slog.Error("failed to connect to discovered peer", "peer_id", pi.ID.String(), "error", err)
	} else {
		slog.Info("successfully connected to discovered peer", "peer_id", pi.ID.String())
	}
}

// getOrGeneratePrivateKey 获取或生成私钥
func getOrGeneratePrivateKey(keyPath string) (crypto.PrivKey, error) {
	// 如果指定了私钥文件路径
	if keyPath != "" {
		// 检查文件是否存在
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			// 文件不存在，生成新私钥并保存
			slog.Info("private key file not found, generating new key", "key_path", keyPath)
			priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, rand.Reader)
			if err != nil {
				return nil, fmt.Errorf("生成私钥失败: %w", err)
			}

			// 保存私钥到文件
			if err := SavePrivateKeyToFile(priv, keyPath); err != nil {
				return nil, fmt.Errorf("保存私钥失败: %w", err)
			}

			slog.Info("new private key saved", "key_path", keyPath)
			return priv, nil
		}

		// 文件存在，尝试读取
		priv, err := loadPrivateKeyFromFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("读取私钥文件失败: %w", err)
		}

		slog.Info("private key loaded from file", "key_path", keyPath)
		return priv, nil
	}

	// 没有指定文件路径，生成临时私钥
	slog.Info("no private key path specified, generating temporary key")
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成私钥失败: %w", err)
	}

	return priv, nil
}

// loadPrivateKeyFromFile 从文件加载私钥
func loadPrivateKeyFromFile(path string) (crypto.PrivKey, error) {
	// 读取私钥文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}

	// 解析PEM格式
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("无效的PEM格式私钥文件")
	}

	// 解析私钥
	priv, err := crypto.UnmarshalPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	return priv, nil
}

// SavePrivateKeyToFile 保存私钥到文件
func SavePrivateKeyToFile(priv crypto.PrivKey, path string) error {
	// 序列化私钥
	privBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return fmt.Errorf("序列化私钥失败: %w", err)
	}

	// 创建PEM块
	block := &pem.Block{
		Type:  "LIBP2P PRIVATE KEY",
		Bytes: privBytes,
	}

	// 编码为PEM格式
	pemData := pem.EncodeToMemory(block)

	// 写入文件
	err = os.WriteFile(path, pemData, 0600) // 只有所有者可读写
	if err != nil {
		return fmt.Errorf("写入私钥文件失败: %w", err)
	}

	return nil
}
