package analyze

import (
	"encoding/json"
	"go.uber.org/zap"
	"sync"
)

// Ethernet 以太网
type Ethernet struct {
	SrcMac string
	DstMac string
}

// Internet 网络层
type Internet struct {
	DstIP string
	TTL   uint8
}

var (
	TTLTables sync.Map
)

const TTLChangeThreshold = 5

func update(ip string, i Internet) {
	value, ok := TTLTables.Load(ip)
	if ok {
		// 如果 TTL 存在，检查差异
		record := value.(Internet)
		if absDiff(i.TTL, record.TTL) >= TTLChangeThreshold {
			TTLTables.Store(ip, Internet{
				DstIP: i.DstIP,
				TTL:   i.TTL,
			})
			zap.L().Debug("Update TTL Cache", zap.String("ip", ip), zap.Any("Internet", i))
		}
	} else {
		TTLTables.Store(ip, i)
		zap.L().Debug("Insert TTL Cache", zap.String("ip", ip), zap.Any("Internet", i))
	}
}

func QueryAll() ([]byte, error) {
	ttlMap := make(map[string]Internet)
	TTLTables.Range(func(key, value interface{}) bool {
		ttlMap[key.(string)] = value.(Internet)
		return true
	})
	// 2json
	jsonData, err := json.Marshal(ttlMap)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func absDiff(new, old uint8) uint8 {
	if new > old {
		return new - old
	}
	return old - new
}
