package cache

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"go.uber.org/zap"
	"sync"
)

var (
	MacTables sync.Map
)

type Ethernet struct {
	IP string
}

func (l *Ethernet) Update(ethernet interface{}) {
	i := ethernet.(capture.Ethernet)
	value, ok := MacTables.Load(l.IP)
	if ok {
		record := value.([]capture.Ethernet)
		for _, item := range record {
			if item.SrcMac == i.SrcMac {
				continue
			}
			record = append(record, i)
		}
		MacTables.Store(l.IP, record)
		zap.L().Debug("Update Mac Cache", zap.String("ip", l.IP), zap.Any("Ethernet", i))
	} else {
		record := []capture.Ethernet{i}
		MacTables.Store(l.IP, record)
		zap.L().Debug("Insert Mac Cache", zap.String("ip", l.IP), zap.Any("Ethernet", i))
	}

}
