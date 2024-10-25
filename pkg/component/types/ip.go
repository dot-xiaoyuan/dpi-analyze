package types

import "time"

type Property string

const (
	TTL       Property = "ttl"
	Mac       Property = "mac"
	UserAgent Property = "user_agent"
)

type Feature string

const (
	SNI         Feature = "sni"
	HTTP        Feature = "http"
	TLSVersion  Feature = "tls_version"
	CipherSuite Feature = "cipher_suite"
	Session     Feature = "session"
)

type TrafficRecord struct {
	IP          string         `bson:"ip" json:"ip"`
	WindowStart time.Time      `bson:"window_start" json:"window_start"`
	WindowEnd   time.Time      `bson:"window_end" json:"window_end"`
	FeatureType string         `bson:"feature_type" json:"feature_type"`
	FeatureData map[string]int `bson:"feature_data" json:"feature_data"`
	CreateAt    time.Time      `bson:"create_at" json:"create_at"`
}
