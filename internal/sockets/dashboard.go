package sockets

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze/memory"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
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
			"appCharts":         memory.GenerateList(),
			"transportCharts":   transportCharts,
			"applicationCharts": protocols.GenerateChartData(),
		},
	}
	result, _ := json.Marshal(res)
	return result
}
