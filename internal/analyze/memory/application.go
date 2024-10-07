package memory

import (
	"context"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/layers"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	"go.uber.org/zap"
)

// 应用层

type Application struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

func GenerateList() []Application {
	rdb := redis.GetRedisClient()
	result := rdb.ZRevRangeWithScores(context.Background(), layers.ZSetApplication, 0, -1).Val()
	zap.L().Info("result", zap.Any("result", result))
	var charts []Application
	for _, v := range result {
		charts = append(charts, Application{
			Name:  v.Member.(string),
			Value: v.Score,
		})
	}
	return charts
}
