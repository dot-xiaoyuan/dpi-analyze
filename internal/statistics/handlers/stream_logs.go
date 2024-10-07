package handlers

import (
	"context"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/layers"
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/db/mongo"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type results struct {
	Logs       []layers.Sessions `json:"logs"`
	TotalCount int64             `json:"totalCount"`
	Err        error             `json:"err"`
}

func StreamLogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		collection := c.DefaultQuery("collection", time.Now().Format("stream-06-01-02-15"))
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
		sortField := c.DefaultQuery("sortField", "_id")
		sortOrder := c.DefaultQuery("sortOrder", "descend")
		var orderBy int
		if sortOrder == "descend" {
			orderBy = -1
		} else {
			orderBy = 1
		}
		// skip
		skip := (page - 1) * pageSize
		zap.L().Info("query",
			zap.String("collection", collection),
			zap.Int("page", page),
			zap.Int("pageSize", pageSize),
			zap.Int("skip", skip),
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
			{"$limit", pageSize},
		}
		skipStage := bson.D{
			{"$skip", skip},
		}
		pipeline := mongo.Pipeline{matchStage, sortStage, skipStage, limitStage}

		coll := mongodb.GetMongoClient().Database("dpi").Collection(collection)
		cursor, err := coll.Aggregate(context.Background(), pipeline)

		var result results
		if err != nil {
			zap.L().Error("mongodb.Aggregate", zap.Error(err))
			result.Err = err
			c.JSON(http.StatusInternalServerError, result)
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var log layers.Sessions
			cursor.Decode(&log)
			result.Logs = append(result.Logs, log)
		}

		result.TotalCount, err = coll.CountDocuments(context.Background(), bson.D{})
		if err != nil {
			zap.L().Error("mongodb.cursor", zap.Error(err))
			c.JSON(http.StatusInternalServerError, result)
		}
		c.JSON(http.StatusOK, gin.H{
			"results": result,
		})
	}
}
