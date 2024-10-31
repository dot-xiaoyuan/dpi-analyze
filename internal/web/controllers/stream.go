package controllers

import (
	"context"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type QueryData struct {
	Collection string `json:"collection"`
	Condition  string `json:"condition"`
	HttpInfo   bool   `json:"http_info"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	SortField  string `json:"sortField"`
	SortOrder  string `json:"sortOrder"`
}

func StreamList() gin.HandlerFunc {
	return func(c *gin.Context) {
		var query QueryData
		if err := c.BindJSON(&query); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
			return
		}
		zap.L().Debug("query", zap.Any("params", query))

		pagination := utils.NewPagination(strconv.Itoa(query.Page), strconv.Itoa(query.PageSize))

		collection := query.Collection
		sortField := query.SortField
		sortOrder := query.SortOrder
		var orderBy int
		if sortOrder == "descend" {
			orderBy = -1
		} else {
			orderBy = 1
		}
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
			{"$sort", bson.D{{sortField, orderBy}}},
		}
		limitStage := bson.D{
			{"$limit", pagination.Limit},
		}
		skipStage := bson.D{
			{"$skip", skip},
		}
		pipeline := mongo.Pipeline{matchStage, sortStage, skipStage, limitStage}

		coll := mongodb.GetMongoClient().Database("dpi").Collection(collection)
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
