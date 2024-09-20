package cache

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"go.uber.org/zap"
	"sync"
)

var (
	TTLTables sync.Map
)

const TTLChangeThreshold = 1

type Internet struct {
	IP string
}

func (l *Internet) Update(internet interface{}) {
	i := internet.(capture.Internet)
	value, ok := TTLTables.Load(l.IP)
	if ok {
		// 如果 TTL 存在，检查差异
		record := value.([]capture.Internet)
		for _, item := range record {
			if item.DstIP == i.DstIP {
				continue
			}
			if absDiff(i.TTL, item.TTL) >= TTLChangeThreshold {
				record = append(record, i)
				break
			}
		}
		TTLTables.Store(l.IP, record)
		zap.L().Debug("Update TTL Cache", zap.String("ip", l.IP), zap.Any("Internet", i))
	} else {
		record := []capture.Internet{i}
		TTLTables.Store(l.IP, record)
		zap.L().Debug("Insert TTL Cache", zap.String("ip", l.IP), zap.Any("Internet", i))
	}
}

func (l *Internet) QueryAll(params capture.Params) ([]byte, error) {
	ttlMap := make(map[string][]capture.Internet)
	TTLTables.Range(func(key, value interface{}) bool {
		ttlMap[key.(string)] = value.([]capture.Internet)
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
