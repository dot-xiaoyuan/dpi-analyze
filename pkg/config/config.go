package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"github.com/spf13/viper"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed dpi.yaml
var YamlConfig []byte

var (
	RunDir    string
	LogDir    string
	EtcDir    string
	BinDir    string
	UploadDir string
)

var (
	Version               string
	Home                  string
	Cfg                   *Yaml
	Language              string
	LogLevel              string
	Debug                 bool
	Geo2IP                string
	UaRegular             string
	CaptureNic            string
	CapturePcap           string
	BerkeleyPacketFilter  string
	IgnoreMissing         bool
	FollowOnlyOnlineUsers bool
	UseTTL                bool
	UseUA                 bool
	UseFeature            bool
	WebPort               uint
	Detach                bool
	IPNet                 *net.IPNet
)

type Yaml struct {
	MachineID             string     `mapstructure:"machine_id" binding:"machine_id" json:"machine_id"`
	Language              string     `mapstructure:"language" bson:"language" json:"language"`
	LogLevel              string     `mapstructure:"log_level" bson:"log_level" json:"log_level"`
	Debug                 bool       `mapstructure:"debug" bson:"debug" json:"debug"`
	Detach                bool       `mapstructure:"detach" bson:"detach" json:"detach"`
	Geo2IP                string     `mapstructure:"geo2ip" bson:"geo2ip" json:"geo2ip"`
	UaRegular             string     `mapstructure:"ua_regular" bson:"ua_regular" json:"ua_regular"`
	BerkeleyPacketFilter  string     `mapstructure:"berkeley_packet_filter" bson:"berkeley_packet_filter" json:"berkeley_packet_filter"`
	IgnoreMissing         bool       `mapstructure:"ignore_missing" bson:"ignore_missing" json:"ignore_missing"`
	FollowOnlyOnlineUsers bool       `mapstructure:"follow_only_online_users" bson:"follow_only_online_users" json:"follow_only_online_users"`
	UseTTL                bool       `mapstructure:"use_ttl" bson:"use_ttl" json:"use_ttl"`
	UseUA                 bool       `mapstructure:"use_ua" bson:"use_ua" json:"use_ua"`
	UseFeature            bool       `mapstructure:"use_feature" bson:"use_feature" json:"use_feature"`
	Capture               Capture    `mapstructure:"capture" bson:"capture" json:"capture"`
	Mongodb               Mongodb    `mapstructure:"mongodb" bson:"mongodb" json:"mongodb"`
	Redis                 Redis      `mapstructure:"redis" bson:"redis" json:"redis"`
	Web                   Web        `mapstructure:"web" bson:"web" json:"web"`
	IgnoreFeature         []string   `mapstructure:"ignore_feature" bson:"ignore_feature" json:"ignore_feature"`
	Thresholds            Thresholds `mapstructure:"thresholds" bson:"thresholds" json:"thresholds"`
	Username              string     `mapstructure:"username" bson:"username" json:"username"`
	Password              string     `mapstructure:"password" bson:"password" json:"password"`
	License               License    `mapstructure:"license" bson:"license" json:"license"`
}

type Capture struct {
	OfflineFile string `mapstructure:"offline_file" bson:"offline_file" json:"offline_file"`
	NIC         string `mapstructure:"nic" bson:"nic" json:"nic"`
	SnapLen     int32  `mapstructure:"snap_len" bson:"snap_len" json:"snap_len"`
}

type Web struct {
	Port uint `mapstructure:"port"`
}

type Thresholds struct {
	SNI         ProtocolFeature `mapstructure:"sni" bson:"sni" json:"sni"`
	HTTP        ProtocolFeature `mapstructure:"http" bson:"http" json:"http"`
	TLSVersion  ProtocolFeature `mapstructure:"tls_version" bson:"tls_version" json:"tls_version"`
	CipherSuite ProtocolFeature `mapstructure:"cipher_suite" bson:"cipher_suite" json:"cipher_suite"`
	Session     ProtocolFeature `mapstructure:"session" bson:"session" json:"session"`
	DNS         ProtocolFeature `mapstructure:"dns" bson:"dns" json:"dns"`
	QUIC        ProtocolFeature `mapstructure:"quic" bson:"quic" json:"quic"`
	SNMP        ProtocolFeature `mapstructure:"snmp" bson:"snmp" json:"snmp"`
}

type ProtocolFeature struct {
	Threshold int    `mapstructure:"threshold" bson:"threshold" json:"threshold"`
	Normal    string `mapstructure:"normal" bson:"normal" json:"normal"`
	Remark    string `mapstructure:"remark" bson:"remark" json:"remark"`
}

type Mongodb struct {
	Host string `mapstructure:"host" bson:"host" json:"host"`
	Port string `mapstructure:"port" bson:"port" json:"port"`
}

type Redis struct {
	DPI    RedisConfig `mapstructure:"dpi" bson:"dpi" json:"dpi"`
	Online RedisConfig `mapstructure:"online" bson:"online" json:"online"`
	Cache  RedisConfig `mapstructure:"cache" bson:"cache" json:"cache"`
	Users  RedisConfig `mapstructure:"users" bson:"users" json:"users"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host" bson:"host" json:"host"`
	Port     string `mapstructure:"port" bson:"port" json:"port"`
	Password string `mapstructure:"password" bson:"password" json:"password"`
	DB       int    `mapstructure:"db" bson:"db" json:"db"`
}

type License struct {
	Sn         string
	CheckTime  time.Time
	ExpireTime time.Time
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
	// 未设置环境变量
	if Home == "" {
		dir, _ := os.Getwd()
		Home = filepath.Join(dir, "dev_home")
		_ = os.Setenv("DPI_HOME", Home)
	}
	Reload()
}

func Reload() {
	// 根据环境变量加载不同的配置文件
	RunDir = filepath.Join(Home, "run")
	LogDir = filepath.Join(Home, "log")
	EtcDir = filepath.Join(Home, "etc")
	BinDir = filepath.Join(Home, "bin")
	UploadDir = filepath.Join(Home, "uploads")

	ensureDirExists(RunDir)
	ensureDirExists(LogDir)
	ensureDirExists(EtcDir)
	ensureDirExists(BinDir)
	ensureDirExists(UploadDir)

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
	// 机器码生成
	if len(Cfg.MachineID) == 0 {
		if Cfg.MachineID, err = machineid.ID(); err != nil {
			panic(err)
		}
		viper.Set("machine_id", Cfg.MachineID)
		err = viper.WriteConfigAs(fmt.Sprintf("%s/dpi.yaml", EtcDir))
		if err != nil {
			panic(err)
		}
	}
}

func ensureDirExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			panic(fmt.Sprintf("Failed to create directory %s: %v", dir, err))
		}
	}
}
