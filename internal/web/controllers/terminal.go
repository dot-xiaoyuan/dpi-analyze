package controllers

import (
	"context"
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/resolve"
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket/models"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

// Identification 终端列表
func Identification() gin.HandlerFunc {
	return func(c *gin.Context) {
		var jsonData struct {
			Page     int64 `json:"page"`
			PageSize int64 `json:"pageSize"`
		}
		err := c.BindJSON(&jsonData)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		zap.L().Info("query", zap.Int64("page", jsonData.Page), zap.Int64("pageSize", jsonData.PageSize))
		now := time.Now()
		result, err := member.TraversalIP(now.Add(-24*time.Hour).Unix(), now.Add(time.Hour).Unix(), jsonData.Page, jsonData.PageSize)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		common.SuccessResponse(c, result)
		return
	}
}

// UseragentRecord ua识别
func UseragentRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		var query QueryData
		if err := c.BindJSON(&query); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
			return
		}
		zap.L().Debug("query", zap.Any("params", query))

		pagination := utils.NewPagination(strconv.Itoa(query.Page), strconv.Itoa(query.PageSize))

		collection := query.Collection
		// skip
		skip := (pagination.Page - 1) * pagination.Limit

		matchStage := bson.D{
			{"$match", bson.D{}},
		}

		// 解析 Condition 字符串为 BSON
		var condition bson.M
		if err := bson.UnmarshalExtJSON([]byte(query.Condition), true, &condition); err != nil {
			zap.L().Error("Invalid condition format", zap.Error(err))
			common.ErrorResponse(c, http.StatusBadRequest, "Invalid condition format")
			return
		}

		if len(condition) > 0 {
			matchStage = bson.D{
				{"$match", condition},
			}
		}
		zap.L().Debug("condition", zap.Any("condition", condition), zap.Any("match", matchStage))
		sortStage := bson.D{
			{"$sort", bson.D{{"last_seen", -1}}},
		}
		limitStage := bson.D{
			{"$limit", pagination.Limit},
		}
		skipStage := bson.D{
			{"$skip", skip},
		}
		pipeline := mongo.Pipeline{matchStage, sortStage, skipStage, limitStage}

		coll := mongodb.GetMongoClient().Database(types.MongoDatabaseUserAgent).Collection(collection)
		cursor, err := coll.Aggregate(context.Background(), pipeline)

		var result []types.UserAgentRecord
		if err != nil {
			zap.L().Error("mongodb.Aggregate", zap.Error(err))
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var log types.UserAgentRecord
			cursor.Decode(&log)
			result = append(result, log)
		}

		pagination.Result = result
		pagination.TotalCount, err = coll.CountDocuments(context.Background(), condition)
		if err != nil {
			zap.L().Error("mongodb.cursor", zap.Error(err))
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		common.SuccessResponse(c, pagination)
	}
}

// Application 应用识别
func Application() gin.HandlerFunc {
	return func(c *gin.Context) {
		var query QueryData
		if err := c.BindJSON(&query); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
			return
		}
		zap.L().Debug("query", zap.Any("params", query))

		pagination := utils.NewPagination(strconv.Itoa(query.Page), strconv.Itoa(query.PageSize))

		collection := query.Collection
		// skip
		skip := (pagination.Page - 1) * pagination.Limit

		matchStage := bson.D{
			{"$match", bson.D{}},
		}

		// 解析 Condition 字符串为 BSON
		var condition bson.M
		if err := bson.UnmarshalExtJSON([]byte(query.Condition), true, &condition); err != nil {
			zap.L().Error("Invalid condition format", zap.Error(err))
			common.ErrorResponse(c, http.StatusBadRequest, "Invalid condition format")
			return
		}

		if len(condition) > 0 {
			matchStage = bson.D{
				{"$match", condition},
			}
		}
		zap.L().Debug("condition", zap.Any("condition", condition), zap.Any("match", matchStage))
		//groupStage := bson.D{
		//	{"$group", bson.D{}},
		//}
		//filterStage := bson.D{
		//	//{"$and", []bson.D{}},
		//}
		//projectStage := bson.D{
		//	//{"$project", bson.D{}},
		//}
		sortStage := bson.D{
			{"$sort", bson.D{{"_id", -1}}},
		}
		limitStage := bson.D{
			{"$limit", pagination.Limit},
		}
		skipStage := bson.D{
			{"$skip", skip},
		}
		pipeline := mongo.Pipeline{matchStage, sortStage, skipStage, limitStage}

		coll := mongodb.GetMongoClient().Database(types.MongoDatabaseStream).Collection(collection)
		cursor, err := coll.Aggregate(context.Background(), pipeline)

		var result []types.Sessions
		if err != nil {
			zap.L().Error("mongodb.Aggregate", zap.Error(err))
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var log types.Sessions
			cursor.Decode(&log)
			result = append(result, log)
		}

		pagination.Result = result
		pagination.TotalCount, err = coll.CountDocuments(context.Background(), condition)
		if err != nil {
			zap.L().Error("mongodb.cursor", zap.Error(err))
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		common.SuccessResponse(c, pagination)
	}
}

// Detail 详情
func Detail() gin.HandlerFunc {
	return func(c *gin.Context) {
		var jsonData struct {
			IP string `json:"ip"`
		}
		_ = c.ShouldBind(&jsonData)
		bytes, err := socket.SendUnixMessage(socket.IPDetail, jsonData.IP)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		var res models.IPDetail
		_ = json.Unmarshal(bytes, &res)
		res.Features, err = getFeature(jsonData.IP)
		res.Devices, err = resolve.GetDevicesByIP(jsonData.IP)
		res.DevicesLogs, err = getDevicesLogs(jsonData.IP)
		common.SuccessResponse(c, res)
	}
}
