package device

import "time"

type Device struct {
	IP       string    `json:"ip" bson:"ip"`
	Origin   string    `json:"origin" bson:"origin"`
	LastSeen time.Time `json:"last_seen" bson:"last_seen"`
}
