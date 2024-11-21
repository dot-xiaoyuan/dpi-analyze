package resolve

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/uaparser"
	"net/url"
	"strings"
	"time"
)

func AnalyzeByUserAgent(ip, ua, host string) {
	lock.Lock()
	defer lock.Unlock()
	client := uaparser.Analyze(ua, host)
	if client == nil {
		return
	}
	// TODO 厂商严格分析开关，开启后如果厂商为空直接跳过
	if len(client.Os.Family) == 0 {
		return
	}
	// 记录mongodb
	uaStr, _ := url.QueryUnescape(ua)
	record := types.UserAgentRecord{
		IP:        ip,
		UserAgent: uaStr,
		Host:      host,
		Ua:        client.UserAgent.ToString(),
		UaVersion: client.UserAgent.ToVersionString(),
		Os:        client.Os.ToString(),
		OsVersion: client.Os.ToVersionString(),
		Device:    client.Device.ToString(),
		Brand:     strings.ToLower(client.Device.Brand),
		Model:     client.Device.Model,
		LastSeen:  time.Now(),
	}
	_, _ = mongo.GetMongoClient().Database(types.MongoDatabaseUserAgent).
		Collection(time.Now().Format("06_01_02_useragent")).
		InsertOne(context.TODO(), record)

	if client.Os.ToString() == "Other" || client.UserAgent.Family == "IE" || len(client.Os.ToVersionString()) == 0 {
		return
	}
	var brand, icon string
	brand = strings.ToLower(client.Device.Brand)
	if len(brand) > 0 {
		icon = fmt.Sprintf("icon-%s", brand)
	} else if len(client.Os.Family) > 0 {
		brand = strings.ToLower(client.Os.Family)
		icon = fmt.Sprintf("icon-%s", strings.ToLower(client.Os.Family))
	}
	dr := types.DeviceRecord{
		IP:           ip,
		OriginChanel: types.UserAgent,
		OriginValue:  ua,
		Os:           client.Os.Family,
		Version:      client.Os.ToVersionString(),
		Device:       client.Device.ToString(),
		Brand:        brand,
		Model:        client.Device.Model,
		Icon:         icon,
		Description:  "UserAgent 解析",
		LastSeen:     time.Now(),
	}
	member.Store(member.Hash{
		IP:    ip,
		Field: types.UserAgent,
		Value: fmt.Sprintf("%s %s", dr.Os, dr.Version),
	})

	DeviceHandle(dr)
}
