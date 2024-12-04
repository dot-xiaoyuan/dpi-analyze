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
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	res := []FeatureLength{
		{"应用特征", len(application.Feature)},
		{"品牌特征", len(brands.Manager.Feature)},
		{"品牌关键词特征", len(brands_keyword.Manager.Feature)},
		{"品牌根域名特征", len(brands_root.Manager.Feature)},
	}

	return res
}
