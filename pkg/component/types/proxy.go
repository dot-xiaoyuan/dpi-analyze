package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ProxyRecord struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	IP          string             `json:"ip" bson:"ip"`
	Username    string             `json:"username" bson:"username"`
	Devices     []DeviceRecord     `json:"devices" bson:"devices"`
	AllCount    int                `json:"all_count" bson:"all_count"`
	MobileCount int                `json:"mobile_count" bson:"mobile_count"`
	PcCount     int                `json:"pc_count" bson:"pc_count"`
	LastSeen    time.Time          `json:"last_seen" bson:"last_seen"`
}

type SuspectedRecord struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	IP             string             `json:"ip" bson:"ip"`
	Username       string             `json:"username" bson:"username"`
	ReasonCategory string             `json:"reason_category" bson:"reason_category"`
	ReasonDetail   ReasonDetail       `json:"reason_detail" bson:"reason_detail"`
	Tags           []string           `json:"tags" bson:"tags"`
	Context        Context            `json:"context" bson:"context"`
	Remark         string             `json:"remark" bson:"remark"`
	LastSeen       time.Time          `json:"last_seen" bson:"last_seen"`
}

type ReasonDetail struct {
	Name        any    `json:"name" bson:"name"`
	Value       any    `json:"value" bson:"value"`
	Threshold   any    `json:"threshold" bson:"threshold"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	ExtraInfo   string `json:"extra_info,omitempty" bson:"extra_info,omitempty"`
}

type Context struct {
	Device string `json:"device" bson:"device"`
}
