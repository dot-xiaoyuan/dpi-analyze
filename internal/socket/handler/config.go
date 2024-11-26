package handler

import (
	"encoding/json"
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"go.uber.org/zap"
)

func ConfigUpdate(raw json.RawMessage) any {
	zap.L().Debug("config update", zap.Any("raw", raw))

	var updates map[string]interface{}
	err := json.Unmarshal(raw, &updates)
	if err != nil {
		zap.L().Error("config update error", zap.Error(err))
		return err
	}

	err = mongodb.UpdateNestedConfig(config.Cfg, updates)
	if err != nil {
		zap.L().Error("config update error", zap.Error(err))
		return err
	}

	zap.L().Info("config update done", zap.Any("config", config.Cfg))
	err = mongodb.Store2Mongo()
	if err != nil {
		zap.L().Error("config update error", zap.Error(err))
		return err
	}
	return "successful"
}

func ConfigList(raw json.RawMessage) any {
	return config.Cfg
}
