package capture

import (
	"context"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	redis2 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"sync"
	"time"
)

var (
	TTLCache sync.Map
	MacCache sync.Map
	UACache  sync.Map
	ipLocks  sync.Map
)

type EventType int

type IPInfoType string

const (
	TTL       string = "ttl"
	UserAgent string = "ua"
	Mac       string = "mac"
)

type IPFieldChangeEvent struct {
	IP       string
	OldValue any
	NewValue any
	Field    string
}

type IPInfo struct {
	TTL string
	UA  string
	Mac string
}

func getIPMutex(ip string) *sync.Mutex {
	val, _ := ipLocks.LoadOrStore(ip, &sync.Mutex{})
	return val.(*sync.Mutex)
}

// StoreIP 加载IP TTL
func StoreIP(ip, field string, val any) {
	// zap.L().Debug("store ip", zap.String("ip", ip), zap.Any("val", val))
	var m *sync.Map
	switch field {
	case TTL:
		m = &TTLCache
		val = val.(uint8)
		break
	case UserAgent:
		m = &UACache
		break
	case Mac:
		m = &MacCache
		break
	}

	// 获取IP锁
	mutex := getIPMutex(ip)
	mutex.Lock()
	defer mutex.Unlock()

	oldVal, ok := GetIPInfoFromMemory(ip, m)
	if !ok {
		// memory 缓存不存在，添加至缓存
		UpdateIPInfoFromMemory(ip, m, val)
		// 查询 redis
		oldVal, ok = GetIPInfoFromRedis(ip, field)
		if !ok {
			// redis 也不存在
			StoreIPInfoInRedis(ip, field, val)
			return
		}
	}
	// zap.L().Debug("old", zap.String("ip", ip), zap.Any("oldVal", oldVal))
	if oldVal == val {
		return
	}
	UpdateIPInfoFromMemory(ip, m, val)
	// 不一致，进行更新
	IPEvents <- IPFieldChangeEvent{
		IP:       ip,
		OldValue: oldVal,
		NewValue: val,
		Field:    field,
	}
}

// GetIPInfoFromMemory 获取IP属性 memory
func GetIPInfoFromMemory(ip string, memory *sync.Map) (any, bool) {
	val, ok := memory.Load(ip)
	if ok {
		return val, ok
	}
	return "", false
}

// UpdateIPInfoFromMemory 更新IP属性 memory
func UpdateIPInfoFromMemory(ip string, memory *sync.Map, val any) {
	memory.Store(ip, val)
}

// GetIPInfoFromRedis 获取IP属性 redis
func GetIPInfoFromRedis(ip string, t string) (any, bool) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := fmt.Sprintf(HashAnalyzeIP, ip)

	val, err := rdb.HMGet(ctx, key, t).Result()
	if errors.Is(err, redis2.Nil) || len(val) == 1 {
		return nil, false
	}
	return val[1], true
}

// StoreIPInfoInRedis 存储IP属性 redis
func StoreIPInfoInRedis(ip, field string, value any) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := fmt.Sprintf(HashAnalyzeIP, ip)

	// z_set 有序集合
	rdb.ZAdd(ctx, ZSetIPTable, redis2.Z{
		Score:  float64(time.Now().Unix()),
		Member: ip,
	})
	// info hash
	rdb.HSet(ctx, key, field, value).Val()
	rdb.Expire(ctx, key, time.Hour)
}

// ProcessChangeEvent IP属性监听事件
func ProcessChangeEvent(events <-chan IPFieldChangeEvent) {
	for e := range events {
		//zap.L().Debug("ProcessChangeEvent", zap.Any("event", e))
		mutex := getIPMutex(e.IP)
		mutex.Lock()

		switch e.Field {
		case TTL:
			// 处理TTL变化
			zap.L().Debug(i18n.T("TTLChange"), zap.String("ip", e.IP), zap.Any("old", e.OldValue), zap.Any("new", e.NewValue))
			// push observer
			ObserverEvents <- TTLChangeObserverEvent{
				IP:   e.IP,
				Prev: e.OldValue,
				Curr: e.NewValue,
			}
			break
		case Mac:
			// 处理Mac地址变化
			zap.L().Debug(i18n.T("MacChange"), zap.String("ip", e.IP), zap.Any("old", e.OldValue), zap.Any("new", e.NewValue))
			UpdateIPInfoFromMemory(e.IP, &MacCache, e.NewValue)
			break
		case UserAgent:
			// 处理UA变化
			zap.L().Debug(i18n.T("UAChange"), zap.String("ip", e.IP), zap.Any("old", e.OldValue), zap.Any("new", e.NewValue))
			UpdateIPInfoFromMemory(e.IP, &UACache, e.NewValue)
			break
		}
		// 更新至redis
		StoreIPInfoInRedis(e.IP, e.Field, e.NewValue)
		// 解锁
		mutex.Unlock()
	}
}
