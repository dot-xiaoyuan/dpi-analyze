package observer

import (
	"context"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/provider"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	v9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"math"
	"sync"
	"time"
)

const (
	MaxTTLCount = 30
	MaxMacCount = 3
	MaxUaCount  = 5
)

var (
	cacheMutex sync.Mutex

	TTLObserver = NewObserver[uint8](types.ZSetObserverTTL, MaxTTLCount)
	MacObserver = NewObserver[string](types.ZSetObserverMac, MaxMacCount)
	UaObserver  = NewObserver[string](types.ZSetObserverUa, MaxUaCount)

	TTLEvents = make(chan ChangeObserverEvent[uint8], 100)
	MacEvents = make(chan ChangeObserverEvent[string], 100)
	UaEvents  = make(chan ChangeObserverEvent[string], 100)
)

// ChangeObserverEvent 观察事件
type ChangeObserverEvent[T string | uint8] struct {
	IP   string
	Prev T
	Curr T
}

// ChangeHistory 变化历史记录
type ChangeHistory[T string | uint8] struct {
	Changes       []ChangeRecord[T]
	ValueChanges  []uint8
	MovingAverage []float64
}

// ChangeRecord 记录每次变化的数据
type ChangeRecord[T string | uint8] struct {
	Time  time.Time `json:"time"`
	Value T         `json:"value"`
}

// Observer 观察者
type Observer[T string | uint8] struct {
	HistoryCache map[string]*ChangeHistory[T]
	MaxCount     int
	Table        string
}

// NewObserver 创建一个观察者
func NewObserver[T uint8 | string](table string, maxCount int) *Observer[T] {
	return &Observer[T]{
		HistoryCache: make(map[string]*ChangeHistory[T]),
		MaxCount:     maxCount,
		Table:        table,
	}
}

// RecordChange 记录变化
func (ob *Observer[T]) RecordChange(ip string, value T) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	history, ok := ob.HistoryCache[ip]
	if !ok {
		ob.Store2Redis(ip)
		history = &ChangeHistory[T]{
			Changes:      make([]ChangeRecord[T], 0, ob.MaxCount),
			ValueChanges: make([]uint8, 0, ob.MaxCount),
		}
		ob.HistoryCache[ip] = history
	}

	if len(history.Changes) == ob.MaxCount {
		history.Changes = history.Changes[1:]
	}

	history.Changes = append(history.Changes, ChangeRecord[T]{
		Time:  time.Now(),
		Value: value,
	})

	if ob.Table == types.ZSetObserverTTL {
		if v, ok := any(value).(uint8); ok {
			history.ValueChanges = append(history.ValueChanges, v)
			history.MovingAverage = MovingAverage(history.ValueChanges, 3)
		}
	}
}

// MovingAverage 泛型约束 T 只允许是数值类型
func MovingAverage(num []uint8, windowSize int) []float64 {
	var result []float64
	var sum uint8
	for i := 0; i < len(num); i++ {
		sum += num[i] // 累加数值
		if i >= windowSize {
			sum -= num[i-windowSize] // 从总和中减去滑出窗口的值
		}
		if i >= windowSize-1 {
			result = append(result, math.Round(float64(sum)/float64(windowSize))) // 计算并添加移动平均
		}
	}
	return result
}

// GetHistory 获取变化历史记录
func (ob *Observer[T]) GetHistory(ip string) []ChangeRecord[T] {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if history, ok := ob.HistoryCache[ip]; ok {
		return history.Changes
	}
	return nil
}

func (ob *Observer[T]) GetMovingAverage(ip string) []float64 {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if history, ok := ob.HistoryCache[ip]; ok {
		return history.MovingAverage
	}
	return nil
}

// Store2Redis 保存IP到Redis
func (ob *Observer[T]) Store2Redis(ip string) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	rdb.ZAdd(ctx, ob.Table, v9.Z{
		Score:  float64(time.Now().Unix()),
		Member: ip,
	})
}

// WatchChange 观察事件channel
func (ob *Observer[T]) WatchChange(events <-chan ChangeObserverEvent[T]) {
	for e := range events {
		ob.RecordChange(e.IP, e.Curr)
	}
}

// Traversal 遍历
func (ob *Observer[T]) Traversal(c provider.Condition) (int64, interface{}, error) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	// 分页的起止索引
	start := (c.Page - 1) * c.PageSize

	// Pipeline 批量查询
	pipe := rdb.Pipeline()
	count := rdb.ZCount(ctx, ob.Table, c.Min, c.Max).Val()
	// step1. 分页查询集合
	zRangCmd := rdb.ZRevRangeByScoreWithScores(ctx, ob.Table, &v9.ZRangeBy{
		Min:    c.Min,      // 查询范围的最小时间戳
		Max:    c.Max,      // 查询范围的最大时间戳
		Offset: start,      // 分页起始位置
		Count:  c.PageSize, // 每页大小
	})

	_, err := pipe.Exec(ctx)
	if err != nil {
		zap.L().Error("ZRange pipe.Exec", zap.Error(err))
		return 0, nil, err
	}

	ips := zRangCmd.Val()
	var result []interface{}
	for _, ip := range ips {
		var detail struct {
			IP            string            `json:"ip"`
			History       []ChangeRecord[T] `json:"history"`
			MovingAverage []float64         `json:"movingAverage"`
		}
		detail.IP = ip.Member.(string)
		detail.History = ob.GetHistory(ip.Member.(string))
		detail.MovingAverage = ob.GetMovingAverage(ip.Member.(string))
		result = append(result, detail)
	}
	return count, result, nil
}

func Setup() {
	_ = ants.Submit(func() {
		TTLObserver.WatchChange(TTLEvents)
	})
	_ = ants.Submit(func() {
		MacObserver.WatchChange(MacEvents)
	})
	_ = ants.Submit(func() {
		UaObserver.WatchChange(UaEvents)
	})
}

func CleanUp() {
	redis.GetRedisClient().Del(context.TODO(), types.ZSetObserverTTL)
	redis.GetRedisClient().Del(context.TODO(), types.ZSetObserverMac)
	redis.GetRedisClient().Del(context.TODO(), types.ZSetObserverUa)
}
