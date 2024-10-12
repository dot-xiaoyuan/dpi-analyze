package provider

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/types"
)

// 数据提供器

type Condition struct {
	Min      string         `json:"min"`
	Max      string         `json:"max"`
	Page     int64          `json:"page"`
	PageSize int64          `json:"page_size"`
	Type     types.Property `json:"type"`
}

type UserEventCondition struct {
	Page      int64  `json:"page"`
	PageSize  int64  `json:"page_size"`
	Year      string `json:"year"`
	Month     string `json:"month"`
	SortField string `json:"sort_field"`
	OrderBy   string `json:"order_by"`
}
