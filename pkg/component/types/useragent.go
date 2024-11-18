package types

import (
	"time"
)

type UserAgentRecord struct {
	IP        string    `json:"ip" bson:"ip"`
	Host      string    `json:"host" bson:"host"`
	UserAgent string    `json:"user_agent" bson:"user_agent"`
	Ua        string    `json:"ua" bson:"ua"`
	UaVersion string    `json:"ua_version" bson:"ua_version"`
	Os        string    `json:"os" bson:"os"`
	OsVersion string    `json:"os_version" bson:"os_version"`
	Device    string    `json:"device" bson:"device"`
	Brand     string    `json:"brand" bson:"brand"`
	Model     string    `json:"model" bson:"model"`
	LastSeen  time.Time `json:"last_seen" bson:"last_seen"`
}
