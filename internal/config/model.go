package config

const (
	EnvDevelopment     = "dev"
	EnvProduction      = "prod"
	DevConfigFileName  = "config/config.yaml"
	ProdConfigFileName = "/etc/config.yaml"
)

type Capture struct {
	OfflineFile string `mapstructure:"offline_file"`
	NIC         string `mapstructure:"nic"`
	SnapLen     int32  `mapstructure:"snap_len"`
}

type Mongodb struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
	Use  bool   `mapstructure:"use"`
}
