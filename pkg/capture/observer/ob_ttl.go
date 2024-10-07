package observer

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/types"
	"sync"
	"time"
)

var (
	TTLEvents       = make(chan ChangeObserverEvent, 100)
	TTLHistoryCache = make(map[string]*TTLChangeHistory)
	cacheMutex      sync.Mutex
)

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
		ob := Observer{Table: types.ZSetObserverTTL}
		ob.Store2Redis(ip)

		history = &TTLChangeHistory{
			Changes: make([]TTLChange, 0, MaxTTLCount),
		}
		TTLHistoryCache[ip] = history
	}

	if len(history.Changes) == MaxTTLCount {
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
func WatchTTLChange(events <-chan ChangeObserverEvent) {
	for e := range events {
		RecordTTLChange(e.IP, e.Curr.(uint8))
	}
}
