package match

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/brands/full"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/brands/partial"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"go.uber.org/zap"
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

func BrandMatch(brand, ip string) types.Domain {
	ok, domain := DomainMatch(brand+".com", ip)
	if !ok {
		zap.L().Warn("mobile icon not found", zap.String("brand", brand))
		return types.Domain{}
	}

	return domain
}
