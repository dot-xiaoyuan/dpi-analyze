package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//go:embed dpi.yaml
var YamlConfig []byte

var (
	RunDir string
	LogDir string
	EtcDir string
	BinDir string
)

var (
	Home                 string
	Signal               string
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
	if Home == "" {
		Home = "./dev_home"
	}
	Reload()
}

func Reload() {
	// 根据环境变量加载不同的配置文件
	RunDir = filepath.Join(Home, "run")
	LogDir = filepath.Join(Home, "log")
	EtcDir = filepath.Join(Home, "etc")
	BinDir = filepath.Join(Home, "bin")

	ensureDirExists(RunDir)
	ensureDirExists(LogDir)
	ensureDirExists(EtcDir)
	ensureDirExists(BinDir)

	if strings.Contains(Home, "dev_home") {
		err := viper.ReadConfig(bytes.NewReader(YamlConfig))
		if err != nil {
			panic(err)
		}
	} else {
		viper.SetConfigFile(fmt.Sprintf("%s/dpi.yaml", EtcDir))
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

func ensureDirExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			panic(fmt.Sprintf("Failed to create directory %s: %v", dir, err))
		}
	}
}
