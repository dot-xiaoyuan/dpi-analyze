package controllers

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

// IPList IP 列表
func IPList() gin.HandlerFunc {
	return func(c *gin.Context) {
		queryPage := c.DefaultQuery("page", "1")
		querySize := c.DefaultQuery("pageSize", "20")

		page, _ := strconv.ParseInt(queryPage, 10, 64)
		pageSize, _ := strconv.ParseInt(querySize, 10, 64)

		zap.L().Info("query", zap.Int64("page", page), zap.Int64("pageSize", pageSize))

		now := time.Now()
		result, err := member.TraversalIP(now.Add(-24*time.Hour).Unix(), now.Add(time.Hour).Unix(), page, pageSize)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
		}
		common.SuccessResponse(c, result)
	}
}

// IPDetail IP 详情
func IPDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Query("ip")
		m := socket.Message{
			Type:   socket.IPDetail,
			Params: ip,
		}
		bytes, err := socket.SendMessage(m)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
		}
		var res any
		_ = json.Unmarshal(bytes, &res)
		common.SuccessResponse(c, res)
	}
}
