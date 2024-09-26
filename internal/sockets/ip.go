package sockets

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
)

type ActionIP struct{}

func (ActionIP) Handle(data json.RawMessage) []byte {
	c := struct {
		IP string `json:"ip"`
	}{}
	_ = json.Unmarshal(data, &c)
	res := Res{
		Code: 200,
		Data: map[string]any{
			"ttlHistory": capture.GetTTLHistory(c.IP),
			"detail":     capture.GetIPInfoFromRedis(c.IP),
		},
	}
	result, _ := json.Marshal(res)
	return result
}
