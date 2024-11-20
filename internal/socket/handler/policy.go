package handler

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/policy"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"go.uber.org/zap"
)

// 策略

func PolicyList(raw json.RawMessage) any {
	zap.L().Debug("policy list", zap.Any("raw", raw))
	return policy.GetList()
}

func PolicyUpdate(raw json.RawMessage) any {
	zap.L().Debug("policy list", zap.Any("raw", raw))
	var params types.Products
	_ = json.Unmarshal(raw, &params)

	return policy.Policy.Update(params)
}
