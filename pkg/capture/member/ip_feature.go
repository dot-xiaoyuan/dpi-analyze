package member

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/types"
	"sync"
)

// 特征

var (
	SNICache  sync.Map
	HTTPCache sync.Map
)

type Feature[T any] struct {
	IP    string
	Field types.Feature
	Value T
}

// IP锁
func getFeatureMutex(ip string) *sync.Mutex {
	val, _ := Mutex.LoadOrStore(ip, &sync.Mutex{})
	return val.(*sync.Mutex)
}

func Increment[T string | int](i interface{}) {
	hash := i.(Feature[T])

	var m *sync.Map
	switch hash.Field {
	case types.SNI:
		m = &SNICache
		break
	case types.HTTP:
		m = &HTTPCache
		break
	}

	mutex := getFeatureMutex(hash.IP)
	mutex.Lock()
	defer mutex.Unlock()

	features, ok := GetFeatureByMemory[T](hash.IP, m)
	if !ok {
		// memory 不存在，缓存
		putFeatureByMemory(hash.IP, m, []T{hash.Value})
		return
	}
	features = append(features, hash.Value)
	// 已存在特征则跳过
	for _, f := range features {
		if f == hash.Value {
			return
		}
	}
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

func putRedis(ip string, field types.Feature) {
	key := fmt.Sprintf(types.HashAnalyzeIP, ip)
	fmt.Printf("redis key: %s field: %s \n", key, field)
	redis.GetRedisClient().HIncrBy(context.Background(), key, string(field), 1).Val()
}
