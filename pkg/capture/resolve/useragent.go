package resolve

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/uaparser"
	"net/url"
	"strings"
	"sync"
	"time"
)

var useragentLock sync.Mutex

func AnalyzeByUserAgent(ip, ua, host string) string {
	client := uaparser.Analyze(ua, host)
	if client == nil {
		return ""
	}
	// TODO 厂商严格分析开关，开启后如果厂商为空直接跳过
	if len(client.Os.Family) == 0 {
		return ""
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

	useragentLock.Lock()
	_, _ = mongo.GetMongoClient().Database(types.MongoDatabaseUserAgent).
		Collection(time.Now().Format("06_01_02_useragent")).
		InsertOne(context.TODO(), record)
	useragentLock.Unlock()

	if client.Os.ToString() == "Other" ||
		client.UserAgent.Family == "IE" ||
		len(client.Os.ToVersionString()) == 0 ||
		client.Device.Family == "Other" {
		return ""
	}
	var brand, icon string
	brand = strings.ToLower(client.Device.Brand)
	if len(brand) > 0 {
		icon = fmt.Sprintf("icon-%s", brand)
		if brand == "apple" && (strings.Contains(strings.ToLower(client.Os.Family), "mac") || strings.Contains(strings.ToLower(client.Device.Model), "windows")) {
			icon = "icon-macos"
		}
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

	DeviceHandle(dr)
	return fmt.Sprintf("%s %s", dr.Os, dr.Version)
}
