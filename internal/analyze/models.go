package analyze

// TCPFlags 结构体，用于保存 TCP 协议的标志
type TCPFlags struct {
	SYN bool `bson:"syn"`
	ACK bool `bson:"ack"`
	FIN bool `bson:"fin"`
}

// UDPFlags 结构体，用于保存 UDP 协议的标志
type UDPFlags struct {
	IsDNS bool `bson:"is_dns"`
}

// ProtocolFlags 结构体，用于保存不同协议的标志
type ProtocolFlags struct {
	TCP TCPFlags `bson:"tcp"`
	UDP UDPFlags `bson:"udp"`
}

// HttpInfo 存储 HTTP 相关信息
type HttpInfo struct {
	Host      string `bson:"host"`
	UserAgent string `bson:"user_agent"`
}

// DnsInfo 存储 DNS 相关信息
type DnsInfo struct {
	QueryName  string `bson:"query_name"`
	ResponseIp string `bson:"response_ip"`
}

// RtpInfo 存储 RTP 相关信息
type RtpInfo struct {
	Codec   string `bson:"codec"`
	Bitrate string `bson:"bitrate"`
}

// TlsInfo 存储 TLS 相关信息
type TlsInfo struct {
	Version     string `bson:"version"`
	CipherSuite string `bson:"cipher_suite"`
	Sni         string `bson:"sni"`
}

// Metadata 存储所有协议相关的附加信息
type Metadata struct {
	HttpInfo HttpInfo `bson:"http_info"`
	DnsInfo  DnsInfo  `bson:"dns_info"`
	RtpInfo  RtpInfo  `bson:"rtp_info"`
	TlsInfo  TlsInfo  `bson:"tls_info"`
}

// CustomFields 存储用户自定义字段
type CustomFields struct {
	FieldName  string `bson:"field_name"`
	FieldValue string `bson:"field_value"`
}

// Sessions 用于存储每个网络会话的信息，包括源 IP、目标 IP、协议、传输层协议等。
type Sessions struct {
	Id                  string        `bson:"_id"`
	SessionId           string        `bson:"session_id"`
	SrcIp               string        `bson:"src_ip"`
	DstIp               string        `bson:"dst_ip"`
	SrcPort             string        `bson:"src_port"`
	DstPort             string        `bson:"dst_port"`
	Protocol            string        `bson:"protocol"`
	StartTime           string        `bson:"start_time"`
	EndTime             string        `bson:"end_time"`
	PacketCount         int8          `bson:"packet_count"`
	ByteCount           int16         `bson:"byte_count"`
	ProtocolFlags       ProtocolFlags `bson:"protocol_flags"` // 协议标志
	ApplicationProtocol string        `bson:"application_protocol"`
	Metadata            Metadata      `bson:"metadata"`
	CustomFields        CustomFields  `bson:"custom_fields"`
}

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
