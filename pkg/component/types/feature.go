package types

import "time"

type FeatureSet struct {
	LastSeen time.Time                 `bson:"last_seen"`
	Features map[Feature][]FeatureData `bson:"features"`
	Total    []Chart                   `bson:"total"`
}

type FeatureData struct {
	LastSeen time.Time `bson:"last_seen"` // 最后一次访问时间
	Value    string    `bson:"value"`     // 特征数值
	Count    int       `bson:"count"`     // 时间窗口内相同数值计数
}

type Chart struct {
	Date       time.Time `bson:"date" json:"date"`
	Industry   Feature   `bson:"industry" json:"industry"`
	Unemployed int       `bson:"unemployed" json:"unemployed"`
}
