package resolve

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/uaparser"
	"go.uber.org/zap"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	useragentLock sync.Mutex
	logQueue      = make(chan types.UserAgentRecord, 10000)
)

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
	logQueue <- record
	//useragentLock.Lock()
	//
	//_, _ = mongo.GetMongoClient().Database(types.MongoDatabaseUserAgent).
	//	Collection(time.Now().Format("06_01_02_useragent")).
	//	InsertOne(context.TODO(), record)
	//useragentLock.Unlock()

	if client.Os.ToString() == "Other" ||
		client.UserAgent.Family == "IE" ||
		len(client.Os.ToVersionString()) == 0 ||
		client.Device.Family == "Other" || client.Device.Brand == "Generic_Android" {
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

	Handle(dr)
	return fmt.Sprintf("%s %s", dr.Os, dr.Version)
}

func StartUserAgentConsumer() {
	go func() {
		buffer := make([]any, 0, 100)
		ticker := time.NewTicker(time.Second) // 每秒批量写入
		defer ticker.Stop()

		for {
			select {
			case log := <-logQueue:
				buffer = append(buffer, log)
				if len(buffer) >= 100 {
					insertManyStream(buffer)
					buffer = buffer[:0]
				}
			case <-ticker.C: // 定时写入
				if len(buffer) > 0 {
					insertManyStream(buffer)
					buffer = buffer[:0]
				}
			}
		}
	}()
}

func insertManyStream(buffer []interface{}) {
	_, err := mongo.GetMongoClient().Database(types.MongoDatabaseUserAgent).Collection(time.Now().Format("06_01_02_useragent")).
		InsertMany(mongo.Context, buffer)
	if err != nil {
		zap.L().Error("insert [useragent] record failed", zap.Error(err))
		return
	}
}
