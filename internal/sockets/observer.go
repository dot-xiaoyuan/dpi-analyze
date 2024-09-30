package sockets

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"go.uber.org/zap"
)

type ActionObserver struct {
}

func (a *ActionObserver) Handle(data json.RawMessage) []byte {
	var condition capture.Condition
	err := json.Unmarshal(data, &condition)
	if err != nil {
		zap.L().Error("condition json unmarshal failed", zap.Error(err))
		return nil
	}
	t := &capture.Observer{
		Table: "ttl",
	}
	res, err := t.Traversal(condition)
	if err != nil {
		zap.L().Error("condition traversal failed", zap.Error(err))
		return nil
	}
	jsonData, err := json.Marshal(res)
	if err != nil {
		return nil
	}
	return jsonData
}
