package capture

import (
	"context"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	redis2 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Observer struct {
	Table string
}

// TTLChangeObserverEvent TTL变化观察事件
type TTLChangeObserverEvent struct {
	IP   string
	Prev any
	Curr any
	Diff uint8
}

type TTLChange struct {
	Time time.Time `json:"time"`
	TTL  uint8     `json:"ttl"`
}

// TTLChangeHistory TTL变化历史记录
type TTLChangeHistory struct {
	Changes []TTLChange
}

var (
	TTLHistoryCache = make(map[string]*TTLChangeHistory)
	cacheMutex      sync.Mutex
)

// RecordTTLChange 记录ttl变化
func RecordTTLChange(ip string, ttl uint8) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	history, exists := TTLHistoryCache[ip]
	if !exists {
		storeToRedis(ip)

		history = &TTLChangeHistory{
			Changes: make([]TTLChange, 0, 30),
		}
		TTLHistoryCache[ip] = history
	}

	if len(history.Changes) == 30 {
		history.Changes = history.Changes[1:]
	}
	// 记录变化
	history.Changes = append(history.Changes, TTLChange{
		Time: time.Now(),
		TTL:  ttl,
	})
}

// GetTTLHistory 获取ttl历史记录
func GetTTLHistory(ip string) []TTLChange {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if history, exists := TTLHistoryCache[ip]; exists {
		return history.Changes
	}
	return nil
}

// WatchTTLChange 观察ttl变化
func WatchTTLChange(events <-chan TTLChangeObserverEvent) {
	for e := range events {
		RecordTTLChange(e.IP, e.Curr.(uint8))
	}
}

func CleanUp() {
	// TODO 清空缓存
	redis.GetRedisClient().Del(context.TODO(), ZSetObserverIPTable)
}

func storeToRedis(ip string) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	rdb.ZAdd(ctx, ZSetObserverIPTable, redis2.Z{
		Score:  float64(time.Now().Unix()),
		Member: ip,
	})
}

type ObserverResults struct {
	TotalCount int64 `json:"totalCount"`
	Results    []any `json:"results"`
}

type TTLDetail struct {
	IP      string      `json:"ip"`
	History []TTLChange `json:"history"`
}

func (o *Observer) Traversal(c Condition) (any, error) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	var result ObserverResults
	var err error
	// 分页的起止索引
	start := (c.Page - 1) * c.PageSize

	// Pipeline 批量查询
	pipe := rdb.Pipeline()
	result.TotalCount = rdb.ZCount(ctx, c.Table, c.Min, c.Max).Val()
	// step1. 分页查询集合
	zRangCmd := rdb.ZRevRangeByScoreWithScores(ctx, c.Table, &redis2.ZRangeBy{
		Min:    c.Min,      // 查询范围的最小时间戳
		Max:    c.Max,      // 查询范围的最大时间戳
		Offset: start,      // 分页起始位置
		Count:  c.PageSize, // 每页大小
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		zap.L().Error("ZRange pipe.Exec", zap.Error(err))
		return nil, err
	}

	ips := zRangCmd.Val()
	for _, ip := range ips {
		var detail TTLDetail
		detail.IP = ip.Member.(string)
		detail.History = GetTTLHistory(ip.Member.(string))
		result.Results = append(result.Results, detail)
	}

	return result, nil
}
