package features

import (
	"github.com/cloudflare/ahocorasick"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"sync"
)

// 手机设备特征

var (
	once              sync.Once
	MobileAhoCorasick *ahocorasick.Matcher
	MobileFeature     []string                     // 手机特征切片
	MobileDeviceMap   = make(map[int]Manufacturer) // 厂商映射关系
)

type Manufacturer struct {
	Name string
	Icon string
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
