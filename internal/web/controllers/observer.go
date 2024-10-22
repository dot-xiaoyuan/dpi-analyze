package controllers

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/provider"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func ObserverHandler(property types.Property) gin.HandlerFunc {
	return func(c *gin.Context) {
		pagination := utils.NewPagination(c.Query("page"), c.Query("pageSize"))

		condition := provider.Condition{
			Min:      strconv.FormatInt(time.Now().Add(-24*time.Hour).Unix(), 10),
			Max:      strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			Page:     pagination.Page,
			PageSize: pagination.Limit,
			Type:     property,
		}
		var err error

		bytes, err := socket.SendUnixMessage(socket.Observer, condition)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		var res any
		_ = json.Unmarshal(bytes, &res)
		common.SuccessResponse(c, res)
	}
}

func ObserverTTL() gin.HandlerFunc {
	return ObserverHandler(types.TTL)
}

func ObserverMac() gin.HandlerFunc {
	return ObserverHandler(types.Mac)
}

func ObserverUa() gin.HandlerFunc {
	return ObserverHandler(types.UserAgent)
}
