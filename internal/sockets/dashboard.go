package sockets

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze/memory"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
)

type ActionDashboard struct{}

func (ActionDashboard) Handle(data json.RawMessage) []byte {
	//TODO implement me
	res := Res{
		Code: 200,
		Data: map[string]any{
			"totalPackets":  capture.PacketsCount,
			"totalTraffics": capture.TrafficCount,
			"totalSessions": capture.SessionCount,
			"trafficCharts": memory.GenerateChartData(),
		},
	}
	result, _ := json.Marshal(res)
	return result
}
