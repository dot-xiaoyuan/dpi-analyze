package types

import "time"

type DeviceRecord struct {
	IP           string    `json:"ip" bson:"ip"`
	OriginChanel Property  `json:"-" bson:"origin_chanel"`
	OriginValue  string    `json:"-" bson:"origin_value"`
	Os           string    `json:"os" bson:"os,omitempty"`
	Version      string    `json:"version" bson:"version,omitempty"`
	Device       string    `json:"device" bson:"device,omitempty"`
	Brand        string    `json:"brand" bson:"brand,omitempty"`
	Model        string    `json:"model" bson:"model,omitempty"`
	Icon         string    `json:"icon" bson:"icon,omitempty"`
	LastSeen     time.Time `json:"-" bson:"last_seen,omitempty"`
}

type DeviceRecordByFront struct {
	IP           string    `json:"ip" bson:"ip"`
	OriginChanel Property  `json:"origin_chanel" bson:"origin_chanel"`
	OriginValue  string    `json:"origin_value" bson:"origin_value"`
	Os           string    `json:"os" bson:"os,omitempty"`
	Version      string    `json:"version" bson:"version,omitempty"`
	Device       string    `json:"device" bson:"device,omitempty"`
	Brand        string    `json:"brand" bson:"brand,omitempty"`
	Model        string    `json:"model" bson:"model,omitempty"`
	Icon         string    `json:"icon" bson:"icon,omitempty"`
	LastSeen     time.Time `json:"last_seen" bson:"last_seen,omitempty"`
}
