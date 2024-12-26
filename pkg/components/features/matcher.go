package features

import (
	"errors"
	"fmt"
	"github.com/allegro/bigcache"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features/brands"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features/brands_keyword"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features/brands_root"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/parser"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

var (
	ipCache, _ = bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
)

func HandleFeatureMatch(input, ip string, dr types.DeviceRecord) (ok bool, domain parser.Domain) {
	// 使用 IP 计数逻辑和品牌匹配逻辑
	ok, result, _ := Match(input, ip, ipCounter)
	if ok {
		domain = result.(parser.Domain)
		//zap.L().Debug("Brand matched Result", zap.String("input", input), zap.String("source", source), zap.String("brand", domain.DomainName))
		// 特殊处理
		// macos修改icon
		if domain.BrandName == "apple" {
			if strings.Contains(strings.ToLower(dr.Os), "mac") || strings.Contains(strings.ToLower(dr.Model), "mac") {
				domain.Icon = "icon-macos"
			}
		}
		return ok, domain
	}

	// 未匹配到任何品牌
	//zap.L().Warn("No match found for input", zap.String("input", input), zap.String("ip", ip))
	return ok, parser.Domain{
		Icon:      fmt.Sprintf("icon-%s", strings.ToLower(input)),
		BrandName: input,
	}
}

func Match(input, ip string, ipCounter func(string, string) (bool, error)) (ok bool, result any, matchSource string) {
	// 优先精确匹配，无需计数
	if ok, result = brands.Match(input); ok {
		return true, result, "exact"
	}

	// 关键词匹配，进行计数
	if ok, result = brands_keyword.Match(input); ok {
		domain := result.(parser.Domain)
		if ipCounter != nil {
			if thresholdReached, _ := ipCounter(ip, domain.BrandName); thresholdReached {
				return true, result, "keyword"
			}
		}
	}

	// 根域名匹配，进行计数
	if ok, result = brands_root.Match(input); ok {
		domain := result.(parser.Domain)
		if ipCounter != nil {
			if thresholdReached, _ := ipCounter(ip, domain.BrandName); thresholdReached {
				return true, result, "root"
			}
		}
	}

	return false, parser.Domain{}, "none"
}

func ipCounter(ip, source string) (bool, error) {
	// 构造缓存键，区分来源
	cacheKey := fmt.Sprintf("%s:%s", source, ip)

	// 获取当前计数
	count, err := getIpCount(cacheKey)
	if err != nil {
		return false, err
	}
	count++

	// 达到阈值
	if count >= 50 {
		// 重置计数
		zap.L().Debug("threshold reached", zap.Int("input", count), zap.String("source", source))
		resetIpCount(cacheKey)
		return true, nil
	}

	// 未达阈值，更新计数
	err = incrementIpCount(cacheKey)
	return false, err
}

func getIpCount(key string) (int, error) {
	count, err := ipCache.Get(key)
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			return 0, nil // 如果没有找到，说明 IP 没有访问过
		}
		return 0, err
	}
	return strconv.Atoi(string(count))
}

func incrementIpCount(key string) error {
	count, err := getIpCount(key)
	if err != nil {
		return err
	}
	count++
	return ipCache.Set(key, []byte(strconv.Itoa(count)))
}

func resetIpCount(key string) {
	ipCache.Set(key, []byte("0")) // 重置为 0
}
