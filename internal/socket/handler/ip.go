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

	res := models.IPDetail{
		Detail: member.GetHashForRedis(ip),
		History: models.History{
			TTL: observer.TTLObserver.GetHistory(ip),
			Mac: observer.MacObserver.GetHistory(ip),
			Ua:  observer.UaObserver.GetHistory(ip),
		},
	}

	return res
}
