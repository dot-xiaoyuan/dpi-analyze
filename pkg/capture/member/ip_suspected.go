package member

import (
	"context"
	"errors"
	"fmt"
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
			Shards:           1024,
			LifeWindow:       10 * time.Minute,
			MaxEntrySize:     500,
			CleanWindow:      5 * time.Minute,
			HardMaxCacheSize: 128,
			Verbose:          true,
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
		return
	}
	// 检查缓存是否存在
	_, err := GetSuspectedCache().Get(ip)
	if err != nil {
		if !errors.Is(bigcache.ErrEntryNotFound, err) {
			// 如果是其他错误，记录日志
			return
		}
		// 缓存不存在，继续处理
	} else {
		// 如果缓存已存在，直接返回
		return
	}
	if count > pf.Threshold {
		// 超过阈值，记录疑似代理
		record := types.SuspectedRecord{
			IP: ip,
			//Username:       username,
			ReasonCategory: "protocol_threshold",
			ReasonDetail: types.ReasonDetail{
				Name:        ft,
				Value:       count,
				Threshold:   pf.Threshold,
				Description: fmt.Sprintf("短时间内%s次数超过限定阈值:%d", ft, pf.Threshold),
			},
			Tags:     []string{pf.Normal},
			Context:  types.Context{},
			Remark:   pf.Remark,
			LastSeen: time.Now(),
		}
		_, err = mongo.GetMongoClient().Database(types.MongoDatabaseSuspected).
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
