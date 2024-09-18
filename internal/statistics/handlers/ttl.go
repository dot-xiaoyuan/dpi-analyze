package handlers

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze"
	"github.com/gin-gonic/gin"
	"net"
)

func TTL() gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := net.Dial("unix", "/tmp/capture.sock")
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}
		defer conn.Close()

		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}

		var ttlMap map[string]analyze.Internet
		if err := json.Unmarshal(buf[:n], &ttlMap); err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "OK",
			"data":    ttlMap,
		})
		return
	}
}
