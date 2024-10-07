package observer

import (
	"context"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/provider"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	v9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"sync"
)

var (
	Events          = make(chan TTLChangeObserverEvent, 100)
	TTLHistoryCache = make(map[string]*TTLChangeHistory)
	cacheMutex      sync.Mutex
)

type Observer struct {
	Table string
}

func Setup() {
	_ = ants.Submit(func() {
		WatchTTLChange(Events)
	})
}

func CleanUp() {
	// TODO 清空缓存
	redis.GetRedisClient().Del(context.TODO(), types.ZSetObserverTTL)
}

type Results struct {
	TotalCount int64 `json:"totalCount"`
	Results    []any `json:"results"`
}

type TTLDetail struct {
	IP      string      `json:"ip"`
	History []TTLChange `json:"history"`
}

var handlers = map[string]func(z v9.Z) any{
	types.ZSetObserverTTL: func(z v9.Z) any {
		var detail TTLDetail
		detail.IP = z.Member.(string)
		detail.History = GetTTLHistory(z.Member.(string))
		return detail
	},
}

func (o *Observer) Traversal(c provider.Condition) (any, error) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	var result Results
	var err error
	// 分页的起止索引
	start := (c.Page - 1) * c.PageSize

	// Pipeline 批量查询
	pipe := rdb.Pipeline()
	result.TotalCount = rdb.ZCount(ctx, c.Table, c.Min, c.Max).Val()
	// step1. 分页查询集合
	zRangCmd := rdb.ZRevRangeByScoreWithScores(ctx, c.Table, &v9.ZRangeBy{
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
		if handle, ok := handlers[c.Table]; ok {
			r := handle(ip)
			result.Results = append(result.Results, r)
		}
	}

	return result, nil
}
