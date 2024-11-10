package observer

import (
	"context"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/provider"
	v9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"math"
	"sync"
	"time"
)

const (
	MaxTTLCount = 15
	MaxMacCount = 3
	MaxUaCount  = 5
)

var (
	cacheMutex sync.Mutex

	TTLObserver    = NewObserver[uint8](types.ZSetObserverTTL, MaxTTLCount)
	MacObserver    = NewObserver[string](types.ZSetObserverMac, MaxMacCount)
	UaObserver     = NewObserver[string](types.ZSetObserverUa, MaxUaCount)
	DeviceObserver = NewObserver[types.DeviceRecord](types.ZSetObserverDevice, MaxMacCount)

	TTLEvents    = make(chan ChangeObserverEvent[uint8], 100)
	MacEvents    = make(chan ChangeObserverEvent[string], 100)
	UaEvents     = make(chan ChangeObserverEvent[string], 100)
	DeviceEvents = make(chan ChangeObserverEvent[types.DeviceRecord], 100)
)

// ChangeObserverEvent 观察事件
type ChangeObserverEvent[T any] struct {
	IP   string
	Prev T
	Curr T
}

// ChangeHistory 变化历史记录
type ChangeHistory[T any] struct {
	Changes       []ChangeRecord[T] `json:"origin_changes"`
	ValueChanges  []uint            `json:"value_changes"`
	MovingAverage []float64         `json:"moving_average"`
	IsProxy       bool              `json:"is_proxy"`
}

// ChangeRecord 记录每次变化的数据
type ChangeRecord[T any] struct {
	Time  time.Time `json:"time"`
	Value T         `json:"value"`
}

// Observer 观察者
type Observer[T any] struct {
	HistoryCache map[string]*ChangeHistory[T]
	MaxCount     int
	Table        string
}

// NewObserver 创建一个观察者
func NewObserver[T any](table string, maxCount int) *Observer[T] {
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
			ValueChanges: make([]uint, 0, ob.MaxCount),
		}
		ob.HistoryCache[ip] = history
	}

	if len(history.Changes) == ob.MaxCount {
		history.Changes = history.Changes[1:]
	}
	if len(history.MovingAverage) == ob.MaxCount {
		history.MovingAverage = history.MovingAverage[1:]
	}
	if len(history.ValueChanges) == ob.MaxCount {
		history.ValueChanges = history.ValueChanges[1:]
	}

	history.Changes = append(history.Changes, ChangeRecord[T]{
		Time:  time.Now(),
		Value: value,
	})

	if ob.Table == types.ZSetObserverTTL {
		if v, ok := any(value).(uint8); ok {
			num := append(history.ValueChanges, uint(v))
			history.ValueChanges = num
			history.MovingAverage, history.IsProxy = detectProxyUsingSMA(num, 3, 3)
		}
	}
	if ob.Table == types.ZSetObserverDevice {
		_, _ = mongo.GetMongoClient().Database(string(types.Device)).Collection("record").InsertOne(mongo.Context, value)
	}
}

func detectProxyUsingSMA(num []uint, windowSize int, threshold float64) ([]float64, bool) {
	// 计算平滑处理后的TTL序列（SMA）
	sma := movingAverage(num, windowSize)

	// 比较原始TTL和平滑后的TTL差异
	for i := windowSize - 1; i < len(num); i++ {
		if math.Abs(float64(num[i])-sma[i-windowSize+1]) > threshold {
			return sma, true // 如果TTL的变化幅度超过阈值，认为有代理存在
		}
	}

	return sma, false // 没有检测到代理
}

// movingAverage 泛型约束 T 只允许是数值类型
func movingAverage(num []uint, windowSize int) []float64 {
	var result []float64
	var sum uint
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
func (ob *Observer[T]) GetHistory(ip string) *ChangeHistory[T] {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if history, ok := ob.HistoryCache[ip]; ok {
		return history
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
	}).Val()
}

func (ob *Observer[T]) DeleteRedis(ip string) {
	redis.GetRedisClient().ZRem(context.TODO(), ob.Table, ip).Val()
}

// WatchChange 观察事件channel
func (ob *Observer[T]) WatchChange(events <-chan ChangeObserverEvent[T]) {
	for e := range events {
		ob.RecordChange(e.IP, e.Curr)
	}
}

type WebResult[T any] struct {
	IP      string           `json:"ip"`
	History ChangeHistory[T] `json:"history"`
}

// Traversal 遍历
func (ob *Observer[T]) Traversal(c provider.Condition) (int64, interface{}, error) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	// 分页的起止索引
	start := (c.Page - 1) * c.PageSize

	zap.L().Debug("Observer 偏移量", zap.Int64("start", start), zap.Int64("page", c.Page), zap.Int64("size", c.PageSize))
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
	var result []WebResult[T]
	for _, ip := range ips {
		wr := WebResult[T]{
			IP:      ip.Member.(string),
			History: *ob.GetHistory(ip.Member.(string)),
		}
		result = append(result, wr)
	}
	return count, result, nil
}

func Setup() {
	// 程序运行前清空有序集合
	CleanUp()
	go TTLObserver.WatchChange(TTLEvents)
	go MacObserver.WatchChange(MacEvents)
	go UaObserver.WatchChange(UaEvents)
	go DeviceObserver.WatchChange(DeviceEvents)
}

func CleanUp() {
	redis.GetRedisClient().Del(context.TODO(), types.ZSetObserverTTL).Val()
	redis.GetRedisClient().Del(context.TODO(), types.ZSetObserverMac).Val()
	redis.GetRedisClient().Del(context.TODO(), types.ZSetObserverUa).Val()
	redis.GetRedisClient().Del(context.TODO(), types.ZSetObserverDevice).Val()
}
