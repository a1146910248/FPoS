package router

import (
	. "FPoS/dashboard/global"
	. "FPoS/dashboard/model"
	"FPoS/p2p"
	. "FPoS/pkg/logging"
	"github.com/gin-gonic/gin"
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
