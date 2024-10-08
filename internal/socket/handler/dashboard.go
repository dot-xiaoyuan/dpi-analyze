package handler

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze/memory"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
	"sync/atomic"
)

type Charts struct {
	Name  string `json:"type"`
	Value int64  `json:"value"`
}

func Dashboard(params interface{}) []byte {
	tcpCount := atomic.LoadInt64(&capture.TCPCount)
	udpCount := atomic.LoadInt64(&capture.TCPCount)

	transportCharts := []Charts{
		{Name: "TCP", Value: tcpCount},
		{Name: "UDP", Value: udpCount},
	}

	// TODO 应用排行方法迁移
	data := map[string]interface{}{
		"totalPackets":      capture.PacketsCount,
		"totalTraffics":     capture.TrafficCount,
		"totalSessions":     capture.SessionCount,
		"trafficCharts":     memory.GenerateChartData(),
		"appCharts":         types.GenerateChartData(),
		"transportCharts":   transportCharts,
		"applicationCharts": protocols.GenerateChartData(),
	}

	result, _ := json.Marshal(data)
	return result
}
