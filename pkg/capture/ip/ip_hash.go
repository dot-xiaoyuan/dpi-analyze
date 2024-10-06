package ip

import (
	"context"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/layers"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	redis2 "github.com/redis/go-redis/v9"
	"sync"
	"time"
)

// IP Hash

type Hash struct {
	IP    string
	Field Property
	Value any
}

func Store(i interface{}) {
	hash := i.(Hash)

	var m *sync.Map
	var v any
	switch hash.Field {
	case TTL:
		m = &TTLCache
		v = hash.Value.(uint8)
		break
	case Mac:
		m = &MacCache
		v = hash.Value.(string)
		break
	case UserAgent:
		m = &UaCache
		v = hash.Value.(string)
		break
	}

	mutex := getIPMutex(hash.IP)
	mutex.Lock()
	defer mutex.Unlock()

	oldVal, ok := getMemory(hash.IP, m)
	if !ok {
		// memory 不存在,进行缓存
		putMemory(hash.IP, m, v)
		// 查询 redis
		oldVal, ok = getPropertyForRedis(hash.IP, hash.Field)
		if !ok {
			// redis 也不存在
			storeHash2Redis(hash.IP, hash.Field, v)
			return
		}
	}
	if oldVal == v {
		return
	}
	// 数值不一致， 更新缓存并推送事件
	putMemory(hash.IP, m, v)

}

func getMemory(ip string, m *sync.Map) (any, bool) {
	val, ok := m.Load(ip)
	if ok {
		return val, ok
	}
	return nil, false
}

func putMemory(ip string, m *sync.Map, v any) {
	m.Store(ip, v)
}

func getPropertyForRedis(ip string, property Property) (any, bool) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := fmt.Sprintf(layers.HashAnalyzeIP, ip)

	val, err := rdb.HMGet(ctx, key, string(property)).Result()
	if errors.Is(err, redis2.Nil) || len(val) == 1 {
		return nil, false
	}
	return val[1], true
}

func GetHashForRedis(ip string) map[string]string {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := fmt.Sprintf(layers.HashAnalyzeIP, ip)

	val, err := rdb.HGetAll(ctx, key).Result()
	if errors.Is(err, redis2.Nil) || len(val) == 1 {
		return nil
	}
	return val
}

func storeHash2Redis(ip string, property Property, value any) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := fmt.Sprintf(layers.HashAnalyzeIP, ip)

	// z_set 有序集合
	rdb.ZAdd(ctx, layers.ZSetIPTable, redis2.Z{
		Score:  float64(time.Now().Unix()),
		Member: ip,
	})
	// info hash
	rdb.HSet(ctx, key, string(property), value).Val()
	rdb.Expire(ctx, key, time.Hour)
}
