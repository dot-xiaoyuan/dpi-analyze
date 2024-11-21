package member

import (
	"context"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"time"
)

// 疑似代理

func TriggerSuspected(ip string, ft types.FeatureType, count int) {
	pf := getThreshold(ft)
	if pf.Threshold == 0 {
		return
	}
	if count > pf.Threshold {
		// 超过阈值，记录疑似代理
		record := types.SuspectedRecord{
			IP: ip,
			//Username:       username,
			ReasonCategory: "protocol_threshold",
			ReasonDetail: types.ReasonDetail{
				Name:      ft,
				Value:     count,
				Threshold: pf.Threshold,
			},
			Tags:     []string{pf.Normal},
			Context:  types.Context{},
			Remark:   pf.Remark,
			LastSeen: time.Now(),
		}
		_, _ = mongo.GetMongoClient().Database(types.MongoDatabaseSuspected).
			Collection(time.Now().Format("06_01")).
			InsertOne(context.TODO(), record)
	}
}

// 获取协议的阈值
func getThreshold(ft types.FeatureType) config.ProtocolFeature {
	switch ft {
	case types.SNI:
		return config.Cfg.Thresholds.SNI
	case types.HTTP:
		return config.Cfg.Thresholds.HTTP
	case types.TLSVersion:
		return config.Cfg.Thresholds.TLSVersion
	case types.CipherSuite:
		return config.Cfg.Thresholds.CipherSuite
	case types.Session:
		return config.Cfg.Thresholds.Session
	case types.DNS:
		return config.Cfg.Thresholds.DNS
	case types.QUIC:
		return config.Cfg.Thresholds.QUIC
	case types.SNMP:
		return config.Cfg.Thresholds.SNMP
	default:
		return config.ProtocolFeature{}
	}
}
