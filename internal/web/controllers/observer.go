package controllers

import (
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/observer"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/provider"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func ObserverHandler(observer interface {
	Traversal(condition provider.Condition) (int64, interface{}, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		pagination := utils.NewPagination(c.Query("page"), c.Query("pageSize"))

		condition := provider.Condition{
			Min:      strconv.FormatInt(time.Now().Add(-24*time.Hour).Unix(), 10),
			Max:      strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			Page:     pagination.Page,
			PageSize: pagination.Limit,
		}
		var err error

		pagination.TotalCount, pagination.Data, err = observer.Traversal(condition)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
		}
		common.SuccessResponse(c, pagination)
	}
}

func ObserverTTL() gin.HandlerFunc {
	return ObserverHandler(observer.TTLObserver)
}

func ObserverMac() gin.HandlerFunc {
	return ObserverHandler(observer.MacObserver)
}

func ObserverUa() gin.HandlerFunc {
	return ObserverHandler(observer.UaObserver)
}
