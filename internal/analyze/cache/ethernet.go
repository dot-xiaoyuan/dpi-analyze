package cache

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"go.uber.org/zap"
	"sync"
)

var (
	MacTables sync.Map
)

type Ethernet struct{}

func (Ethernet) Update(ip string, ethernet interface{}) {
	i := ethernet.(capture.Ethernet)
	value, ok := MacTables.Load(ip)
	if ok {
		record := value.([]capture.Ethernet)
		for _, item := range record {
			if item.SrcMac == i.SrcMac {
				continue
			}
			record = append(record, i)
		}
		MacTables.Store(ip, record)
		zap.L().Debug("Update Mac Cache", zap.String("ip", ip), zap.Any("Ethernet", i))
	} else {
		record := []capture.Ethernet{i}
		MacTables.Store(ip, record)
		zap.L().Debug("Insert Mac Cache", zap.String("ip", ip), zap.Any("Ethernet", i))
	}

}

func (Ethernet) QueryAll() ([]byte, error) {
	macMap := make(map[string][]capture.Ethernet)
	MacTables.Range(func(key, value interface{}) bool {
		macMap[key.(string)] = value.([]capture.Ethernet)
		return true
	})
	// 2json
	jsonData, err := json.Marshal(macMap)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}
