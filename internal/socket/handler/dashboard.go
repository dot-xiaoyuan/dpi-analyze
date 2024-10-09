package handler

import (
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze/memory"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket/models"
	"sync/atomic"
)

type Charts struct {
	Name  string `json:"type"`
	Value int64  `json:"value"`
}

func Dashboard(params string) any {
	tcpCount := atomic.LoadInt64(&capture.TCPCount)
	udpCount := atomic.LoadInt64(&capture.TCPCount)

	transportCharts := []Charts{
		{Name: "TCP", Value: tcpCount},
		{Name: "UDP", Value: udpCount},
	}

	// TODO 应用排行方法迁移
	res := models.Dashboard{
		Total: models.Total{
			Packets:  capture.PacketsCount,
			Traffics: capture.TrafficCount,
			Sessions: capture.SessionCount,
		},
		Charts: models.Charts{
			ApplicationLayer: protocols.GenerateChartData(),
			TransportLayer:   transportCharts,
			Traffic:          memory.GenerateChartData(),
			Application:      types.GenerateChartData(),
		},
	}
	return res
}
