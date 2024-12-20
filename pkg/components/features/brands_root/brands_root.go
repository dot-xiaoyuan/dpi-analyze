package brands_root

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features/manager"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/parser"
	"strings"
)

// Manager 全局变量
var Manager *manager.Manager

// Setup 初始化
func Setup() error {
	Manager = manager.NewManager(manager.Config{
		Filename:              fmt.Sprintf("%s/brands_root.yaml", config.EtcDir),
		CollectionName:        types.MongoCollectionFeatureBrandsRoot, // 对应 Mongo 集合名
		HistoryCollectionName: types.MongoCollectionFeatureBrandsRootHistory,
		DatabaseName:          types.MongoDatabaseConfigs,
		ParserFunc: func(data []byte) ([]string, map[int]interface{}, error) {
			brands, err := parser.ParseBrands(data)
			if err != nil {
				return nil, nil, err
			}

			var features []string
			mapping := make(map[int]interface{})
			for _, brand := range brands {
				for _, domain := range brand.Domains {
					domain.BrandName = strings.ToLower(brand.BrandName)
					domain.Icon = fmt.Sprintf("icon-%s", domain.BrandName)
					features = append(features, domain.DomainName)
					mapping[len(features)-1] = domain
				}
			}
			return features, mapping, nil
		},
	})

	return Manager.Setup()
}

// Match 匹配
func Match(input string) (ok bool, domain interface{}) {
	return Manager.Match(input)
}
