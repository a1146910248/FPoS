package main

//
//import (
//	"context"
//	"fmt"
//	"github.com/libp2p/go-libp2p"
//	"github.com/libp2p/go-libp2p/core/host"
//	"github.com/libp2p/go-libp2p/core/network"
//	"github.com/libp2p/go-libp2p/core/peer"
//	"github.com/libp2p/go-libp2p/core/protocol"
//	"github.com/multiformats/go-multiaddr"
//	"log"
//	"time"
//)
//
//const gossipTopic = "gossip-topic"
//
//// 初始化节点并加入 Gossip 网络
//func setupNode(ctx context.Context, listenPort int) (*pubsub.PubSub, *host.Host, *pubsub.Topic, error) {
//	// 创建节点
//	host, err := libp2p.New(
//		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort)),
//	)
//	if err != nil {
//		return nil, nil, nil, fmt.Errorf("failed to create libp2p host: %v", err)
//	}
//
//	// 创建 PubSub 系统
//	ps, err := pubsub.NewFloodSub(ctx, host)
//	if err != nil {
//		return nil, nil, nil, fmt.Errorf("failed to create floodsub: %v", err)
//	}
//
//	// 加入 Gossip 主题
//	topic, err := ps.Join(gossipTopic)
//	if err != nil {
//		return nil, nil, nil, fmt.Errorf("failed to join gossip topic: %v", err)
//	}
//
//	host.SetStreamHandler(protocol.ID(gossipTopic), func(s network.Stream) {
//		fmt.Printf("Node %s received new stream\n", host.ID().Pretty())
//	})
//
//	fmt.Printf("Node %s started and listening on port %d\n", host.ID().Pretty(), listenPort)
//	return ps, &host, topic, nil
//}
//
//// 发送 Gossip 消息
//func sendGossipMessage(ctx context.Context, topic *pubsub.Topic, message string) error {
//	return topic.Publish(ctx, []byte(message))
//}
//
//// 监听 Gossip 消息
//func listenGossipMessages(ctx context.Context, sub *pubsub.Subscription) {
//	for {
//		msg, err := sub.Next(ctx)
//		if err != nil {
//			log.Println("Error receiving message:", err)
//			return
//		}
//		fmt.Printf("Received message: %s\n", string(msg.Data))
//	}
//}
//
//func main() {
//	ctx := context.Background()
//
//	// 设置第一个节点
//	ps1, host1, topic1, err := setupNode(ctx, 10000)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 设置第二个节点
//	ps2, host2, topic2, err := setupNode(ctx, 10001)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 使两个节点互联
//	addr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/10000/p2p/%s", (*host1).ID().Pretty()))
//	(*host2).Connect(ctx, peer.AddrInfo{ID: (*host1).ID(), Addrs: []multiaddr.Multiaddr{addr}})
//
//	// 订阅 Gossip 消息
//	sub1, err := ps1.Subscribe(gossipTopic)
//	if err != nil {
//		log.Fatal(err)
//	}
//	sub2, err := ps2.Subscribe(gossipTopic)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 启动 Gossip 监听器
//	go listenGossipMessages(ctx, sub1)
//	go listenGossipMessages(ctx, sub2)
//
//	// 发送 Gossip 消息
//	time.Sleep(1 * time.Second) // 等待连接建立
//	if err := sendGossipMessage(ctx, topic1, "Hello from Node 1"); err != nil {
//		log.Println("Error sending message:", err)
//	}
//	if err := sendGossipMessage(ctx, topic2, "Hello from Node 2"); err != nil {
//		log.Println("Error sending message:", err)
//	}
//
//	select {} // 保持程序运行
//}
