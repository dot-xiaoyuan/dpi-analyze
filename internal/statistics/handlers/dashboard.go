package handlers

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/provider"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/gin-gonic/gin"
	"net"
)

// 仪表盘

func Dashboard() gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := net.Dial("unix", "/tmp/capture.sock")
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}
		defer conn.Close()

		p := provider.Request{
			Action: "dashboard",
			Data:   []byte(`{"offset":0,"limit":20}`),
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
				"message": "dashboard is empty",
			})
			return
		}

		c.JSON(200, res)
		return
	}
}
