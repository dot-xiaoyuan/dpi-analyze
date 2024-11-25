package handler

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze/memory"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket/models"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"sync/atomic"
)

type Charts struct {
	Name  string `json:"type"`
	Value int64  `json:"value"`
}

func Dashboard(raw json.RawMessage) any {
	tcpCount := atomic.LoadInt64(&capture.TCPCount)
	udpCount := atomic.LoadInt64(&capture.UDPCount)

	transportCharts := []Charts{
		{Name: "TCP", Value: tcpCount},
		{Name: "UDP", Value: udpCount},
	}

	// TODO 应用排行方法迁移
	res := models.Dashboard{
		Total: models.Total{
			Packets:  capture.PacketsCount,
			Traffics: utils.FormatBytes(capture.TrafficCount),
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
