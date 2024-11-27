package handler

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze/memory"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket/models"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/statictics"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/users"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
)

type Charts struct {
	Name  string `json:"type"`
	Value int64  `json:"value"`
}

// Dashboard 仪表盘
// 总流量、总包数、总会话数、在线数
// 流量趋势图表、应用排行图表、应用分类图表
func Dashboard(raw json.RawMessage) any {
	res := models.Dashboard{
		Total: models.Total{
			Packets:  capture.PacketsCount,
			Traffics: utils.FormatBytes(capture.TrafficCount),
			Sessions: capture.SessionCount,
			Users:    users.GetTotalCount(),
		},
		Charts: models.Charts{
			Traffic:        memory.GenerateChartData(),
			TransportLayer: statictics.TransportLayer.GetStats(),
			Application:    statictics.Application.GetStats(),
			AppCategory:    statictics.AppCategory.GetStats(),
			Protocol:       statictics.ApplicationLayer.GetStats(),
		},
	}
	return res
}
