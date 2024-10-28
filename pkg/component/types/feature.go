package types

import "time"

const (
	Features           = "features"
	OnlineUsersFeature = "online_users"
)

type Feature struct {
	IP    string
	Field FeatureType
	Value string
}

type FeatureSet struct {
	IP       string                        `bson:"ip" json:"ip"`
	LastSeen time.Time                     `bson:"last_seen"`
	Features map[FeatureType][]FeatureData `bson:"features"`
	Total    []Chart                       `bson:"total"`
}

type FeatureData struct {
	LastSeen time.Time `bson:"last_seen"` // 最后一次访问时间
	Value    string    `bson:"value"`     // 特征数值
	Count    int       `bson:"count"`     // 时间窗口内相同数值计数
}

type Chart struct {
	Date       time.Time   `bson:"date" json:"date"`
	Industry   FeatureType `bson:"industry" json:"industry"`
	Unemployed int         `bson:"unemployed" json:"unemployed"`
}
