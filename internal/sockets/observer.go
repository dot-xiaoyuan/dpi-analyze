package sockets

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/observer"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/provider"
	"go.uber.org/zap"
)

type ActionObserver struct {
}

func (a *ActionObserver) Handle(data json.RawMessage) []byte {
	var condition provider.Condition
	err := json.Unmarshal(data, &condition)
	if err != nil {
		zap.L().Error("condition json unmarshal failed", zap.Error(err))
		return nil
	}
	t := &observer.Observer[string]{
		Table: condition.Table,
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
