package sockets

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze/memory"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
)

type ActionInternet struct {
}

func (a *ActionInternet) Handle(data any) []byte {
	// FIXME 偏移量和分页
	ttlMap := make(map[string][]capture.Internet)
	memory.TTLTables.Range(func(key, value interface{}) bool {
		ttlMap[key.(string)] = value.([]capture.Internet)
		return true
	})
	// 2json
	jsonData, err := json.Marshal(ttlMap)
	if err != nil {
		return nil
	}
	return jsonData
}
