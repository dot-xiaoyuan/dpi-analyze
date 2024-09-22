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
			"packets":  capture.PacketsCount,
			"flows":    capture.FlowCount,
			"sessions": capture.SessionCount,
			"traffic":  memory.GenerateChartData(),
		},
	}
	result, _ := json.Marshal(res)
	return result
}
