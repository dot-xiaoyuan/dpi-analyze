package handler

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features/application"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features/brands"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features/brands_keyword"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features/brands_root"
)

func FeatureLibrary(raw json.RawMessage) any {
	type FeatureLength struct {
		Name    string `json:"name"`
		Count   int    `json:"count"`
		Version string `json:"version"`
	}

	res := []FeatureLength{
		{Name: "应用特征", Count: len(application.Feature), Version: application.LoaderManger.Version()},
		{Name: "品牌特征", Count: len(brands.Manager.Feature), Version: brands.Manager.Loader.Version()},
		{Name: "品牌关键词特征", Count: len(brands_keyword.Manager.Feature), Version: brands_keyword.Manager.Loader.Version()},
		{Name: "品牌根域名特征", Count: len(brands_root.Manager.Feature), Version: brands_root.Manager.Loader.Version()},
	}

	return res
}
