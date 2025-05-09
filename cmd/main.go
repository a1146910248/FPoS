package main

import (
	"FPoS/config"
	"FPoS/core/ethereum"
	"FPoS/dashboard/app"
	"FPoS/p2p"
	"FPoS/pkg/logging"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

var logger *logging.Logger

func init() {
	logger = logging.GetLogger()
	// 初始化配置文件
	InitConfig()
}

func main() {
	ctx := context.Background()
	isBootstrap := os.Getenv("BOOTSTRAP") == "true"
	enableTx := os.Getenv("ENABLE_TX") == "true" // 控制是否启用定时交易

	var privKeyBytes []byte
	// 首先尝试从环境变量获取私钥
	if envKey := os.Getenv("NODE_PRIVATE_KEY"); envKey != "" {
		privKeyBytes = []byte(envKey)
	}
	if privKeyBytes != nil {
		fmt.Println(privKeyBytes)
	}

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

	// 添加DAC节点相关的命令行参数
	//isDACNode := flag.Bool("dac", false, "启动为DAC节点")
	isDACNode := true
	//dacStake := flag.Uint64("dacstake", 10000, "设置DAC节点的质押金额")
	dacStake := flag.Uint64("dacstake", 200000000, "设置DAC节点的质押金额")

	// 解析命令行参数
	flag.Parse()

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
		// 初始化gin后端
		app.App.Init()
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

		// 添加DAC节点启动逻辑
		if isDACNode {
			// 注册为DAC成员并启动DAC服务
			go func() {
				// 等待节点初始化完成
				time.Sleep(2 * time.Second)

				logger.Info("正在启动DAC节点服务...")

				// 注册为ethClient
				err := node.RegisterEthClient(config)
				if err != nil {
					logger.Error("注册DAC成员失败:", err)
					return
				}

				if err := node.Start(); err != nil {
					panic(err)
				}
				// 注册为DAC成员
				err = node.RegisterAsDACMember(*dacStake)
				if err != nil {
					logger.Error("注册DAC成员失败:", err)
					return
				}

				logger.Info("DAC节点服务已启动")

				// 创建定时器，定期打印DAC状态
				//ticker := time.NewTicker(30 * time.Second)
				//defer ticker.Stop()

				//for {
				//	select {
				//	case <-ticker.C:
				//		// 打印DAC状态信息
				//		printDACStatus(node)
				//	case <-ctx.Done():
				//		return
				//	}
				//}
			}()
		} else {
			//if isSequencer {
			// 启动排序器节点
			sequencer, _ := p2p.NewSequencer(node, config)
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
		}

		fmt.Printf("Node started with bootstrap: %s\n", bootstrapAddr)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

// 打印DAC节点状态
func printDACStatus(node *p2p.Layer2Node) {
	dacMgr := node.GetDACManager()
	if dacMgr == nil {
		logger.Warn("DAC管理器未初始化")
		return
	}

	state := dacMgr.GetState()
	logger.Info("======= DAC状态 =======")
	logger.Infof("当前活跃DAC成员: %d", len(state.CurrentMembers))
	for i, member := range state.CurrentMembers {
		logger.Infof("  成员 %d: %s", i+1, member)
	}
	logger.Infof("当前任期: %d", state.CurrentTerm)
	logger.Infof("下次轮换时间: %v", state.NextRotationTime)
	logger.Infof("总成员数量: %d", len(state.Members))
	logger.Info("======================")
}

const (
	CONFIG_PATH = "config/"
)

func InitConfig() {
	// 环境变量指定加载环境
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}
	// 设置配置文件名和路径
	viper.SetConfigName(fmt.Sprintf("env.%s", env))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(CONFIG_PATH)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	logger.Info("配置文件读取成功！！")
}
