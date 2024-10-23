package member

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	types2 "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"go.uber.org/zap"
	"sync"
)

// 特征

var (
	SNICache  sync.Map
	HTTPCache sync.Map
)

type Feature[T any] struct {
	IP    string
	Field types2.Feature
	Value T
}

func Increment[T string | int](i interface{}) {
	hash := i.(Feature[T])

	var m *sync.Map
	switch hash.Field {
	case types2.SNI:
		m = &SNICache
		break
	case types2.HTTP:
		m = &HTTPCache
		break
	}

	features, ok := GetFeatureByMemory[T](hash.IP, m)
	if !ok {
		// memory 不存在，缓存
		putFeatureByMemory(hash.IP, m, []T{hash.Value})
		return
	}
	// 已存在特征则跳过
	for _, f := range features {
		if f == hash.Value {
			return
		}
	}
	zap.L().Debug("append", zap.Any("sni", hash.Value))
	features = append(features, hash.Value)
	putFeatureByMemory(hash.IP, m, features)
	putRedis(hash.IP, hash.Field)
	return
}

func GetFeatureByMemory[T any](ip string, m *sync.Map) ([]T, bool) {
	val, ok := m.Load(ip)
	if ok {
		return val.([]T), true
	}
	return nil, false
}

func putFeatureByMemory[T any](ip string, m *sync.Map, values []T) {
	m.Store(ip, values)
}

func putRedis(ip string, field types2.Feature) {
	key := fmt.Sprintf(types2.HashAnalyzeIP, ip)
	zap.L().Debug("redis key", zap.String("key", key))
	redis.GetRedisClient().HIncrBy(context.Background(), key, string(field), 1).Val()
}
