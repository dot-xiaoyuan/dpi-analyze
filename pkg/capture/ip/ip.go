package ip

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/observer"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"go.uber.org/zap"
	"sync"
)

// IP 相关的核心逻辑

var (
	TTLCache sync.Map
	MacCache sync.Map
	UaCache  sync.Map
	Mutex    sync.Map
)

type EventType int

type Property string

const (
	TTL       Property = "ttl"
	Mac       Property = "mac"
	UserAgent Property = "user-agent"
)

func getIPMutex(ip string) *sync.Mutex {
	val, _ := Mutex.LoadOrStore(ip, &sync.Mutex{})
	return val.(*sync.Mutex)
}

type PropertyChangeEvent struct {
	IP       string
	OldValue any
	NewValue any
	Property Property
}

func ChangeEventIP(events <-chan PropertyChangeEvent) {
	for e := range events {

		mutex := getIPMutex(e.IP)
		mutex.Lock()

		switch e.Property {
		case TTL:
			// process ttl
			zap.L().Debug(i18n.T("TTL Changed"), zap.String("ip", e.IP), zap.Any("old", e.OldValue), zap.Any("new", e.NewValue))
			_ = ants.Submit(func() {
				observer.RecordTTLChange(e.IP, e.NewValue.(uint8))
			})
			break
		case Mac:
			// process mac
			break
		case UserAgent:
			// process ua
			break
		}
		// 更新redis
		storeHash2Redis(e.IP, e.Property, e.NewValue)
		mutex.Unlock()
	}
}
