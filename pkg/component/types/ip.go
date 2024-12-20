package types

import "time"

type Property string

const (
	TTL        Property = "ttl"
	Mac        Property = "mac"
	UserAgent  Property = "user_agent"
	Device     Property = "device"
	DeviceName Property = "device_name"
	DeviceType Property = "device_type"
)

type FeatureType string

const (
	SNI         FeatureType = "sni"
	HTTP        FeatureType = "http"
	TLSVersion  FeatureType = "tls_version"
	CipherSuite FeatureType = "cipher_suite"
	Session     FeatureType = "session"
	DNS         FeatureType = "dns"
	DHCP        FeatureType = "dhcp"
	DHCPv6      FeatureType = "dhcp_v6"
	NTP         FeatureType = "ntp"
	QUIC        FeatureType = "quic"
	TFTP        FeatureType = "tftp"
	SNMP        FeatureType = "snmp"
	MDNS        FeatureType = "mdns"
	VXLAN       FeatureType = "vxlan"
	SIP         FeatureType = "sip"
	SFlow       FeatureType = "s_flow"
	Geneve      FeatureType = "geneve"
	BFD         FeatureType = "bfd"
	GTPv1U      FeatureType = "gtp_v1u"
	RMCP        FeatureType = "rmcp"
	Radius      FeatureType = "radius"
)

type TrafficRecord struct {
	IP          string         `bson:"ip" json:"ip"`
	WindowStart time.Time      `bson:"window_start" json:"window_start"`
	WindowEnd   time.Time      `bson:"window_end" json:"window_end"`
	FeatureType string         `bson:"feature_type" json:"feature_type"`
	FeatureData map[string]int `bson:"feature_data" json:"feature_data"`
	CreateAt    time.Time      `bson:"create_at" json:"create_at"`
}
