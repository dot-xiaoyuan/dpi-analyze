package member

import (
	"context"
	"github.com/allegro/bigcache"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"go.uber.org/zap"
	"sync"
	"time"
)

var (
	suspectedCache *bigcache.BigCache
	once           sync.Once
)

func GetSuspectedCache() *bigcache.BigCache {
	once.Do(func() {
		var err error
		suspectedCache, err = bigcache.NewBigCache(bigcache.Config{
			LifeWindow: 10 * time.Minute,
		})
		if err != nil {
			zap.L().Fatal("GetSuspectedCache", zap.Error(err))
		}
	})
	return suspectedCache
}

// 疑似代理

func TriggerSuspected(ip string, ft types.FeatureType, count int) {
	pf := getThreshold(ft)
	if pf.Threshold == 0 {
		//zap.L().Warn("threshold is empty", zap.Any("ft", ft))
		return
	}
	// 检查缓存是否存在
	if _, err := GetSuspectedCache().Get(ip); err == nil {
		// 如果 IP 已在缓存中，不再重复记录
		//zap.L().Debug("IP is already cached, skipping", zap.String("ip", ip))
		return
	}
	//zap.L().Debug("trigger suspected for ip", zap.String("ip", ip), zap.Int("count", count), zap.Int("threshold", pf.Threshold))
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
		_, err := mongo.GetMongoClient().Database(types.MongoDatabaseSuspected).
			Collection(time.Now().Format("06_01")).
			InsertOne(context.TODO(), record)
		if err != nil {
			zap.L().Error("failed to insert suspected record", zap.String("ip", ip), zap.Error(err))
			return
		}

		// 缓存
		err = GetSuspectedCache().Set(ip, []byte("cached"))
		if err != nil {
			zap.L().Error("failed to insert suspected record", zap.String("ip", ip), zap.Error(err))
			return
		}
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
