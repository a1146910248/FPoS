package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/network"
	maddr "github.com/multiformats/go-multiaddr"
	"log"
)

func main() {
	// 从命令行接收引导节点地址
	bootstrapAddr := flag.String("bootstrap", "", "Bootstrap peer multiaddress")
	flag.Parse()

	if *bootstrapAddr == "" {
		log.Fatal("Please provide the bootstrap peer address using -bootstrap flag")
	}

	// 创建上下文
	ctx := context.Background()

	// 创建libp2p主机
	host, err := libp2p.New()
	if err != nil {
		log.Fatalf("Failed to create libp2p host: %v", err)
	}
	defer host.Close()

	fmt.Printf("Non-bootstrap node running with ID: %s\n", host.ID())
	for _, addr := range host.Addrs() {
		fmt.Printf("Address: %s/p2p/%s\n", addr, host.ID())
	}

	// 解析引导节点地址
	addr, err := maddr.NewMultiaddr(*bootstrapAddr)
	if err != nil {
		log.Fatalf("Invalid bootstrap address: %v", err)
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		log.Fatalf("Failed to parse bootstrap peer info: %v", err)
	}

	// 启动DHT
	kademliaDHT, err := dht.New(ctx, host, dht.BootstrapPeers(*peerInfo))
	if err != nil {
		log.Fatalf("Failed to create DHT: %v", err)
	}

	// 引导DHT
	if err := kademliaDHT.Bootstrap(ctx); err != nil {
		log.Fatalf("DHT bootstrap error: %v", err)
	}

	// 设置流处理器
	host.SetStreamHandler(protocol.ID("/myapp/1.0.0"), handleStream)

	// 启动一个协程搜索其他节点
	go discoverPeers(ctx, kademliaDHT)

	// 保持程序运行
	select {}
}

// 处理传入的数据流
func handleStream(s network.Stream) {
	fmt.Println("Received a connection from:", s.Conn().RemotePeer())
	s.Close()
}

func discoverPeers(ctx context.Context, dht *dht.IpfsDHT) {
	peerChan, err := dht.FindPeers(ctx, "rendezvous")
	if err != nil {
		log.Fatalf("Peer discovery error: %v", err)
	}
	for peer := range peerChan {
		fmt.Println("Found peer:", peer)
	}
}
