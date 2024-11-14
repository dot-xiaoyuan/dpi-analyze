package controllers

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/users"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

// UserList 在线用户列表
func UserList() gin.HandlerFunc {
	return func(c *gin.Context) {
		pagination := utils.NewPagination(c.Query("page"), c.Query("pageSize"))

		condition := types.Condition{
			Min:      strconv.FormatInt(time.Now().Add(-24*time.Hour).Unix(), 10),
			Max:      strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			Page:     pagination.Page,
			PageSize: pagination.Limit,
		}
		var err error

		bytes, err := socket.SendUnixMessage(socket.UserList, condition)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		var res any
		_ = json.Unmarshal(bytes, &res)
		common.SuccessResponse(c, res)
	}
}

// UserEventsLog 用户事件日志
func UserEventsLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		pagination := utils.NewPagination(c.Query("page"), c.Query("pageSize"))

		condition := types.UserEventCondition{
			Page:      pagination.Page,
			PageSize:  pagination.Limit,
			Year:      c.DefaultQuery("year", time.Now().Format("06")),
			Month:     c.DefaultQuery("month", time.Now().Format("01")),
			SortField: c.DefaultQuery("field", "add_time"),
			OrderBy:   c.DefaultQuery("sort", "descend"),
		}

		var err error
		pagination.TotalCount, pagination.Result, err = users.UserEventQuery(condition)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		common.SuccessResponse(c, pagination)
	}
}
