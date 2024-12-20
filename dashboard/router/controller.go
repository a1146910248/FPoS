package router

import (
	. "FPoS/dashboard/global"
	. "FPoS/dashboard/model"
	"FPoS/p2p"
	. "FPoS/pkg/logging"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	logger *Logger
)

type HttpController struct{}

func init() {
	logger = GetLogger()
}

func (t HttpController) GetStats(c *gin.Context) {
	stats := p2p.GetStats()
	l1Blocks, l1Balance := stats.GetL1Stats()
	l2Blocks, l2TPS := stats.GetL2Stats()

	chainStats := ChainStats{
		CurrentTPS:  stats.GetCurrentTPS(),
		PeakTPS:     stats.GetPeakTPS(),
		TotalTx:     stats.GetTotalTransactions(),
		BlockHeight: stats.GetCurrentHeight(),
		ActiveUsers: stats.GetActiveUsers(),
		L1Blocks:    l1Blocks,
		L2Blocks:    l2Blocks,
		L1Balance:   l1Balance,
		L2TPS:       l2TPS,
	}

	Success(c, chainStats)
}

//func (t HttpController) GetTransactions(c *gin.Context) {
//	limit := 20
//	if limitStr := c.Query("limit"); limitStr != "" {
//		if l, err := strconv.Atoi(limitStr); err == nil {
//			limit = l
//		}
//	}
//
//	txs := n.GetLatestTransactions(limit)
//	Success(c, txs)
//}
//
//func (t HttpController) GetBlocks(c *gin.Context) {
//	limit := 20
//	if limitStr := c.Query("limit"); limitStr != "" {
//		if l, err := strconv.Atoi(limitStr); err == nil {
//			limit = l
//		}
//	}
//
//	blocks := node.GetLatestBlocks(limit)
//	Success(c, blocks)
//}
//
//func (t HttpController) StreamUpdates(c *gin.Context) {
//	upgrader := websocket.Upgrader{
//		CheckOrigin: func(r *http.Request) bool {
//			return true
//		},
//	}
//
//	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
//	if err != nil {
//		logger.Error("websocket upgrade failed:", err)
//		return
//	}
//	defer ws.Close()
//
//	// 创建更新通道
//	updates := make(chan interface{})
//	defer close(updates)
//
//	// 订阅区块链事件
//	node.SubscribeEvents(updates)
//
//	for {
//		select {
//		case update := <-updates:
//			if err := ws.WriteJSON(update); err != nil {
//				logger.Error("websocket write failed:", err)
//				return
//			}
//		case <-c.Done():
//			return
//		}
//	}
//}

func (t HttpController) StreamUpdates(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("websocket upgrade failed:", err)
		return
	}
	defer ws.Close()

	// 创建更新通道
	updates := make(chan *ChainStats)
	defer close(updates)

	// 启动定时器，每秒获取最新数据
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// 心跳检测
	go func() {
		pingTicker := time.NewTicker(time.Second * 30)
		defer pingTicker.Stop()

		for {
			select {
			case <-pingTicker.C:
				if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-c.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			// 获取最新统计数据
			stats := p2p.GetStats()
			l1Blocks, l1Balance := stats.GetL1Stats()
			l2Blocks, l2TPS := stats.GetL2Stats()

			chainStats := ChainStats{
				CurrentTPS:  stats.GetCurrentTPS(),
				PeakTPS:     stats.GetPeakTPS(),
				TotalTx:     stats.GetTotalTransactions(),
				BlockHeight: stats.GetCurrentHeight(),
				ActiveUsers: stats.GetActiveUsers(),
				L1Blocks:    l1Blocks,
				L2Blocks:    l2Blocks,
				L1Balance:   l1Balance,
				L2TPS:       l2TPS,
			}

			// 发送更新
			if err := ws.WriteJSON(chainStats); err != nil {
				logger.Error("websocket write failed:", err)
				return
			}

		case <-c.Done():
			return
		}
	}
}
func (t HttpController) GetTransactions(c *gin.Context) {
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			offset = (page - 1) * limit
		}
	}

	// 获取原始交易数据
	coreTxs := p2p.GetTransactions(limit, offset)
	total := p2p.GetTotalTransactions()

	// 转换为dashboard的Transaction类型
	txs := make([]Transaction, len(coreTxs))
	for i, tx := range coreTxs {
		txs[i] = Transaction{
			Hash:      tx.Hash,
			From:      tx.From,
			To:        tx.To,
			Value:     tx.Value,
			Nonce:     tx.Nonce,
			GasPrice:  tx.GasPrice,
			GasLimit:  tx.GasLimit,
			GasUsed:   tx.GasUsed,
			Timestamp: tx.Timestamp,
			Status:    tx.StatLog.Status,
			BlockHash: tx.StatLog.BlockHash,
		}
	}

	Success(c, TransactionList{
		Total: int64(total),
		List:  txs,
	})
}
