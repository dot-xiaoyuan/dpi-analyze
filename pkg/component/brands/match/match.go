package match

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/brands/full"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/brands/keywords"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/brands/partial"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"go.uber.org/zap"
	"strings"
)

func DomainMatch(origin, ip string) (ok bool, domain types.Domain) {
	// 尝试精确匹配
	ok, domain = full.Brands.ExactMatch(origin)
	if ok {
		// 精确匹配成功
		return ok, domain
	}
	// 如果精确匹配失败，尝试部分匹配
	return partial.Brands.PartialMatch(origin, ip)
}

func BrandMatch(brand, ip string, dr types.DeviceRecord) types.Domain {
	ok, domain := DomainMatch(brand+".com", ip)
	if !ok {
		zap.L().Warn("mobile icon not found", zap.String("brand", brand))
		// 域名未匹配到，进行关键词匹配
		ok, domain = keywords.Brands.PartialMatch(brand, ip)
		if !ok {
			return types.Domain{
				Icon:        fmt.Sprintf("icon-%s", strings.ToLower(brand)),
				BrandName:   brand,
				DomainName:  "",
				Description: "",
			}
		}
	}

	if domain.BrandName == "apple" {
		zap.L().Debug("apple 品牌标识", zap.Any("domain", domain), zap.Any("dr", dr))
		if strings.Contains(strings.ToLower(dr.Os), "mac") || strings.Contains(strings.ToLower(dr.Model), "mac") {
			domain.Icon = "icon-macos"
		}
	}
	return domain
}
