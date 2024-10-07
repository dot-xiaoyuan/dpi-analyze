package sockets

import (
	"context"
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze/memory"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
	"go.uber.org/zap"
	"sync/atomic"
)

type ActionDashboard struct{}

type Chart struct {
	Name  string `json:"type"`
	Value int64  `json:"value"`
}

func (ActionDashboard) Handle(data json.RawMessage) []byte {
	tcpCount := atomic.LoadInt64(&capture.TCPCount)
	udpCount := atomic.LoadInt64(&capture.UDPCount)

	transportCharts := []Chart{
		{Name: "TCP", Value: tcpCount},
		{Name: "UDP", Value: udpCount},
	}
	res := Res{
		Code: 200,
		Data: map[string]any{
			"totalPackets":      capture.PacketsCount,
			"totalTraffics":     capture.TrafficCount,
			"totalSessions":     capture.SessionCount,
			"trafficCharts":     memory.GenerateChartData(),
			"appCharts":         GenerateList(),
			"transportCharts":   transportCharts,
			"applicationCharts": protocols.GenerateChartData(),
		},
	}
	result, _ := json.Marshal(res)
	return result
}

type Application struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

func GenerateList() []Application {
	rdb := redis.GetRedisClient()
	result := rdb.ZRevRangeWithScores(context.Background(), types.ZSetApplication, 0, -1).Val()
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
