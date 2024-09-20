package iptables

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"sync"
	"time"
)

// IP 活动记录

var IPTables sync.Map

type Detail struct {
	IP string
}

func (d Detail) Update(i interface{}) {

}

func (d Detail) QueryAll(params capture.Params) ([]byte, error) {
	ipMap := make(map[string]capture.IPActivityLogs)
	if params.Limit == 0 {
		params.Limit = 10
	}
	var i int
	IPTables.Range(func(key, value interface{}) bool {
		i++
		// limit
		if len(ipMap) >= params.Limit {
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
		return nil, err
	}
	return jsonData, nil
}

// 新增/更新 IP表

func Load(ip string, val any) {
	d, ok := IPTables.Load(ip)
	if ok {
		old_ := d.(capture.IPActivityLogs)
		new_ := val.(capture.IPActivityLogs)
		CompareTTLAndUpdate(&old_, &new_)
		CompareMacAndUpdate(&old_, &new_)
		CompareUaAndUpdate(&old_, &new_)
		// 更新时间
		new_.LastSeen = time.Now()
		IPTables.Store(ip, new_)
	} else {
		IPTables.Store(ip, val)
	}
}

// CompareTTLAndUpdate 比对TTL并更新
func CompareTTLAndUpdate(old_, new_ *capture.IPActivityLogs) {
	// ttl
	if old_.CurrentTTL != new_.CurrentTTL && new_.CurrentTTL > 0 {
		new_.TTLHistory = append(old_.TTLHistory, capture.TTLHistory{
			TTL:       old_.CurrentTTL,
			Timestamp: time.Now(),
		})
	}

}

// CompareMacAndUpdate 比对Mac地址并更新
func CompareMacAndUpdate(old_, new_ *capture.IPActivityLogs) {
	if old_.CurrentMac != new_.CurrentMac && new_.CurrentMac != "" {
		new_.MacHistory = append(old_.MacHistory, capture.MacHistory{
			MacAddress: old_.CurrentMac,
			Timestamp:  time.Now(),
		})
	}
}

// CompareUaAndUpdate 比对Ua并更新
func CompareUaAndUpdate(old_, new_ *capture.IPActivityLogs) {
	if old_.CurrentUserAgent != new_.CurrentUserAgent && new_.CurrentUserAgent != "" {
		new_.UAHistory = append(old_.UAHistory, capture.UAHistory{
			UserAgent: old_.CurrentUserAgent,
			Timestamp: time.Now(),
		})
	}
}
