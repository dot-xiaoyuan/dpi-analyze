package observer

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/types"
	"time"
)

var (
	MacEvents       = make(chan MacChangeObserverEvent, 100)
	MacHistoryCache = make(map[string]*MacChangeHistory)
)

// MacChangeObserverEvent Mac变化观察事件
type MacChangeObserverEvent struct {
	IP   string
	Prev any
	Curr any
}

type MacChange struct {
	Time time.Time `json:"time"`
	Mac  string    `json:"mac"`
}

// MacChangeHistory Mac变化历史记录
type MacChangeHistory struct {
	Changes []MacChange
}

// RecordMacChange 记录mac变化
func RecordMacChange(ip string, mac string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	history, exists := MacHistoryCache[ip]
	if !exists {
		ob := Observer{
			Table: types.ZSetObserverMac,
		}
		ob.Store2Redis(ip)

		history = &MacChangeHistory{
			Changes: make([]MacChange, 0, MaxMacCount),
		}
		MacHistoryCache[ip] = history
	}

	if len(history.Changes) == MaxMacCount {
		history.Changes = history.Changes[1:]
	}
	// 记录变化
	history.Changes = append(history.Changes, MacChange{
		Time: time.Now(),
		Mac:  mac,
	})
}

// GetMacHistory 获取ttl历史记录
func GetMacHistory(ip string) []MacChange {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if history, exists := MacHistoryCache[ip]; exists {
		return history.Changes
	}
	return nil
}

// WatchMacChange 观察ttl变化
func WatchMacChange(events <-chan MacChangeObserverEvent) {
	for e := range events {
		RecordMacChange(e.IP, e.Curr.(string))
	}
}
