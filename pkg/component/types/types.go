package types

import (
	"context"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"os"
	"time"
)

var (
	ApplicationCount int // 应用总数
)

const (
	ZSetApplication    = "z_set:application"
	ZSetIP             = "z_set:ip"
	ZSetObserverTTL    = "z_set:observer:ttl"
	ZSetObserverMac    = "z_set:observer:mac"
	ZSetObserverUa     = "z_set:observer:ua"
	ZSetObserverDevice = "z_set:observer:device"
	ZSetOnlineUsers    = "z_set:online:users"
	HashAnalyzeIP      = "hash:analyze:ip:%s"
	SetIPDevices       = "set:ip:devices:%s"
	KeyDiscoverIP      = "key:discover:ip:%s"
	KeyDevicesAllIP    = "key:devices:all:ip:%s"
	KeyDevicesMobileIP = "key:devices:mobile:ip:%s"
	KeyDevicesPcIP     = "key:devices:pc:ip:%s"
)

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
	SYN bool `bson:"syn" json:"syn"`
	ACK bool `bson:"ack" json:"ack"`
	FIN bool `bson:"fin" json:"fin"`
	RST bool `bson:"rst" json:"rst"`
}

// UDPFlags 结构体，用于保存 UDP 协议的标志
type UDPFlags struct {
	IsDNS bool `bson:"is_dns" json:"is_dns"`
}

// ProtocolFlags 结构体，用于保存不同协议的标志
type ProtocolFlags struct {
	TCP TCPFlags `bson:"tcp" json:"tcp"`
	UDP UDPFlags `bson:"udp" json:"udp"`
}

// HttpInfo 存储 HTTP 相关信息
type HttpInfo struct {
	Host        string   `bson:"host,omitempty" json:"host"`
	UserAgent   string   `bson:"user_agent,omitempty" json:"user_agent"`
	Urls        []string `bson:"urls,omitempty" json:"urls"`
	ContentType string   `bson:"content_type,omitempty" json:"content_type"`
	Upgrade     string   `bson:"upgrade,omitempty" json:"upgrade"`
}

// DnsInfo 存储 DNS 相关信息
type DnsInfo struct {
	QueryName  string `bson:"query_name,omitempty" json:"query_name"`
	ResponseIp string `bson:"response_ip,omitempty" json:"response_ip"`
}

// RtpInfo 存储 RTP 相关信息
type RtpInfo struct {
	Codec   string `bson:"codec,omitempty" json:"codec"`
	Bitrate string `bson:"bitrate,omitempty" json:"bitrate"`
}

// TlsInfo 存储 TLS 相关信息
type TlsInfo struct {
	Version     string `bson:"version,omitempty" json:"version"`
	CipherSuite string `bson:"cipher_suite,omitempty" json:"cipher_suite"`
	Sni         string `bson:"sni,omitempty" json:"sni"`
}

type ApplicationInfo struct {
	AppName     string `bson:"app_name,omitempty" json:"app_name"`
	AppCategory string `bson:"app_category,omitempty" json:"app_category"`
}

// Metadata 存储所有协议相关的附加信息
type Metadata struct {
	HttpInfo        HttpInfo        `bson:"http_info,omitempty" json:"http_info"`
	DnsInfo         DnsInfo         `bson:"dns_info,omitempty" json:"dns_info"`
	RtpInfo         RtpInfo         `bson:"rtp_info,omitempty" json:"rtp_info"`
	TlsInfo         TlsInfo         `bson:"tls_info,omitempty" json:"tls_info"`
	ApplicationInfo ApplicationInfo `bson:"application_info,omitempty" json:"application_info"`
}

// CustomFields 存储用户自定义字段
type CustomFields struct {
	FieldName  string `bson:"field_name,omitempty" json:"field_name"`
	FieldValue string `bson:"field_value,omitempty" json:"field_value"`
}

// Sessions 用于存储每个网络会话的信息，包括源 IP、目标 IP、协议、传输层协议等。
type Sessions struct {
	ID                  primitive.ObjectID     `bson:"_id,omitempty" json:"_id"`
	Ident               string                 `bson:"ident" json:"ident"`
	SessionId           string                 `bson:"session_id" json:"session_id"`
	SrcIp               string                 `bson:"src_ip" json:"src_ip"`
	DstIp               string                 `bson:"dst_ip" json:"dst_ip"`
	SrcPort             string                 `bson:"src_port" json:"src_port"`
	DstPort             string                 `bson:"dst_port" json:"dst_port"`
	Protocol            string                 `bson:"protocol" json:"protocol"`
	PacketCount         int                    `bson:"packet_count" json:"packet_count"`
	ByteCount           int                    `bson:"byte_count" json:"byte_count"`
	MissBytes           int                    `bson:"miss_bytes" json:"miss_bytes"`
	OutOfOrderPackets   int                    `bson:"out_of_order_packets" json:"out_of_order_packets"`
	OutOfOrderBytes     int                    `bson:"out_of_order_bytes" json:"out_of_order_bytes"`
	OverlapBytes        int                    `bson:"overlap_bytes" json:"overlap_bytes"`
	OverlapPackets      int                    `bson:"overlap_packets" json:"overlap_packets"`
	StartTime           time.Time              `bson:"start_time" json:"start_time"`
	EndTime             time.Time              `bson:"end_time" json:"end_time"`
	ProtocolFlags       ProtocolFlags          `bson:"protocol_flags" json:"protocol_flags"` // 协议标志
	ApplicationProtocol protocols.ProtocolType `bson:"application_protocol" json:"application_protocol"`
	Metadata            Metadata               `bson:"metadata" json:"metadata"`
	CustomFields        CustomFields           `bson:"custom_fields" json:"custom_fields"`
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
	err := rdb.ZIncrBy(context.Background(), ZSetApplication, 1, a.AppName).Err()
	// 累加全局应用计数
	ApplicationCount++
	if err != nil {
		panic(err)
	}
}

type Charts struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

func GenerateChartData() []Charts {
	rdb := redis.GetRedisClient()
	result := rdb.ZRevRangeWithScores(context.Background(), ZSetApplication, 0, 50).Val()
	zap.L().Info("result", zap.Any("result", result))
	var charts []Charts
	for _, v := range result {
		charts = append(charts, Charts{
			Name:  v.Member.(string),
			Value: v.Score,
		})
	}
	return charts
}
