package capture

import (
	"context"
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

// StoreIP 加载IP TTL
func StoreIP(ip, field string, val any) {
	var m *sync.Map
	switch field {
	case TTL:
		m = &TTLCache
		break
	case UserAgent:
		m = &UACache
		break
	case Mac:
		m = &MacCache
		break
	}

	oldVal, ok := GetIPInfoFromMemory(ip, m)
	if !ok {
		oldVal = GetIPInfoFromRedis(ip, field)
	}
	if oldVal == "" || oldVal == nil {
		// memory 和 redis都不存在，进行缓存
		UpdateIPInfoFromMemory(ip, m, val)
		StoreIPInfoInRedis(ip, field, val)
		return
	}
	if oldVal == val {
		return
	}
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
func GetIPInfoFromRedis(ip string, t string) any {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := fmt.Sprintf(HashAnalyzeIP, ip)

	return rdb.HMGet(ctx, key, t).Val()[0]
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
		switch e.Field {
		case TTL:
			// 处理TTL变化
			zap.L().Debug(i18n.T("TTLChange"), zap.String("ip", e.IP), zap.Any("old", e.OldValue), zap.Any("new", e.NewValue))
			UpdateIPInfoFromMemory(e.IP, &TTLCache, e.NewValue)
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
	}
}
