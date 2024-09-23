package capture

import (
	"context"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
	"go.uber.org/zap"
	"os"
	"time"
)

const ZetApplicationMap = "z_set:application_map"

// Ethernet 以太网
type Ethernet struct {
	SrcMac string `json:"src_mac"`
	DstMac string `json:"dst_mac"`
}

// Internet 网络层
type Internet struct {
	DstIP string `json:"dst_ip"`
	TTL   uint8  `json:"ttl"`
}

// Transmission 传输层
type Transmission struct {
	UpStream   int64 `json:"up_stream"`
	DownStream int64 `json:"down_stream"`
}

// TCPFlags 结构体，用于保存 TCP 协议的标志
type TCPFlags struct {
	SYN bool `bson:"syn"`
	ACK bool `bson:"ack"`
	FIN bool `bson:"fin"`
	RST bool `bson:"rst"`
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
	Host        string   `bson:"host"`
	UserAgent   string   `bson:"user_agent"`
	Urls        []string `bson:"urls"`
	ContentType string   `bson:"content_type"`
	Upgrade     string   `bson:"upgrade"`
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

type ApplicationInfo struct {
	AppName     string `bson:"app_name"`
	AppCategory string `bson:"app_category"`
}

// Metadata 存储所有协议相关的附加信息
type Metadata struct {
	HttpInfo        HttpInfo        `bson:"http_info"`
	DnsInfo         DnsInfo         `bson:"dns_info"`
	RtpInfo         RtpInfo         `bson:"rtp_info"`
	TlsInfo         TlsInfo         `bson:"tls_info"`
	ApplicationInfo ApplicationInfo `bson:"application_info"`
}

// CustomFields 存储用户自定义字段
type CustomFields struct {
	FieldName  string `bson:"field_name"`
	FieldValue string `bson:"field_value"`
}

// Sessions 用于存储每个网络会话的信息，包括源 IP、目标 IP、协议、传输层协议等。
type Sessions struct {
	SessionId           string                 `bson:"session_id"`
	SrcIp               string                 `bson:"src_ip"`
	DstIp               string                 `bson:"dst_ip"`
	SrcPort             string                 `bson:"src_port"`
	DstPort             string                 `bson:"dst_port"`
	Protocol            string                 `bson:"protocol"`
	StartTime           time.Time              `bson:"start_time"`
	EndTime             time.Time              `bson:"end_time"`
	PacketCount         int8                   `bson:"packet_count"`
	ByteCount           int16                  `bson:"byte_count"`
	ProtocolFlags       ProtocolFlags          `bson:"protocol_flags"` // 协议标志
	ApplicationProtocol protocols.ProtocolType `bson:"application_protocol"`
	Metadata            Metadata               `bson:"metadata"`
	CustomFields        CustomFields           `bson:"custom_fields"`
}

type LayerMap interface {
	Update(i interface{})
}

type Application interface {
	AddUp()
}

// AddUp 累加应用数
func (a *ApplicationInfo) AddUp() {
	if a.AppName == "" {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			zap.L().Error("application info", zap.Any("err", r))
			os.Exit(1)
		}
	}()
	rdb := redis.GetRedisClient()
	err := rdb.ZIncrBy(context.Background(), ZetApplicationMap, 1, a.AppName).Err()
	// 累加全局应用计数
	ApplicationCount++
	if err != nil {
		panic(err)
	}
}
