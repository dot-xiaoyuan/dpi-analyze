package sockets

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
)

type ActionDashboard struct{}

func (ActionDashboard) Handle(data json.RawMessage) []byte {
	//TODO implement me
	res := Res{
		Code: 200,
		Data: map[string]int{
			"packets":  capture.PacketsCount,
			"flows":    capture.FlowCount,
			"sessions": capture.SessionCount,
		},
	}
	result, _ := json.Marshal(res)
	return result
}
