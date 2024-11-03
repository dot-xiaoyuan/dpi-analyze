package features

import (
	"encoding/json"
	"fmt"
	"github.com/allegro/bigcache"
	"github.com/cloudflare/ahocorasick"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 手机设备特征

var (
	once              sync.Once
	MobileAhoCorasick *ahocorasick.Matcher
	MobileFeature     []string                     // 手机特征切片
	MobileDeviceMap   = make(map[int]Manufacturer) // 厂商映射关系
	deviceCache, _    = bigcache.NewBigCache(bigcache.DefaultConfig(time.Minute))
)

type Manufacturer struct {
	Name string
	Icon string
}

func (mf Manufacturer) String() string {
	jsonData, _ := json.Marshal(mf)
	return string(jsonData)
}

// LoadMobile2Ac 加载手机设备特征到ac自动机
func LoadMobile2Ac() {
	once.Do(func() {
		for name, mdf := range config.Cfg.MobileDeviceFeature {
			for _, domain := range mdf.Domains {
				MobileFeature = append(MobileFeature, domain)
				mf := Manufacturer{
					Name: name,
					Icon: mdf.Icon,
				}
				MobileDeviceMap[len(MobileFeature)-1] = mf
			}
		}
		MobileAhoCorasick = ahocorasick.NewStringMatcher(MobileFeature)
	})
}

// DeviceMatch 设备匹配
func DeviceMatch(s string) (ok bool, mf Manufacturer) {
	hits := MobileAhoCorasick.Match([]byte(s))
	if hits == nil {
		return false, mf
	}
	if mf, ok = MobileDeviceMap[hits[0]]; ok {
		return true, mf
	}
	return false, mf
}

func GetDeviceCounter(ip, sni string) (counter int, mf Manufacturer) {
	ok, mf := DeviceMatch(sni)
	if !ok {
		return 0, mf
	}
	// 获取当前计数
	cacheKey := fmt.Sprintf("%s:%s", ip, mf.Name)
	entry, err := deviceCache.Get(cacheKey)
	count := 0
	if err == nil {
		count, _ = strconv.Atoi(string(entry))
		zap.L().Debug("entry", zap.ByteString("entry", entry), zap.Int("count", count))
	}
	// 计数加 1
	count++
	// 更新缓存
	if strings.HasPrefix(sni, "www") {
		count = -10
	}
	zap.L().Debug("device", zap.String("sni", sni), zap.Int("count", count))
	_ = deviceCache.Set(cacheKey, []byte(strconv.Itoa(count)))
	return count, mf
}
