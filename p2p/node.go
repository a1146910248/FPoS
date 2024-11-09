package p2p

import (
	"FPoS/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
	"sync"
	"time"
)

type Layer2Node struct {
	host           host.Host
	dht            *dht.IpfsDHT
	pubsub         *pubsub.PubSub
	txTopic        *pubsub.Topic
	blockTopic     *pubsub.Topic
	stateTopic     *pubsub.Topic
	ctx            context.Context
	blockCache     *sync.Map
	txPool         *sync.Map
	stateRoot      string
	latestBlock    uint64
	handlers       types.Handlers
	pingService    *ping.PingService
	bootstrapPeers []string
	mu             sync.RWMutex
}

func NewLayer2Node(ctx context.Context, port int, bootstrapPeers []string) (*Layer2Node, error) {
	if len(bootstrapPeers) != 0 {
		for _, str := range bootstrapPeers {
			println("peer is :" + str)
		}
	}
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
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
	)
	if err != nil {
		return nil, err
	}

	// 创建ping服务
	pingService := ping.NewPingService(host)

	node := &Layer2Node{
		host:           host,
		dht:            kdht,
		pubsub:         ps,
		ctx:            ctx,
		blockCache:     &sync.Map{},
		txPool:         &sync.Map{},
		pingService:    pingService,
		bootstrapPeers: bootstrapPeers, // 保存引导节点地址
	}
	// 设置连接回调
	host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			go node.handleNewPeer(conn.RemotePeer())
		},
	})
	return node, nil
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
}

// 添加获取节点地址的方法
func (n *Layer2Node) GetAddrs() []string {
	var addrs []string
	for _, addr := range n.host.Addrs() {
		addrs = append(addrs, fmt.Sprintf("%s/p2p/%s", addr, n.host.ID()))
	}
	return addrs
}

func (n *Layer2Node) Start() error {
	// 启动DHT
	if err := n.dht.Bootstrap(n.ctx); err != nil {
		return fmt.Errorf("DHT bootstrap failed: %w", err)
	}

	// 连接到引导节点
	if len(n.bootstrapPeers) > 0 {
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

	return nil
}

func (n *Layer2Node) discoverPeers() {
	ticker := time.NewTicker(time.Second * 5)
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
