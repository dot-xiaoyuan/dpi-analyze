package parser

import (
	"bufio"
	"bytes"
	"github.com/spf13/viper"
)

type DeviceBrandsKeyword struct {
	BrandName   string   `json:"brand_name" mapstructure:"brand_name"`
	Keywords    []string `json:"keywords" mapstructure:"keywords"`
	Description string   `json:"description" mapstructure:"description"`
}

type BrandsKeywordList struct {
	Version string                `json:"version" mapstructure:"version"`
	Brands  []DeviceBrandsKeyword `json:"brands" mapstructure:"brands"`
}

func ParseBrandsKeyword(data []byte) ([]DeviceBrandsKeyword, error) {
	reader := bufio.NewReader(bytes.NewBuffer(data))
	err := viper.ReadConfig(reader)
	if err != nil {
		return nil, err
	}

	var brandsList BrandsKeywordList
	if err = viper.Unmarshal(&brandsList); err != nil {
		return nil, err
	}
	return brandsList.Brands, nil
}
