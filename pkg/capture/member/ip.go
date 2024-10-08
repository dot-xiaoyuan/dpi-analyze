package member

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/observer"
	"go.uber.org/zap"
	"sync"
)

// IP 相关的核心逻辑

var (
	TTLCache sync.Map
	MacCache sync.Map
	UaCache  sync.Map
	Mutex    sync.Map
	Events   = make(chan PropertyChangeEvent, 100)
)

type EventType int

type Property string

const (
	TTL       Property = "ttl"
	Mac       Property = "mac"
	UserAgent Property = "user-agent"
)

func Setup() {
	_ = ants.Submit(func() {
		ChangeEventIP(Events)
	})
}

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

var handlers = map[Property]func(e PropertyChangeEvent){
	TTL: func(event PropertyChangeEvent) {
		zap.L().Debug("TTL Changed", zap.String("IP", event.IP), zap.Any("old", event.OldValue), zap.Any("new", event.NewValue))
		_ = ants.Submit(func() {
			// 发送到 TTL 观察者 Channel
			observer.TTLEvents <- observer.ChangeObserverEvent{
				IP:   event.IP,
				Prev: event.OldValue,
				Curr: event.NewValue,
			}
		})
		mutex := getIPMutex(event.IP)
		mutex.Lock()
		storeHash2Redis(event.IP, event.Property, event.NewValue)
		mutex.Unlock()
	},
	Mac: func(event PropertyChangeEvent) {
		zap.L().Debug("MAC Changed", zap.String("IP", event.IP), zap.Any("old", event.OldValue), zap.Any("new", event.NewValue))
		_ = ants.Submit(func() {
			// 发送到 Mac 观察者 Channel
			observer.MacEvents <- observer.MacChangeObserverEvent{
				IP:   event.IP,
				Prev: event.OldValue,
				Curr: event.NewValue,
			}
		})
	},
	UserAgent: func(event PropertyChangeEvent) {
		zap.L().Debug("UA Changed", zap.String("IP", event.IP), zap.Any("old", event.OldValue), zap.Any("new", event.NewValue))
		_ = ants.Submit(func() {
			// 发送到 Mac 观察者 Channel
			observer.MacEvents <- observer.MacChangeObserverEvent{
				IP:   event.IP,
				Prev: event.OldValue,
				Curr: event.NewValue,
			}
		})
	},
}

func ChangeEventIP(events <-chan PropertyChangeEvent) {
	for e := range events {
		mutex := getIPMutex(e.IP)
		if handler, ok := handlers[e.Property]; ok {
			handler(e)
			// update redis
			mutex.Lock()
			storeHash2Redis(e.IP, e.Property, e.NewValue)
			mutex.Unlock()
		}
	}
}
