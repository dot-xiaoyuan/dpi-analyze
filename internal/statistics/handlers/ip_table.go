package handlers

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
		result, err := capture.TraversalIP(now.Add(-24*time.Hour).Unix(), now.Add(time.Hour).Unix(), page, pageSize)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
		}
		c.JSON(http.StatusOK, result)
	}
}
