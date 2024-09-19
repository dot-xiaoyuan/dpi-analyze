package handlers

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/gin-gonic/gin"
	"net"
)

func IpTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := net.Dial("unix", "/tmp/capture.sock")
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}
		defer conn.Close()

		p := capture.Params{
			Action: "iptables",
			Offset: 0,
			Limit:  20,
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

		var macMap map[string]capture.IPActivityLogs
		if err := json.Unmarshal(data, &macMap); err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}

		if len(macMap) == 0 {
			c.JSON(400, gin.H{
				"message": "IP Map is empty",
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "OK",
			"data":    macMap,
		})
		return
	}
}
