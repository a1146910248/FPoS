package router

import "github.com/gin-gonic/gin"

func (t HttpController) RegisterRouter(e *gin.RouterGroup) {
	dashboard := e.Group("/dashboard")
	{
		dashboard.GET("/stats", t.GetStats)
		dashboard.GET("/ws", t.StreamUpdates)
		dashboard.GET("/transactions", t.GetTransactions)
		//dashboard.GET("/blocks", t.GetBlocks)
	}
}
