package main

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

func main() {
	ctx := context.Background()

	// 创建引导节点的libp2p主机
	host, err := libp2p.New()
	if err != nil {
		log.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close()

	// 输出引导节点的地址
	fmt.Printf("Bootstrap node running with ID: %s\n", host.ID())
	for _, addr := range host.Addrs() {
		fmt.Printf("Address: %s/p2p/%s\n", addr, host.ID())
	}

	// 启动DHT
	kademliaDHT, err := dht.New(ctx, host)
	if err != nil {
		log.Fatalf("Failed to create DHT: %v", err)
	}

	// 引导DHT
	if err := kademliaDHT.Bootstrap(ctx); err != nil {
		log.Fatalf("DHT bootstrap error: %v", err)
	}

	// 设置简单的流处理器，处理其他节点的连接
	host.SetStreamHandler(protocol.ID("/myapp/1.0.0"), handleStream)

	// 保持程序运行
	select {}
}

func handleStream(s network.Stream) {
	fmt.Println("Received connection from:", s.Conn().RemotePeer())
	s.Close()
}
