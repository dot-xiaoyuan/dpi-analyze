package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ProxyRecord struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	IP          string             `json:"ip" bson:"ip"`
	Username    string             `json:"username" bson:"username"`
	Devices     []DeviceRecord     `json:"devices" bson:"devices"`
	AllCount    int                `json:"all_count" bson:"all_count"`
	MobileCount int                `json:"mobile_count" bson:"mobile_count"`
	PcCount     int                `json:"pc_count" bson:"pc_count"`
	LastSeen    time.Time          `json:"last_seen" bson:"last_seen"`
}
