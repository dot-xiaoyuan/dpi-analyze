package memory

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
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
		// 使用标志避免重复添加
		duplicate := false
		for _, item := range record {
			if item.SrcMac == i.SrcMac {
				duplicate = true
				break
			}
		}
		// 没有找到相同 Mac 地址，追加Mac
		if !duplicate {
			record = append(record, i)
			MacTables.Store(l.IP, record)
		}
		// zap.L().Debug("Update Mac Cache", zap.String("ip", l.IP), zap.Any("Ethernet", i))
	} else {
		// 不存在该IP记录，直接存储
		record := []capture.Ethernet{i}
		MacTables.Store(l.IP, record)
		// zap.L().Debug("Insert Mac Cache", zap.String("ip", l.IP), zap.Any("Ethernet", i))
	}

}
