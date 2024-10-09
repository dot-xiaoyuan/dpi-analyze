package controllers

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Dashboard 仪表盘
func Dashboard() gin.HandlerFunc {
	return func(c *gin.Context) {
		bytes, err := socket.SendMessage(socket.Message{
			Type: socket.Dashboard,
		})
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
		}
		var res any
		_ = json.Unmarshal(bytes, &res)
		common.SuccessResponse(c, res)
	}
}
