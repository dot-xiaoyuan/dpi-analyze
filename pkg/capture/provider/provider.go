package provider

// 数据提供器

type Provider interface {
	Traversal(c Condition) (any, error)
	Store2Redis(ip string)
}

type Condition struct {
	Min      string `json:"min"`
	Max      string `json:"max"`
	Page     int64  `json:"page"`
	PageSize int64  `json:"page_size"`
}
