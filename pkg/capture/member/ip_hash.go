package member

import (
	"context"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	v9 "github.com/redis/go-redis/v9"
	"sync"
	"time"
)

// IP Hash

type Hash struct {
	IP    string
	Field types.Property
	Value any
}

// Store IP加载/更新
func Store(i interface{}) {
	hash := i.(Hash)

	var m *sync.Map
	var v any
	switch hash.Field {
	case types.TTL:
		m = &TTLCache
		v = hash.Value.(uint8)
		break
	case types.Mac:
		m = &MacCache
		v = hash.Value.(string)
		break
	case types.UserAgent:
		m = &UaCache
		v = hash.Value.(string)
		break
	case types.Device:
		m = &DeviceCache
		v = hash.Value.(string)
		break
	case types.DeviceName:
		m = &DeviceNameCache
		v = hash.Value.(string)
		break
	case types.DeviceType:
		m = &DeviceTypeCache
		v = hash.Value.(string)
		break
	}

	mutex := getIPMutex(hash.IP)
	mutex.RLock()
	oldVal, ok := getMemory(hash.IP, m)
	mutex.RUnlock()

	mutex.Lock()
	defer mutex.Unlock()
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
	// 推送至channel
	Events <- PropertyChangeEvent{
		IP:       hash.IP,
		OldValue: oldVal,
		NewValue: v,
		Property: hash.Field,
	}
	return
}

// 从缓存中获取
func getMemory(ip string, m *sync.Map) (any, bool) {
	val, ok := m.Load(ip)
	if ok {
		return val, ok
	}
	return nil, false
}

// 更新缓存
func putMemory(ip string, m *sync.Map, v any) {
	m.Store(ip, v)
}

func GetAnalyze(ip string) bool {
	_, ok := TTLAnalyze.Load(ip)
	return ok
}

func PutAnalyze(ip string) {
	TTLAnalyze.Store(ip, true)
}

// DelMemory 删除缓存
func DelMemory(ip string) {
	TTLAnalyze.Delete(ip)
	TTLCache.Delete(ip)
	UaCache.Delete(ip)
	DeviceCache.Delete(ip)
	DeviceNameCache.Delete(ip)
	DeviceTypeCache.Delete(ip)
	MacCache.Delete(ip)
}

// 从redis获取属性
func getPropertyForRedis(ip string, property types.Property) (any, bool) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := fmt.Sprintf(types.HashAnalyzeIP, ip)

	val, err := rdb.HMGet(ctx, key, string(property)).Result()
	if errors.Is(err, v9.Nil) || len(val) == 1 {
		return nil, false
	}
	return val[1], true
}

// GetHashForRedis 从redis获取hash
func GetHashForRedis(ip string) map[string]string {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := fmt.Sprintf(types.HashAnalyzeIP, ip)

	val, err := rdb.HGetAll(ctx, key).Result()
	if errors.Is(err, v9.Nil) || len(val) == 1 {
		return nil
	}
	return val
}

// 存储ip hash 至redis
func storeHash2Redis(ip string, property types.Property, value any) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := fmt.Sprintf(types.HashAnalyzeIP, ip)

	// z_set 有序集合
	rdb.ZAdd(ctx, types.ZSetIP, v9.Z{
		Score:  float64(time.Now().Unix()),
		Member: ip,
	}).Val()
	// info hash
	rdb.HSet(ctx, key, string(property), value).Val()
}

func CleanUp() {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	rdb.Del(ctx, types.ZSetIP).Val()
}
