package resolve

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/uaparser"
	"strings"
	"time"
)

func AnalyzeByUserAgent(ip, ua, host string) {
	client := uaparser.Analyze(ua, host)
	if client == nil {
		return
	}
	// TODO 厂商严格分析开关，开启后如果厂商为空直接跳过
	if len(client.Os.Family) == 0 {
		return
	}
	if client.Os.ToString() == "Other" || client.UserAgent.Family == "IE" || client.Device.Brand == "Generic_Android" {
		return
	}
	dr := types.DeviceRecord{
		IP:           ip,
		OriginChanel: types.UserAgent,
		OriginValue:  ua,
		Os:           client.Os.Family,
		Version:      client.Os.ToVersionString(),
		Device:       client.Device.ToString(),
		Brand:        strings.ToLower(client.Device.Brand),
		Model:        client.Device.Model,
		LastSeen:     time.Now(),
	}
	member.Store(member.Hash{
		IP:    ip,
		Field: types.UserAgent,
		Value: fmt.Sprintf("%s %s", dr.Os, dr.Version),
	})

	DeviceHandle(dr)
}
