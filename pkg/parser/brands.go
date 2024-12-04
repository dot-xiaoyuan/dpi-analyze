package parser

import (
	"bufio"
	"bytes"
	"github.com/spf13/viper"
)

type Domain struct {
	Icon        string `json:"icon" mapstructure:"icon"`
	BrandName   string `json:"brand_name" mapstructure:"brand_name"`
	DomainName  string `json:"domain_name" mapstructure:"domain_name"`
	Description string `json:"description" mapstructure:"description"`
}

type DeviceBrands struct {
	BrandName string   `json:"brand_name" mapstructure:"brand_name"`
	Domains   []Domain `json:"domains" mapstructure:"domains"`
}

type BrandsList struct {
	Brands []DeviceBrands `json:"brands" mapstructure:"brands"`
}

func ParseBrands(data []byte) ([]DeviceBrands, error) {
	reader := bufio.NewReader(bytes.NewBuffer(data))
	err := viper.ReadConfig(reader)
	if err != nil {
		return nil, err
	}

	var brandsList BrandsList
	if err = viper.Unmarshal(&brandsList); err != nil {
		return nil, err
	}
	return brandsList.Brands, nil
}
