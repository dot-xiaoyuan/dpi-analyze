package handler

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/observer"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket/models"
	"go.uber.org/zap"
)

func IPDetail(raw json.RawMessage) any {
	zap.L().Debug("IPDetail", zap.Any("raw", raw))

	var ip string
	_ = json.Unmarshal(raw, &ip)

	featureSni, _ := member.GetFeatureByMemory(ip, member.GetCache(types.SNI))
	featureHttp, _ := member.GetFeatureByMemory(ip, member.GetCache(types.HTTP))
	featureTLSVersion, _ := member.GetFeatureByMemory(ip, member.GetCache(types.TLSVersion))
	featureCipherSuite, _ := member.GetFeatureByMemory(ip, member.GetCache(types.CipherSuite))
	featureSession, _ := member.GetFeatureByMemory(ip, member.GetCache(types.Session))

	res := models.IPDetail{
		QPS:    member.GetQPS(ip),
		Detail: member.GetHashForRedis(ip),
		History: models.History{
			TTL: observer.TTLObserver.GetHistory(ip),
			Mac: observer.MacObserver.GetHistory(ip),
			Ua:  observer.UaObserver.GetHistory(ip),
		},
		Features: models.Features{
			SNI:         featureSni,
			HTTP:        featureHttp,
			TLSVersion:  featureTLSVersion,
			CipherSuite: featureCipherSuite,
			Session:     featureSession,
		},
	}

	return res
}
