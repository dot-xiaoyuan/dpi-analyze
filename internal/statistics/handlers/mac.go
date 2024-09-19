package handlers

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/gin-gonic/gin"
	"net"
)

func Mac() gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := net.Dial("unix", "/tmp/capture.sock")
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}
		defer conn.Close()

		conn.Write([]byte("ethernet"))
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}

		var macMap map[string][]capture.Ethernet
		if err := json.Unmarshal(buf[:n], &macMap); err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}

		if len(macMap) == 0 {
			c.JSON(400, gin.H{
				"message": "TTL Map is empty",
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
