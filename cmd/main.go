package main

import (
	"FPoS/p2p"
	"FPoS/types"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()
	isBootstrap := os.Getenv("BOOTSTRAP") == "true"
	enableTx := os.Getenv("ENABLE_TX") == "true" // 控制是否启用定时交易
	//isSequencer := os.Getenv("SEQUENCER") == "true" // 是否为排序器节点

	var privKeyBytes []byte
	// 首先尝试从环境变量获取私钥
	if envKey := os.Getenv("NODE_PRIVATE_KEY"); envKey != "" {
		privKeyBytes = []byte(envKey)
	}
	if privKeyBytes != nil {
		fmt.Println(privKeyBytes)
	}

	if isBootstrap {
		// 启动引导节点
		node, err := p2p.NewLayer2Node(ctx, 6666, nil, privKeyBytes)
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
		node, err := p2p.NewLayer2Node(ctx, 0, []string{bootstrapAddr}, privKeyBytes)
		if err != nil {
			panic(err)
		}
		// TODO 如果是排序器节点，之后会加入共识

		//if isSequencer {
		// 启动排序器节点
		sequencer := p2p.NewSequencer(node)
		//}
		if err := node.Start(); err != nil {
			panic(err)
		}
		if enableTx {
			// 启动定时交易
			node.StartPeriodicTransaction()
			fmt.Println("已启动定时交易发送")
		}
		sequencer.Start()
		fmt.Println("Sequencer node started")
		fmt.Printf("Node started with bootstrap: %s\n", bootstrapAddr)
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

func setHandler(node *p2p.Layer2Node) {
	// 设置交易处理器
	node.SetTransactionHandler(func(tx types.Transaction) bool {
		// 实现交易验证逻辑
		return true
	})

	// 设置区块处理器
	node.SetBlockHandler(func(block types.Block) bool {
		// 实现区块验证逻辑
		return true
	})
}
