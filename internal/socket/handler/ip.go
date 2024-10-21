package handler

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/observer"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket/models"
	"go.uber.org/zap"
)

func IPDetail(raw json.RawMessage) any {
	zap.L().Debug("IPDetail", zap.Any("raw", raw))

	var ip string
	_ = json.Unmarshal(raw, &ip)

	featureSni, _ := member.GetFeatureByMemory[string](ip, &member.SNICache)

	res := models.IPDetail{
		TTLHistory: observer.TTLObserver.GetHistory(ip),
		MacHistory: observer.MacObserver.GetHistory(ip),
		UaHistory:  observer.UaObserver.GetHistory(ip),
		SNIHistory: featureSni,
		Detail:     member.GetHashForRedis(ip),
	}

	return res
}
