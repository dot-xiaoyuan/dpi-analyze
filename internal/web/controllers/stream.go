package controllers

import (
	"context"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/types"
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func StreamList() gin.HandlerFunc {
	return func(c *gin.Context) {
		pagination := utils.NewPagination(c.Query("page"), c.Query("page_size"))

		collection := c.DefaultQuery("collection", time.Now().Format("stream-06-01-02-15"))
		sortField := c.DefaultQuery("sortField", "_id")
		sortOrder := c.DefaultQuery("sortOrder", "descend")
		var orderBy int
		if sortOrder == "descend" {
			orderBy = -1
		} else {
			orderBy = 1
		}
		// skip
		skip := (pagination.Page - 1) * pagination.Limit
		zap.L().Info("query",
			zap.String("collection", collection),
			zap.Int64("page", pagination.Page),
			zap.Int64("limit", pagination.Limit),
			zap.Int64("skip", skip),
			zap.String("sortField", sortField),
			zap.String("sortOrder", sortOrder),
		)
		matchStage := bson.D{
			{"$match", bson.D{}},
		}
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
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var log types.Sessions
			cursor.Decode(&log)
			result = append(result, log)
		}

		pagination.Result = result
		pagination.TotalCount, err = coll.CountDocuments(context.Background(), bson.D{})
		if err != nil {
			zap.L().Error("mongodb.cursor", zap.Error(err))
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
		}
		common.SuccessResponse(c, pagination)
	}
}
