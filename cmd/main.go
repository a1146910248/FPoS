package main

import (
	"FPoS/p2p"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

//	func main() {
//		ctx := context.Background()
//
//		// 创建Layer2节点
//		node, err := p2p.NewLayer2Node(ctx, 6666)
//		if err != nil {
//			panic(err)
//		}
//
//		// 设置交易处理器
//		node.SetTransactionHandler(func(tx types.Transaction) bool {
//			// 实现交易验证逻辑
//			return true
//		})
//
//		// 设置区块处理器
//		node.SetBlockHandler(func(block types.Block) bool {
//			// 实现区块验证逻辑
//			return true
//		})
//
//		// 启动节点
//		if err := node.Start(); err != nil {
//			panic(err)
//		}
//
//		fmt.Printf("Layer2 节点已启动，地址: %s\n", node.Host().ID())
//
//		// 等待中断信号
//		sigChan := make(chan os.Signal, 1)
//		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
//		<-sigChan
//	}
func main() {
	ctx := context.Background()
	isBootstrap := os.Getenv("BOOTSTRAP") == "true"

	if isBootstrap {
		// 启动引导节点
		node, err := p2p.NewLayer2Node(ctx, 6666, nil)
		if err != nil {
			panic(err)
		}
		if err := node.Start(); err != nil {
			panic(err)
		}

		// 保存引导节点信息
		addr := node.GetAddrs()[0]
		if err := p2p.SaveBootstrapInfo(addr); err != nil {
			panic(err)
		}
		fmt.Printf("Bootstrap node started: %s\n", addr)
	} else {
		// 读取引导节点信息
		bootstrapAddr, err := p2p.LoadBootstrapInfo()
		if err != nil {
			panic(err)
		}

		// 启动普通节点
		node, err := p2p.NewLayer2Node(ctx, 6667, []string{bootstrapAddr})
		if err != nil {
			panic(err)
		}
		if err := node.Start(); err != nil {
			panic(err)
		}
		fmt.Printf("Node started with bootstrap: %s\n", bootstrapAddr)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
