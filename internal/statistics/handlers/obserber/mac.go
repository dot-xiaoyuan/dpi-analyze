package obserber

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/statistics/common"
	provider2 "github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/provider"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/provider"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/gin-gonic/gin"
	"net"
	"strconv"
	"time"
)

func ObserverMac() gin.HandlerFunc {
	return func(c *gin.Context) {
		var condition provider2.Condition
		condition.Table = types.ZSetObserverMac
		condition.Min = strconv.FormatInt(time.Now().Add(-24*time.Hour).Unix(), 10)
		condition.Max = strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10)

		condition.Page, _ = strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
		condition.PageSize, _ = strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

		// proxy socket
		conn, err := net.Dial("unix", "/tmp/capture.sock")
		if err != nil {
			common.Error(err, c)
			return
		}
		defer conn.Close()

		jsonParams, _ := json.Marshal(condition)
		p := provider.Request{
			Action: "observer",
			Data:   jsonParams,
		}
		params, err := json.Marshal(p)
		if err != nil {
			common.Error(err, c)
			return
		}
		conn.Write(params)
		// 读取所有数据
		data, err := utils.ReadByConn(conn, 4096)
		if err != nil {
			common.Error(err, c)
			return
		}

		var res any
		if err = json.Unmarshal(data, &res); err != nil {
			common.Error(err, c)
			return
		}

		if res == nil {
			common.Error(err, c)
			return
		}

		c.JSON(200, res)
		return
	}
}
