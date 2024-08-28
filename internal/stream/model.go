package stream

import "time"

type SessionCollection struct {
	SessionID      string `bson:"session_id" comment:"会话ID"`
	Identification string `bson:"identification" comment:"标识ID"`
	SrcIP          string `bson:"src_ip" comment:"源IP"`
	DstIP          string `bson:"dst_ip" comment:"目标IP"`
	SrcPort        string `bson:"src_port" comment:"源端口"`
	DstPort        string `bson:"dst_port" comment:"目标端口"`
	Protocol       string `bson:"protocol" comment:"协议"`
	Application    string `bson:"application" comment:"应用"`
	Flags          string `bson:"flags" comment:"标志"`
}

type PacketCollection struct {
	SessionID    string    `bson:"session_id"`
	PacketNumber int       `bson:"packet_number"`
	Timestamp    time.Time `bson:"timestamp"`
	Direction    string    `bson:"direction"`
}
