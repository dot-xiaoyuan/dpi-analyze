package handler

import (
	"encoding/json"
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"go.uber.org/zap"
)

func ConfigUpdate(raw json.RawMessage) any {
	zap.L().Debug("config update", zap.Any("raw", raw))
	var params config.Yaml
	_ = json.Unmarshal(raw, &params)

	return mongodb.UpdateConfig(params)
}
