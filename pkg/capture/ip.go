package capture

import "time"

type TTLHistory struct {
	TTL       uint8     `bson:"ttl"`
	Timestamp time.Time `bson:"timestamp"`
}

type UAHistory struct {
	UserAgent string    `bson:"user_agent"`
	Timestamp time.Time `bson:"timestamp"`
}

type MacHistory struct {
	MacAddress string    `bson:"mac_address"`
	Timestamp  time.Time `bson:"timestamp"`
}

type IPActivityLogs struct {
	IP               string       `bson:"ip"`
	CurrentTTL       uint8        `bson:"current_ttl"`
	TTLHistory       []TTLHistory `bson:"ttl_history"`
	CurrentUserAgent string       `bson:"current_user_agent"`
	UAHistory        []UAHistory  `bson:"ua_history"`
	CurrentMac       string       `bson:"current_mac"`
	MacHistory       []MacHistory `bson:"mac_history"`
	LastSeen         time.Time    `bson:"last_seen"`
}
