package member

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/observer"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"sync"
	"time"
)

// IP 相关的核心逻辑

var (
	TTLAnalyze      sync.Map
	TTLCache        sync.Map
	MacCache        sync.Map
	UaCache         sync.Map
	DeviceCache     sync.Map
	DeviceNameCache sync.Map
	DeviceTypeCache sync.Map
	Mutex           sync.Map
	Events          = make(chan PropertyChangeEvent, 100)
)

// IP锁
func getIPMutex(ip string) *sync.RWMutex {
	val, _ := Mutex.LoadOrStore(ip, &sync.RWMutex{})
	return val.(*sync.RWMutex)
}

// PropertyChangeEvent 属性变化事件结构
type PropertyChangeEvent struct {
	IP       string
	OldValue any
	NewValue any
	Property types.Property
}

// ChangeEventIP IP 属性变化事件
func ChangeEventIP(events <-chan PropertyChangeEvent) {
	for e := range events {
		if handler, ok := handlers[e.Property]; ok {
			handler(e)
			// update redis
			storeHash2Redis(e.IP, e.Property, e.NewValue)
		}
	}
}

// 具体属性变化事件触发方法
var handlers = map[types.Property]func(e PropertyChangeEvent){
	types.TTL: func(event PropertyChangeEvent) {
		//zap.L().Debug("TTL Changed", zap.String("IP", event.IP), zap.Any("old", event.OldValue), zap.Any("new", event.NewValue))
		// 发送到 TTL 观察者 Channel
		observer.TTLEvents <- observer.ChangeObserverEvent[uint8]{
			IP:   event.IP,
			Prev: event.OldValue.(uint8),
			Curr: event.NewValue.(uint8),
		}
	},
	types.Mac: func(event PropertyChangeEvent) {
		//zap.L().Debug("MAC Changed", zap.String("IP", event.IP), zap.Any("old", event.OldValue), zap.Any("new", event.NewValue))
		// 发送到 Mac 观察者 Channel
		observer.MacEvents <- observer.ChangeObserverEvent[string]{
			IP:   event.IP,
			Prev: event.OldValue.(string),
			Curr: event.NewValue.(string),
		}
	},
	types.UserAgent: func(event PropertyChangeEvent) {
		//zap.L().Debug("UA Changed", zap.String("IP", event.IP), zap.Any("old", event.OldValue), zap.Any("new", event.NewValue))
		// 发送到 Ua 观察者 Channel
		observer.UaEvents <- observer.ChangeObserverEvent[string]{
			IP:   event.IP,
			Prev: event.OldValue.(string),
			Curr: event.NewValue.(string),
		}
	},
	types.Device: func(event PropertyChangeEvent) {
		//zap.L().Debug("UA Changed", zap.String("IP", event.IP), zap.Any("old", event.OldValue), zap.Any("new", event.NewValue))
		// 发送到 Ua 观察者 Channel
		observer.DeviceEvents <- observer.ChangeObserverEvent[types.DeviceRecord]{
			IP:   event.IP,
			Prev: event.OldValue.(types.DeviceRecord),
			Curr: event.NewValue.(types.DeviceRecord),
		}
	},
}

func Setup() {
	go ChangeEventIP(Events)
	EnsureIndexOnce()
	go StartFlushScheduler(time.Minute)
}
