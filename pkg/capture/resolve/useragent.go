package resolve

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/uaparser"
	"go.uber.org/zap"
)

func AnalyzeByUserAgent(ip, ua, host string) {
	client := uaparser.Analyze(ua, host)
	if client == nil {
		return
	}
	// TODO 厂商严格分析开关，开启后如果厂商为空直接跳过
	//if client.Os {
	//
	//}
	if len(client.Os.Family) == 0 {
		return
	}
	if client.Os.ToString() == "Other" {
		return
	}
	UserAgent := types.DeviceRecord{
		Os:      client.Os.Family,
		Version: client.Os.ToVersionString(),
		Device:  client.Device.ToString(),
		Brand:   client.Device.Brand,
		Model:   client.Device.Model,
	}
	member.Store(member.Hash{
		IP:    ip,
		Field: types.UserAgent,
		Value: fmt.Sprintf("%s %s", UserAgent.Os, UserAgent.Version),
	})

	zap.L().Debug("UserAgent", zap.Any("UserAgent", UserAgent))
}
