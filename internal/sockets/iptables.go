package sockets

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze/iptables"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
)

type ActionIPTables struct{}

func (ActionIPTables) Handle(data json.RawMessage) []byte {
	// FIXME offset
	ipMap := make(map[string]capture.IPActivityLogs)
	req := ListReq{}
	_ = json.Unmarshal(data, &req)
	if req.Limit == 0 {
		req.Limit = 10
	}
	var i int
	iptables.IPTables.Range(func(key, value interface{}) bool {
		i++
		// limit
		if len(ipMap) >= req.Limit {
			return false
		}
		// TODO offset
		//if i < params.Offset && params.Offset != 0 {
		//	return false
		//}
		ipMap[key.(string)] = value.(capture.IPActivityLogs)
		return true
	})
	// 2json
	jsonData, err := json.Marshal(ipMap)
	if err != nil {
		return nil
	}
	return jsonData
}
