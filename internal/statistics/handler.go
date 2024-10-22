package statistics

import (
	"context"
	mongo2 "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"net/http"
)

func Index(c *gin.Context) {
	collection, _ := c.GetQuery("collection")
	zap.L().Info("collection: " + collection)
	matchStage := bson.D{
		{"$match", bson.D{
			{"metadata.http_info.urls", bson.D{
				{"$ne", nil},
			}},
			{"metadata.http_info.upgrade", "mmtls"},
		}},
	}
	groupStage := bson.D{
		{"$group", bson.D{
			{"_id", bson.D{
				{"url", "$metadata.http_info.urls"},
				//{"host", "$metadata.http_info.host"},
			}},
			{"dst_ips", bson.D{{"$addToSet", "$dst_ip"}}},
			{"count", bson.D{{"$sum", 1}}},
		}},
	}
	filterStage := bson.D{
		{"$match", bson.D{
			{"dst_ips.1", bson.D{{"$exists", true}}},
		}},
	}
	projectStage := bson.D{
		{"$project", bson.D{
			{"url", "$_id.url"},
			//{"host", "$_id.host"},
			{"dst_ips", 1},
			{"count", 1},
		}},
	}
	pipeline := mongo.Pipeline{matchStage, groupStage, filterStage, projectStage}

	cursor, err := mongo2.GetMongoClient().Database("dpi").
		Collection(collection).
		Aggregate(context.Background(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}
	//for _, result := range results {
	c.JSON(http.StatusOK, gin.H{
		"data": results,
	})
	//}
}
