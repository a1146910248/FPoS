package p2p

import (
	"FPoS/config"
	"FPoS/core/consensus"
	"FPoS/core/ethereum"
	"FPoS/types"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/multiformats/go-multiaddr"
	"os"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

var (
	globalNode *Layer2Node
	nodeOnce   sync.Once
)

// GetNode 获取全局节点实例
func GetNode() *Layer2Node {
	return globalNode
}

type Layer2Node struct {
	host                 host.Host
	dht                  *dht.IpfsDHT
	pubsub               *pubsub.PubSub
	ctx                  context.Context
	topic                P2PTopic
	blockCache           *sync.Map
	txPool               *sync.Map
	txCountMu            sync.RWMutex
	txCount              uint64    //交易数量统计
	txHistory            *sync.Map // 存储所有历史交易
	txHistoryMu          sync.RWMutex
	stateRoot            string
	latestBlock          uint64
	handlers             types.Handlers
	pingService          *ping.PingService
	bootstrapPeers       []string
	mu                   sync.RWMutex
	privateKey           crypto.PrivKey
	publicKey            crypto.PubKey
	minGasPrice          uint64
	isSequencer          bool
	sequencer            *Sequencer
	currentSyncRequestID string // 当前交易同步请求ID
	isSyncing            bool
	initialized          bool
	stateDB              *StateDB
	electionMgr          *consensus.ElectionManager
	periodicTxStarted    bool
}

type P2PTopic struct {
	txTopic        *pubsub.Topic
	blockTopic     *pubsub.Topic
	stateTopic     *pubsub.Topic
	txSyncTopic    *pubsub.Topic
	validatorTopic *pubsub.Topic
	txStatTopic    *pubsub.Topic
}

const pubsubMaxSize = 1 << 22 // 4 MB

// SetGlobalNode 设置全局节点实例
func SetGlobalNode(node *Layer2Node) {
	nodeOnce.Do(func() {
		globalNode = node
	})
}

func NewLayer2Node(ctx context.Context, port int, bootstrapPeers []string, privKeyBytes []byte) (*Layer2Node, error) {
	var privateKey crypto.PrivKey
	var err error
	if len(bootstrapPeers) == 0 {
		// 获取私钥，先从命令行，其次文件，再其次生成并存入文件
		privateKey, err = getOrCreatePrivateKey(privKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("私钥处理失败: %w", err)
		}
	} else {
		// 如果是非启动节点，暂时生成新的 Ed25519 密钥对
		privateKey, _, err = crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, rand.Reader)
		if err != nil {
			return nil, err
		}
	}

	host, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
		libp2p.Identity(privateKey),
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
	)
	if err != nil {
		return nil, err
	}

	kdht, err := dht.New(ctx, host,
		dht.Mode(dht.ModeAutoServer),  // 使用自动服务器模式
		dht.ProtocolPrefix("/layer2"), // 添加自定义协议前缀，隔离网络
	)
	if err != nil {
		return nil, err
	}

	ps, err := pubsub.NewGossipSub(ctx, host,
		pubsub.WithPeerExchange(true),             // 启用节点交换
		pubsub.WithDirectPeers([]peer.AddrInfo{}), // 允许直接连接
		pubsub.WithMaxMessageSize(pubsubMaxSize),
	)
	if err != nil {
		return nil, err
	}
	// 打印节点ID，可以验证重启后ID是否相同
	fmt.Printf("节点ID: %s\n", host.ID().String())
	// 创建ping服务
	pingService := ping.NewPingService(host)

	node := &Layer2Node{
		host:           host,
		dht:            kdht,
		pubsub:         ps,
		ctx:            ctx,
		blockCache:     &sync.Map{},
		txPool:         &sync.Map{},
		txHistory:      &sync.Map{},
		pingService:    pingService,
		bootstrapPeers: bootstrapPeers, // 保存引导节点地址
		privateKey:     privateKey,
		publicKey:      privateKey.GetPublic(),
		minGasPrice:    0,
		isSequencer:    false,
		stateDB:        NewStateDB(),
	}
	// 设置连接回调
	host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			go node.handleNewPeer(conn.RemotePeer())
		},
	})
	// 初始化状态
	err = initState(node, bootstrapPeers)
	if err != nil {
		return nil, err
	}

	// 初始化排序器管理器
	consensusConfig := &consensus.ConsensusConfig{
		MinStakeAmount:   1000000,
		RotationInterval: 60 * time.Second,
		ValidatorQuorum:  3,
	}
	node.electionMgr = consensus.NewElectionManager(node.ctx, consensusConfig)
	return node, nil
}
func initState(node *Layer2Node, bootstrapPeers []string) error {
	// 为启动节点设置初始状态
	pub, err := types.PublicKeyToAddress(node.publicKey)
	if err != nil {
		return fmt.Errorf("私钥处理失败: %w", err)
	}
	// 为启动节点设置初始状态
	if len(bootstrapPeers) == 0 {
		// 如果是启动节点，设置一个较大的初始余额
		node.stateDB.UpdateBalance(pub, 1000000000000000000) // 1 ETH
		node.initialized = true
	} else {
		// 如果是普通节点，设置较小的初始余额用于支付gas费
		node.stateDB.UpdateBalance(pub, 1000000000000) // 0.001 ETH
	}

	// 初始化nonce为0
	node.stateDB.GetAccount(pub)

	// 设置全局节点实例
	SetGlobalNode(node)
	fmt.Printf("节点地址: %s, 初始余额: %d\n", pub, node.stateDB.GetBalance(pub))

	return nil
}

// 生成并保存私钥
func generateAndSaveKey() (crypto.PrivKey, error) {
	// 生成新的 Ed25519 密钥对
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, rand.Reader)
	if err != nil {
		return nil, err
	}

	// 将私钥转换为字节
	privBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, err
	}

	// 保存到文件
	return priv, os.WriteFile("node_key.bin", privBytes, 0600)
}

func getOrCreatePrivateKey(privKeyBytes []byte) (crypto.PrivKey, error) {
	// 如果提供了私钥字节，直接使用
	if len(privKeyBytes) > 0 {
		return crypto.UnmarshalPrivateKey(privKeyBytes)
	}

	// 尝试从文件读取私钥
	privKeyBytes, err := os.ReadFile("node_key.bin")
	if err == nil {
		return crypto.UnmarshalPrivateKey(privKeyBytes)
	}

	// 文件不存在，生成新密钥
	if errors.Is(err, os.ErrNotExist) {
		return generateAndSaveKey()
	}

	// 其他错误
	return nil, err
}

// 添加新方法处理新连接的节点
func (n *Layer2Node) handleNewPeer(p peer.ID) {
	// 发送ping
	result := <-n.pingService.Ping(n.ctx, p)
	if result.Error != nil {
		fmt.Printf("Ping to peer %s failed: %s\n", p.String(), result.Error)
		return
	}
	fmt.Printf("New peer connected %s, ping RTT = %s\n", p.String(), result.RTT)

	// 获取新节点的公钥
	pubKey := n.host.Peerstore().PubKey(p)

	// 生成节点地址
	peerAddress, err := types.PublicKeyToAddress(pubKey)
	if err != nil {
		fmt.Printf("Failed to generate address for peer %s: %s\n", p.String(), err)
		return
	}

	// 将新节点添加到状态数据库
	if _, exists := n.stateDB.accounts[peerAddress]; !exists {
		// 为新节点设置初始状态
		err = n.stateDB.SetAccountPublicKey(peerAddress, pubKey)
		if err != nil {
			fmt.Printf("Failed to generate acoount for peer %s: %s\n", peerAddress, err)
			return
		}
		n.stateDB.UpdateBalance(peerAddress, 1000000000000) // 0.001 ETH 初始余额
		fmt.Printf("Added new peer to state: %s with initial balance\n", peerAddress)
	}
}

// 添加获取节点地址的方法
func (n *Layer2Node) GetAddrs() []string {
	var addrs []string
	for _, addr := range n.host.Addrs() {
		addrs = append(addrs, fmt.Sprintf("%s/p2p/%s", addr, n.host.ID()))
	}
	return addrs
}

func (n *Layer2Node) InitConsensus(config *consensus.ConsensusConfig) error {
	// 加入验证者网络
	validator, err := n.electionMgr.RegisterValidator(
		n.publicKey,
		config.MinStakeAmount,
	)
	if err != nil {
		return fmt.Errorf("failed to join as validator: %w", err)
	}
	// 广播加入
	err = n.BroadcastValidatorJoin(*validator, consensus.ValidatorJoin)
	if err != nil {
		return err
	}
	// 开始管理器循环
	n.electionMgr.Start()
	return nil
}

func (n *Layer2Node) IsCurrentSequencer() bool {
	if n.electionMgr == nil {
		return false
	}
	return n.electionMgr.IsCurrentSequencer(n.publicKey)
}

func (n *Layer2Node) Start() error {
	// 启动DHT
	if err := n.dht.Bootstrap(n.ctx); err != nil {
		return fmt.Errorf("DHT bootstrap failed: %w", err)
	}

	// 当为启动节点时也需要开启循环
	if len(n.bootstrapPeers) == 0 {
		err := n.bootstrapStart()
		if err != nil {
			return err
		}
	}
	// 初始化状态记录节点
	InitStats(n, n.electionMgr.GetEth(), n.electionMgr)
	// 连接到引导节点
	if len(n.bootstrapPeers) > 0 {
		// 先等待同步
		n.isSyncing = true
		for _, addrStr := range n.bootstrapPeers {
			addr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				fmt.Printf("Invalid bootstrap peer address: %s\n", err)
				continue
			}

			peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
			if err != nil {
				fmt.Printf("Failed to parse bootstrap peer address: %s\n", err)
				continue
			}

			if err := n.host.Connect(n.ctx, *peerInfo); err != nil {
				fmt.Printf("Failed to connect to bootstrap peer: %s\n", err)
				continue
			}
			fmt.Printf("Connected to bootstrap peer: %s\n", peerInfo.ID)
		}
	}

	// 设置话题
	if err := n.setupTopics(); err != nil {
		return err
	}
	// 寻找网络中的其他节点
	go n.discoverPeers()

	// 如果是验证者节点，初始化共识
	if len(n.bootstrapPeers) > 0 {
		consensusConfig := &consensus.ConsensusConfig{
			MinStakeAmount:   1000000,
			RotationInterval: 60 * time.Second,
			ValidatorQuorum:  3,
		}
		time.Sleep(2 * time.Second)
		if err := n.InitConsensus(consensusConfig); err != nil {
			panic(err)
		}
	}
	// 启动状态监听
	go n.watchTxStatus()

	// 同步其他的节点世界状态
	if len(n.bootstrapPeers) > 0 {
		if err := n.syncStateFromPeers(); err != nil {
			fmt.Printf("Failed to sync state from peers: %s\n", err)
		}
	}
	return nil
}

func (n *Layer2Node) watchTxStatus() {
	ethClient := n.electionMgr.GetEth()
	statusChan := ethClient.GetStatusChannel()
	for event := range statusChan {
		pubTxs := make([]types.Transaction, 0)

		for _, eventTx := range event.Block.Transactions {
			if tx, ok := n.txHistory.Load(eventTx.Hash); ok {
				transaction := tx.(types.Transaction)
				transaction.StatLog.Status = event.Status
				transaction.StatLog.L1TxHash = event.L1TxHash
				transaction.StatLog.L1Timestamp = event.L1Timestamp

				// 更新交易历史
				n.txHistory.Store(eventTx.Hash, transaction)
				pubTxs = append(pubTxs, transaction)
			}
		}
		n.broadcastTxStat(pubTxs)
	}
}

func (n *Layer2Node) bootstrapStart() error {
	// 加载配置
	config, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		fmt.Printf("Warning: Failed to load config: %v\n", err)
		// 使用默认配置继续运行
		config.Ethereum = &ethereum.EthereumConfig{
			RPCURL:        "http://localhost:8545",
			GasLimit:      3000000,
			GasPrice:      20000000000,
			ConfirmBlocks: 2,
		}
	}
	ethClient, err := ethereum.NewEthereumClient(config.Ethereum)
	if err != nil {
		fmt.Printf("connect ethereum failed:" + err.Error())
		return err
	}
	n.electionMgr.SetEth(ethClient)
	n.electionMgr.Start()
	return nil
}

func (n *Layer2Node) discoverPeers() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	routingDiscovery := routing.NewRoutingDiscovery(n.dht)

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			// 广播自己的存在
			_, err := routingDiscovery.Advertise(n.ctx, "layer2-network")
			if err != nil {
				fmt.Printf("Failed to advertise: %s\n", err)
				continue
			}

			// 打印当前连接的节点数量
			connectedPeers := n.host.Network().Peers()
			fmt.Printf("Current connected peers: %d\n", len(connectedPeers))

			// 继续寻找新节点
			peersChan, err := routingDiscovery.FindPeers(n.ctx, "layer2-network")
			if err != nil {
				fmt.Printf("Failed to find peers: %s\n", err)
				continue
			}

			for peer := range peersChan {
				if peer.ID == n.host.ID() {
					continue
				}
				if n.host.Network().Connectedness(peer.ID) != network.Connected {
					if err := n.host.Connect(n.ctx, peer); err == nil {
						fmt.Printf("Connected to discovered peer: %s\n", peer.ID)
					}
				}
			}

		}
	}
}

func (n *Layer2Node) Host() host.Host {
	return n.host
}

func (n *Layer2Node) sendMessage(peer peer.ID, msg types.Message) error {
	stream, err := n.host.NewStream(n.ctx, peer, protocol.ID("/l2/1.0.0"))
	if err != nil {
		return err
	}
	defer stream.Close()

	return json.NewEncoder(stream).Encode(msg)
}
