package controllers

import (
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Dashboard() gin.HandlerFunc {
	return func(c *gin.Context) {
		bytes, err := socket.SendMessage("dashboard")
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
		}
		common.SuccessResponse(c, bytes)
	}
}
