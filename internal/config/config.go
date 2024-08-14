package config

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const EnvDevelopmentMode = "dev"
const EnvProductionMode = "prod"
const DevConfigFilename = "config/config.yaml"
const ProdConfigFilename = "/etc/config.prod.yaml"

var (
	Cfg  *Config
	Env  = pflag.StringP("env", "e", EnvDevelopmentMode, "environment name")
	File = pflag.StringP("config", "c", DevConfigFilename, "path to the config file")

	LogLevel           *string
	Debug              *bool
	CaptureNic         *string
	CaptureOfflineFile *string
	UseMongo           *bool
)

// 配置加载管理

type Config struct {
	LogLevel string `mapstructure:"log_level"`
	Debug    bool   `mapstructure:"debug"`
	Capture  Capture
	Mongodb  Mongodb
}

func (c *Config) setDefault() {
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	if c.Capture.SnapLen == 0 {
		c.Capture.SnapLen = 16 << 10
	}
}

func init() {
	// 加载配置
	err := LoadConfig()
	if err != nil {
		zap.L().Panic("Failed to load config:", zap.Error(err))
	}

	LogLevel = pflag.StringP("log_level", "l", Cfg.LogLevel, "Log level")
	Debug = pflag.BoolP("debug", "d", Cfg.Debug, "enable debug mode")

	CaptureOfflineFile = pflag.StringP("pcap_file", "p", Cfg.Capture.OfflineFile, "path to the pcap file")
	CaptureNic = pflag.StringP("nic", "n", Cfg.Capture.NIC, "network interface")

	UseMongo = pflag.BoolP("mongo", "m", Cfg.Mongodb.Use, "enable mongodb")

	pflag.Parse()

}

func LoadConfig() error {
	viper.SetConfigFile(*File)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(&Cfg)
	if err != nil {
		return err
	}

	// 默认值
	Cfg.setDefault()
	// flag 优先
	if *LogLevel != Cfg.LogLevel {
		Cfg.LogLevel = *LogLevel
	}
	// 网卡
	if len(*CaptureNic) > 0 {
		Cfg.Capture.OfflineFile = ""
		Cfg.Capture.NIC = *CaptureNic
	}
	// 离线包优先
	if len(*CaptureOfflineFile) > 0 {
		Cfg.Capture.OfflineFile = *CaptureOfflineFile
	}

	// Default Setting
	if Cfg.Capture.SnapLen == 0 {
		Cfg.Capture.SnapLen = 16 << 10
	}

	if len(Cfg.Mongodb.Host) > 0 {
		mongoUri := fmt.Sprintf("mongodb://%s:%s", Cfg.Mongodb.Host, Cfg.Mongodb.Port)
		if *UseMongo != Cfg.Mongodb.Use {

		}
		db.Setup(mongoUri, *UseMongo)
	}
	return nil

}

// TODO Change Event.
