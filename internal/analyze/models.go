package analyze

// Packet 用于记录特定会话中的数据包信息。
type Packet struct {
	Id          string `bson:"_id"`
	SessionId   string `bson:"session_id"`
	Timestamp   string `bson:"timestamp"`
	SrcIp       string `bson:"src_ip"`
	DstIp       string `bson:"dst_ip"`
	SrcPort     string `bson:"src_port"`
	DstPort     string `bson:"dst_port"`
	Protocol    string `bson:"protocol"`
	PayloadSize string `bson:"payload_size"`
	Flags       string `bson:"flags"`
}

// FlowStatistics 用来统计特定应用流量的分析结果，例如各个应用层协议的流量统计等。
type FlowStatistics struct {
	Id           string `bson:"_id"`
	Protocol     string `bson:"protocol"`
	SessionCount string `bson:"session_count"`
	TotalBytes   string `bson:"total_bytes"`
	TotalPackets string `bson:"total_packets"`
	Timestamp    string `bson:"timestamp"`
}

//func (a *ApplicationInfo) AddUp() {
//	rdb := redis.GetRedisClient()
//	score := rdb.ZScore(context.Background(), ZetApplicationMap, member).Val()
//	if score >= 0 {
//		score++
//	} else {
//		score = 1
//	}
//	err := rdb.ZAdd(context.Background(), ZetApplicationMap, redis2.Z{
//		Score:  score,
//		Member: member,
//	}).Err()
//	if err != nil {
//		panic(err)
//	}
//}
