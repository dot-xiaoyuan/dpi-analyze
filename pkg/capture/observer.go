package capture

import (
	"sync"
	"time"
)

type TTLChangeObserverEvent struct {
	IP   string
	Prev any
	Curr any
	Diff uint8
}

type TTLChange struct {
	Time time.Time
	TTL  uint8
}

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
		history = &TTLChangeHistory{
			Changes: make([]TTLChange, 0),
		}
		TTLHistoryCache[ip] = history
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
		//diff := utils.AbsDiff(e.Curr.(uint8), e.Prev.(uint8))
		//zap.L().Debug("diff", zap.Uint8("diff", diff))
	}
}
