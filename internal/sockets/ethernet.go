package sockets

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze/memory"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
)

type ActionEthernet struct{}

func (ActionEthernet) Handle(data json.RawMessage) []byte {
	// FIXME 偏移量与分页
	macMap := make(map[string][]capture.Ethernet)
	memory.MacTables.Range(func(key, value interface{}) bool {
		macMap[key.(string)] = value.([]capture.Ethernet)
		return true
	})
	// 2json
	jsonData, err := json.Marshal(macMap)
	if err != nil {
		return nil
	}
	return jsonData
}
