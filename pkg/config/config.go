package config

import (
	"bytes"
	_ "embed"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/spf13/viper"
	"os"
)

//go:embed dpi.yaml
var YamlConfig []byte

const (
	EnvDevelopment     = "dev"
	EnvProduction      = "prod"
	DevConfigFileName  = "etc/dpi.yaml"
	ProdConfigFileName = "/etc/dpi.yaml"
)

var (
	Cfg         *Yaml
	Translate   *i18n.Translator
	Language    string
	LogLevel    string
	Debug       bool
	CaptureNic  string
	CapturePcap string
	UseMongo    bool
)

type Yaml struct {
	Language string `mapstructure:"language"`
	LogLevel string `mapstructure:"log_level"`
	Debug    bool   `mapstructure:"debug"`
	UseMongo bool   `mapstructure:"use_mongo"`
	Capture
	Mongodb
}

type Capture struct {
	OfflineFile string `mapstructure:"offline_file"`
	NIC         string `mapstructure:"nic"`
	SnapLen     int32  `mapstructure:"snap_len"`
}

type Mongodb struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

func init() {
	viper.SetConfigFile("yaml")
	// 默认读取编译进去的配置
	err := viper.ReadConfig(bytes.NewReader(YamlConfig))
	if err != nil {
		panic(err)
	}
	// 根据环境变量加载不同的配置文件
	switch os.Getenv("DPI_ENV") {
	case EnvDevelopment:
		viper.SetConfigFile(DevConfigFileName)
	case EnvProduction:
		viper.SetConfigFile(ProdConfigFileName)
	}

	// 如果环境变量指定了配置文件，则尝试读取它
	if err = viper.ReadInConfig(); err != nil {

	}

	// 将最终的配置解析到结构体中
	err = viper.Unmarshal(&Cfg)
	if err != nil {
		panic(err)
	}
}
