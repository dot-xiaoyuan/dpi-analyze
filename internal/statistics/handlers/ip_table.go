package handlers

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func IpTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		queryStart := c.DefaultQuery("page", "0")
		//queryEnd := c.DefaultQuery("end", "10")
		querySize := c.DefaultQuery("pageSize", "20")

		s, _ := strconv.ParseInt(queryStart, 10, 64)
		//e, _ := strconv.ParseInt(queryEnd, 10, 64)
		pageSize, _ := strconv.ParseInt(querySize, 10, 64)

		start := (s - 1) * pageSize
		end := start + pageSize - 1
		zap.L().Info("query", zap.Int64("start", start), zap.Int64("end", end))

		result := capture.TraversalIP(start, end)
		c.JSON(http.StatusOK, result)
	}
}
