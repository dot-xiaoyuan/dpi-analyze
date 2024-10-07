package observer

import (
	"context"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/layers"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	v9 "github.com/redis/go-redis/v9"
	"time"
)

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

// RecordTTLChange 记录ttl变化
func RecordTTLChange(ip string, ttl uint8) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	history, exists := TTLHistoryCache[ip]
	if !exists {
		store2Redis(ip)
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

func store2Redis(ip string) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	rdb.ZAdd(ctx, layers.ZSetObserverTTL, v9.Z{
		Score:  float64(time.Now().Unix()),
		Member: ip,
	})
}
