package config

import (
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const EnvDevelopmentMode = "dev"
const EnvProductionMode = "prod"
const DevConfigFilename = "config/config.yaml"
const ProdConfigFilename = "/etc/config.prod.yaml"

var (
	Env      = flag.StringP("env", "e", EnvProductionMode, "environment name")
	File     = flag.StringP("config", "c", DevConfigFilename, "path to the config file")
	LogLevel = flag.StringP("log_level", "l", "info", "Log level")
	PcapFile = flag.StringP("pcap_file", "p", "", "path to the pcap file")
	Debug    = flag.BoolP("debug", "d", false, "enable debug mode")
	Nic      = flag.StringP("nic", "n", "", "network interface")
	Cfg      *Config
)

// 配置加载管理

type Config struct {
	LogLevel string `mapstructure:"log_level"`
	Capture  Capture
}

func init() {
	flag.Parse()

	if *File == DevConfigFilename && *Env == EnvProductionMode {
		*File = ProdConfigFilename
	}
	if *Debug {
		*LogLevel = "debug"
	}
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
	// flag 优先
	if *LogLevel != Cfg.LogLevel {
		Cfg.LogLevel = *LogLevel
	}
	// 网卡
	if len(*Nic) > 0 {
		Cfg.Capture.OfflineFile = ""
		Cfg.Capture.NIC = *Nic
	}
	// 离线包优先
	if len(*PcapFile) > 0 {
		Cfg.Capture.OfflineFile = *PcapFile
	}

	// Default Setting
	if Cfg.Capture.SnapLen == 0 {
		Cfg.Capture.SnapLen = 16 << 10
	}
	return nil

}

// TODO Change Event.
