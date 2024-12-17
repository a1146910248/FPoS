package app

import (
	"FPoS/dashboard/router"
	"FPoS/pkg/logging"
	log "FPoS/pkg/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"os"
	"sync"
	"time"
)

var logger = logging.GetLogger()

type app struct {
	Name    string
	Version string
	Date    time.Time

	// 启动时间
	LaunchTime time.Time
	Uptime     time.Duration

	// 环境：开发环境dev、测试环境test、生产环境Pro
	Env string

	Host string
	Port string

	locker sync.Mutex
}

var App = &app{}

const (
	DEV  = "dev"
	TEST = "test"
	PRO  = "pro"
)

func init() {
	App.Name = os.Args[0]
	App.Version = viper.GetString("global.VERSION")
	App.LaunchTime = time.Now()
	App.Env = os.Getenv("ENV")

	fileInfo, err := os.Stat(os.Args[0])
	if err != nil {
		panic(err)
	}

	App.Date = fileInfo.ModTime()
}

// gin 初始化
func (t *app) Init() {
	r := initGin()

	port := ":" + viper.GetString("global.PORT")
	if err := r.Run(port); err != nil {
		logger.Error("gin init failed!")
		// 退出程序，可能需要一些清理工作
		os.Exit(1)
	}

}

func initGin() *gin.Engine {
	// 创建 Gin 引擎
	r := gin.Default()
	// 允许跨域请求
	r.Use(cors.Default())
	// 初始化日志
	r.Use(log.NewLogger())
	// Jwt校验

	// 注册路由,生产环境需要加上/api前缀
	var api *gin.RouterGroup
	if App.Env == "prod" {
		api = r.Group("/dashboardApi")
	} else {
		api = r.Group("")
	}

	new(router.HttpController).RegisterRouter(api)
	return r
}
