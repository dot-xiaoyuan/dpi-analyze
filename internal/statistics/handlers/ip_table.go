package handlers

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/ip"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/provider"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strconv"
	"time"
)

func IpTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		queryPage := c.DefaultQuery("page", "1")
		querySize := c.DefaultQuery("pageSize", "20")

		page, _ := strconv.ParseInt(queryPage, 10, 64)
		pageSize, _ := strconv.ParseInt(querySize, 10, 64)

		zap.L().Info("query", zap.Int64("page", page), zap.Int64("pageSize", pageSize))

		now := time.Now()
		result, err := ip.TraversalIP(now.Add(-24*time.Hour).Unix(), now.Add(time.Hour).Unix(), page, pageSize)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
		}
		c.JSON(http.StatusOK, result)
	}
}

func IPInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Query("ip")

		conn, err := net.Dial("unix", "/tmp/capture.sock")
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}
		defer conn.Close()

		p := provider.Request{
			Action: "ip-detail",
			Data:   []byte(`{"ip":"` + ip + `"}`),
		}
		params, err := json.Marshal(p)
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
		}
		conn.Write(params)
		// 读取所有数据
		data, err := utils.ReadByConn(conn, 4096)
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}

		var res any
		if err := json.Unmarshal(data, &res); err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}

		if res == nil {
			c.JSON(400, gin.H{
				"message": "ip detail is empty",
			})
			return
		}

		c.JSON(200, res)
		return
	}
}
