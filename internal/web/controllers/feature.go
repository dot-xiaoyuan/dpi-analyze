package controllers

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"
	"github.com/gin-gonic/gin"
	"net/http"
)

func FeatureLibrary() gin.HandlerFunc {
	return func(c *gin.Context) {
		bytes, err := socket.SendUnixMessage(socket.FeatureLibrary, nil)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		var res any
		_ = json.Unmarshal(bytes, &res)
		common.SuccessResponse(c, res)
		return
	}
}
