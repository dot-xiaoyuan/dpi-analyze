package memory

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"sync"
)

const MaxLists = 7

var (
	Table = sync.Map{}
	Lists = make([]string, 7)
)

type Traffic struct {
	Date string
}

func (t *Traffic) Update(transmission interface{}) {
	i := transmission.(types.Transmission)
	value, ok := Table.Load(t.Date)
	if ok {
		record := value.(types.Transmission)
		record.UpStream += i.UpStream
		record.DownStream += i.DownStream
		Table.Store(t.Date, record)
	} else {
		// 不存在该IP记录，直接存储
		Table.Store(t.Date, i)
		// 删除头部元素
		if len(Lists) >= MaxLists {
			Lists = Lists[1:]
		}
		Lists = append(Lists, t.Date)
	}
}

type Record struct {
	Date  string `json:"date"`
	Type  string `json:"type"`
	Value int64  `json:"value"`
}

func GenerateChartData() []Record {
	var result []Record
	for _, date := range Lists {
		if v, ok := Table.Load(date); ok {
			temp := v.(types.Transmission)
			result = append(result, Record{
				Date:  date,
				Type:  "up_stream",
				Value: temp.UpStream,
			}, Record{
				Date:  date,
				Type:  "down_stream",
				Value: temp.DownStream,
			})
		}
	}
	return result
}
