package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/spf13/viper"
	"log"
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
	Home                 string
	Cfg                  *Yaml
	Language             string
	LogLevel             string
	Debug                bool
	Geo2IP               string
	CaptureNic           string
	CapturePcap          string
	UseMongo             bool
	UnixSocket           string
	ParseFeature         bool
	BerkeleyPacketFilter string
	IgnoreMissing        bool
	UseTTL               bool
	UseUA                bool
	WebPort              uint
	Detach               bool
)

type Yaml struct {
	Language             string  `mapstructure:"language"`
	LogLevel             string  `mapstructure:"log_level"`
	Debug                bool    `mapstructure:"debug"`
	Detach               bool    `mapstructure:"detach"`
	Geo2IP               string  `mapstructure:"geo2ip"`
	UseMongo             bool    `mapstructure:"use_mongo"`
	UnixSocket           string  `mapstructure:"unix_socket"`
	ParseFeature         bool    `mapstructure:"parse_app"`
	BerkeleyPacketFilter string  `mapstructure:"berkeley_packet_filter"`
	IgnoreMissing        bool    `mapstructure:"ignore_missing"`
	UseTTL               bool    `mapstructure:"use_ttl"`
	UseUA                bool    `mapstructure:"use_ua"`
	Capture              Capture `mapstructure:"capture"`
	Mongodb              Mongodb `mapstructure:"mongodb"`
	Redis                Redis   `mapstructure:"redis"`
	Web                  Web
}

type Capture struct {
	OfflineFile string `mapstructure:"offline_file"`
	NIC         string `mapstructure:"nic"`
	SnapLen     int32  `mapstructure:"snap_len"`
}

type Web struct {
	Port uint `mapstructure:"port"`
}

type Mongodb struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type Redis struct {
	DPI    RedisConfig `mapstructure:"dpi"`
	Online RedisConfig `mapstructure:"online"`
	Cache  RedisConfig `mapstructure:"cache"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func init() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC : %v", err)
			fmt.Println("发生严重错误，请联系支持人员。")
			os.Exit(1)
		}
	}()

	viper.SetConfigType("yaml")

	Home = os.Getenv("DPI_HOME")
	env := os.Getenv("DPI_ENV")
	// 根据环境变量加载不同的配置文件
	if len(env) == 0 {
		err := viper.ReadConfig(bytes.NewReader(YamlConfig))
		if err != nil {
			panic(err)
		}
	} else {
		switch os.Getenv("DPI_ENV") {
		case EnvDevelopment:
			viper.SetConfigFile(DevConfigFileName)
			break
		case EnvProduction:
			viper.SetConfigFile(ProdConfigFileName)
			break
		}
		// 如果环境变量指定了配置文件，则尝试读取它
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
	}
	// 将最终的配置解析到结构体中
	err := viper.Unmarshal(&Cfg)
	if err != nil {
		panic(err)
	}
}
