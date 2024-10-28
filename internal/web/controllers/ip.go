package controllers

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"log"
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
			return
		}
		common.SuccessResponse(c, result)
	}
}

// IPDetail IP 详情
func IPDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Query("ip")
		bytes, err := socket.SendUnixMessage(socket.IPDetail, ip)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		var res models.IPDetail
		_ = json.Unmarshal(bytes, &res)
		res.Features, err = getFeature(ip)
		common.SuccessResponse(c, res)
	}
}

func getFeature(ip string) (any, error) {
	collection := mongo.GetMongoClient().Database(types.Features).Collection(types.OnlineUsersFeature)

	// 查询条件
	filter := bson.D{{"ip", ip}}
	// 执行查询
	cursor, err := collection.Find(mongo.Context, filter)
	if err != nil {
		return nil, err
	}
	var results []types.FeatureSet
	if err = cursor.All(mongo.Context, &results); err != nil {
		log.Fatal(err)
	}
	var charts []types.Chart
	for _, result := range results {
		charts = append(charts, result.Total...)
	}
	return charts, nil
}
