package observer

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/types"
	"time"
)

var (
	UaEvents       = make(chan UaChangeObserverEvent, 100)
	UaHistoryCache = make(map[string]*UaChangeHistory)
)

// UaChangeObserverEvent UA变化观察事件
type UaChangeObserverEvent struct {
	IP   string
	Prev any
	Curr any
}

type UaChange struct {
	Time time.Time `json:"time"`
	Ua   string    `json:"ua"`
}

// UaChangeHistory Ua变化历史记录
type UaChangeHistory struct {
	Changes []UaChange
}

// RecordUaChange 记录ua变化
func RecordUaChange(ip string, ua string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	history, exists := UaHistoryCache[ip]
	if !exists {
		ob := Observer{
			Table: types.ZSetObserverUa,
		}
		ob.Store2Redis(ip)

		history = &UaChangeHistory{
			Changes: make([]UaChange, 0, MaxMacCount),
		}
		UaHistoryCache[ip] = history
	}

	if len(history.Changes) == MaxMacCount {
		history.Changes = history.Changes[1:]
	}
	// 记录变化
	history.Changes = append(history.Changes, UaChange{
		Time: time.Now(),
		Ua:   ua,
	})
}

// GetUaHistory 获取ttl历史记录
func GetUaHistory(ip string) []UaChange {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if history, exists := UaHistoryCache[ip]; exists {
		return history.Changes
	}
	return nil
}

// WatchUaChange 观察ttl变化
func WatchUaChange(events <-chan UaChangeObserverEvent) {
	for e := range events {
		RecordUaChange(e.IP, e.Curr.(string))
	}
}
