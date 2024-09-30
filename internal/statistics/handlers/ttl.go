package handlers

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/layers"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/provider"
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

		p := provider.Request{
			Action: "internet",
			Data:   []byte(`{"offset":0,"limit":20}`),
		}
		params, err := json.Marshal(p)
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
		}
		conn.Write(params)
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}

		var ttlMap map[string][]layers.Internet
		if err := json.Unmarshal(buf[:n], &ttlMap); err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}

		if len(ttlMap) == 0 {
			c.JSON(400, gin.H{
				"message": "TTL Map is empty",
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
