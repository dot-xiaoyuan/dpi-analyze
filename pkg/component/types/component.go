package types

type Component interface {
	Setup() error
}

type AhoCorasick interface {
	newAc()
	Match(origin string) (ok bool, domain Domain)
}

type Aggregation interface {
	GetTotalCount() int64
}

type Domain struct {
	Icon        string `json:"icon"`
	BrandName   string `json:"brand_name"`
	DomainName  string `json:"domain_name" mapstructure:"domain_name"`
	Description string `json:"description" mapstructure:"description"`
}

type DeviceBrand struct {
	BrandName string   `json:"brand_name" mapstructure:"brand_name"`
	Domains   []Domain `json:"domains" mapstructure:"domains"`
}

type DeviceList struct {
	Brands []DeviceBrand `json:"brands" mapstructure:"brands"`
}
