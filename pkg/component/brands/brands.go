package brands

import (
	"errors"
	"fmt"
	"github.com/allegro/bigcache"
	"github.com/cloudflare/ahocorasick"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"strings"
	"sync"
)

type Brands struct {
	once        sync.Once
	initialized bool
	ahoCorasick *ahocorasick.Matcher
	lists       types.DeviceList
	domains     []string
	maps        map[int]types.Domain
	configFile  string // 配置文件路径
	ipCache     *bigcache.BigCache
}

func NewBrands(configFile string) *Brands {
	return &Brands{
		configFile: configFile,
	}
}

func (b *Brands) Setup() error {
	var setupErr error
	b.once.Do(func() {
		if b.initialized {
			setupErr = fmt.Errorf("brands already initialized")
			return
		}
		// 加载yaml
		viper.SetConfigFile(fmt.Sprintf("%s/%s", config.EtcDir, b.configFile))
		if err := viper.ReadInConfig(); err != nil {
			zap.L().Error(err.Error())
			setupErr = err
			return
		}

		err := viper.Unmarshal(&b.lists)
		if err != nil {
			zap.L().Error(err.Error())
			setupErr = err
			return
		}
		b.maps = make(map[int]types.Domain)
		// 加载ac自动机
		zap.L().Info("Brands initialized successful!", zap.String("file", b.configFile))
		b.newAc()
		b.initialized = true
	})
	return setupErr
}

// 创建Ac自动机
func (b *Brands) newAc() {
	var domains []string
	maps := make(map[int]types.Domain, len(b.lists.Brands)) // 预分配 maps 容量

	for _, brand := range b.lists.Brands {
		for _, domain := range brand.Domains {
			// 品牌名转小写
			domain.BrandName = strings.ToLower(brand.BrandName)
			domain.Icon = fmt.Sprintf("icon-%s", domain.BrandName)
			domains = append(domains, domain.DomainName)
			maps[len(domains)-1] = domain
		}
	}
	b.domains = domains
	b.maps = maps
	// 创建ac自动机
	b.ahoCorasick = ahocorasick.NewStringMatcher(b.domains)
}

func (b *Brands) ExactMatch(origin string) (ok bool, domain types.Domain) {
	hits := b.ahoCorasick.Match([]byte(origin))
	if hits == nil {
		return false, types.Domain{}
	}
	if domain, ok = b.maps[hits[0]]; ok {
		return true, domain
	}
	return false, types.Domain{}
}

// PartialMatch 部分匹配：需要 IP 计数
func (b *Brands) PartialMatch(origin string, ip string) (ok bool, domain types.Domain) {
	hits := b.ahoCorasick.Match([]byte(origin))
	if hits == nil {
		return false, types.Domain{}
	}

	if domain, ok = b.maps[hits[0]]; ok {
		// 检查是否是以 www 开头的官网域名
		if len(domain.DomainName) > 4 && domain.DomainName[:4] == "www." {
			// 如果是 www 开头的域名，则重新计数
			b.resetIpCount(ip)
			return false, types.Domain{}
		}

		// 获取 IP 的访问次数
		count, err := b.getIpCount(ip)
		if err != nil {
			zap.L().Error("Error fetching IP count", zap.Error(err))
			return false, types.Domain{}
		}

		// 判断是否达到了 5 次
		if count >= 5 {
			// 达到阈值则返回成功
			b.resetIpCount(ip) // 重置计数
			return true, domain
		}
	}

	// 如果不匹配或计数未达到 5 次，返回失败
	return false, types.Domain{}
}

// 获取 IP 的计数
func (b *Brands) getIpCount(ip string) (int, error) {
	count, err := b.ipCache.Get(ip)
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			return 0, nil // 如果没有找到，说明 IP 没有访问过
		}
		return 0, err
	}
	// 转换为整数
	return int(count[0] - '0'), nil
}

// 增加 IP 计数
func (b *Brands) incrementIpCount(ip string) error {
	count, err := b.getIpCount(ip)
	if err != nil {
		return err
	}

	// 增加计数
	count++

	// 保存回缓存
	return b.ipCache.Set(ip, []byte(fmt.Sprintf("%d", count)))
}

// 重置 IP 计数
func (b *Brands) resetIpCount(ip string) {
	b.ipCache.Set(ip, []byte("0")) // 重置为 0
}

// Match Match：先进行精确匹配，再进行部分匹配
func (b *Brands) Match(origin string, ip string) (ok bool, domain types.Domain) {
	// 先进行精确匹配
	if ok, domain = b.ExactMatch(origin); ok {
		return true, domain
	}
	// 精确匹配失败，再进行部分匹配
	return b.PartialMatch(origin, ip)
}
