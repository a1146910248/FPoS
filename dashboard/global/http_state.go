package global

import (
	"FPoS/pkg/logging"
	"github.com/gin-gonic/gin"
	"net/http"
)

var logger = logging.GetLogger()

func Success(c *gin.Context, data interface{}) {
	result := map[string]interface{}{
		"ok":   1,
		"msg":  "操作成功",
		"data": data,
	}

	//b, err := json.Marshal(result)
	//if err != nil {
	//	logger.Error(err)
	//	return
	//}

	// 没写前端，先不要
	//// 对http缓存进行验证，没有更改则返回304
	//oldEtag := c.Request.Header.Get("If-None-Match")
	//// 对弱匹配做处理，得到真正对Etag
	//if strings.HasPrefix(oldEtag, "W/") {
	//	oldEtag = oldEtag[2:]
	//}
	//newEtag := goutils.Md5Buf(b)
	//if oldEtag == newEtag {
	//	c.Status(http.StatusNotModified)
	//	return
	//}

	// TODO 加到nosql 缓存？

	//c.Writer.Header().Add("Etag", newEtag)

	// 跨域请求，有callback字段采用 JsonP方式，无则普通Json
	c.Header("Content-Type", "application/json")
	callback := c.Query("callback")
	if callback != "" {
		c.JSONP(http.StatusOK, result)
		return
	}
	c.JSON(http.StatusOK, result)
}

func Fail(c *gin.Context, msg string) {
	result := map[string]interface{}{
		"ok":  0,
		"msg": msg,
	}
	logger.Error("operate fail:", result)

	c.JSON(http.StatusOK, result)
}
